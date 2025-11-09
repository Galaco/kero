package physics

import (
	"fmt"
	"strings"

	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/kero/framework/physics/collision"
	"github.com/galaco/kero/framework/physics/collision/bullet"
	"github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/messages"
	"github.com/go-gl/mathgl/mgl32"
)

type PhysicsSystem struct {
	eventBus     *event.Dispatcher
	sceneManager *scene.Manager
	dataScene    *scene.StaticScene

	// Phase 4: Pure ECS
	ecsWorld *ecs.World

	// Bullet
	sdk   bullet.BulletPhysicsSDKHandle
	world bullet.BulletDynamicWorldHandle

	bspRigidBody               *bspCollisionMesh
	displacementRigidBody      *displacementCollisionMesh
	studiomodelCollisionMeshes map[string]studiomodelCollisionMesh

	// Player physics (Phase 4: ECS entity only, no legacy player)
	playerEntity       interface{} // ECS player entity (ecs.Entity stored as interface{})
	playerCapsuleShape bullet.BulletCollisionShapeHandle
}

func (system *PhysicsSystem) Initialize() {
	// Register typed event listeners (Phase 3)
	event.RegisterTypedEvent(system.eventBus, system.onChangeLevelTyped)
	event.RegisterTypedEvent(system.eventBus, system.onLoadingLevelParsedTyped)
	event.RegisterTypedEvent(system.eventBus, func(e messages.EngineDisconnectEvent) {
		system.Cleanup()
	})

	// Physics debug and monitoring console variables
	console.AddConvarBool("r_drawcollisionmodels", "Render collision mode vertices", false)
	console.AddConvarBool("physics_debug", "Show physics timing debug info", false)
}

// FixedUpdate runs the physics simulation at a fixed timestep.
// This method should be called with a constant dt value (typically 1.0/60.0 = 0.0166667 seconds).
// Fixed timestep ensures stable, deterministic physics simulation regardless of frame rate.
func (system *PhysicsSystem) FixedUpdate(dt float64) {
	// Safety check: Don't run physics if world not initialized
	// (Bullet world is created in onLoadingLevelParsedTyped)
	if system.dataScene == nil {
		return
	}

	//if !input.Keyboard().IsKeyPressed(input.KeyQ) {
	//	return
	//}

	// Debug logging
	if console.GetConvarBoolean("physics_debug") {
		console.PrintString(console.LevelInfo, fmt.Sprintf("Physics dt: %.6f", dt))
	}

	// Query entities with Transform + Physics components
	query := system.ecsWorld.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypePhysics).
		Build()

	entities := query.Entities()
	if len(entities) == 0 {
		return
	}

	// Phase 1: Write ECS transforms to Bullet physics engine
	for _, entity := range entities {
		transform, _ := ecs.GetComponent[components.Transform](system.ecsWorld, entity)
		physics, _ := ecs.GetComponent[components.Physics](system.ecsWorld, entity)

		// Check if rigid body is initialized
		if !physics.HasRigidBody() {
			continue
		}

		// Get RigidBody from component (cast from interface{} to interface)
		rigidBody := physics.GetRigidBodyHandle().(collision.RigidBody)

		// Convert transform to matrix and update Bullet
		transformMatrix := mgl32.Translate3D(transform.Position.X(), transform.Position.Y(), transform.Position.Z()).
			Mul4(transform.Orientation.Mat4())

		// Update Bullet rigid body transform
		rigidBody.SetTransform(transformMatrix)
	}

	// Step physics simulation with fixed timestep
	bullet.BulletStepSimulation(system.world, dt)

	// Phase 2: Read physics results from Bullet back to ECS components
	for _, entity := range entities {
		transform, _ := ecs.GetComponent[components.Transform](system.ecsWorld, entity)
		physics, _ := ecs.GetComponent[components.Physics](system.ecsWorld, entity)

		// Check if rigid body is initialized
		if !physics.HasRigidBody() {
			continue
		}

		// Get RigidBody from component
		rigidBody := physics.GetRigidBodyHandle().(collision.RigidBody)

		// Read physics results from Bullet
		transform.Position = rigidBody.GetTranslation()
		transform.Orientation = rigidBody.GetOrientation()
	}

	// Phase 4: No sync needed - renderer reads from ECS components directly
}

// Update is a wrapper around FixedUpdate for backward compatibility.
// Deprecated: Use FixedUpdate for physics simulation with fixed timestep.
// This method delegates to FixedUpdate but should not be used directly.
func (system *PhysicsSystem) Update(dt float64) {
	system.FixedUpdate(dt)
}

// PrepareDebug populates the debug buffer with collision mesh visualization
// This method should be called before rendering to allow the renderer to display physics debug info
func (system *PhysicsSystem) PrepareDebug(buffer interface{}) {
	// Type assert to the debug buffer interface
	type debugBuffer interface {
		AddLines(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4)
	}

	debugBuf, ok := buffer.(debugBuffer)
	if !ok || debugBuf == nil {
		return
	}

	// Convert BSP collision mesh vertices to Vec3 format
	if system.bspRigidBody != nil && len(system.bspRigidBody.vertices) > 0 {
		verts := make([]mgl32.Vec3, 0, len(system.bspRigidBody.vertices))
		for _, vert := range system.bspRigidBody.vertices {
			verts = append(verts, mgl32.Vec3{vert[0], vert[1], vert[2]})
		}
		debugBuf.AddLines(verts, mgl32.Vec3{1, 0, 1}, mgl32.Ident4())
	}

	// Draw collision meshes for physics entities
	query := system.ecsWorld.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypePhysics).
		With(ecs.ComponentTypeModel).
		Build()

	for _, entity := range query.Entities() {
		transform, _ := ecs.GetComponent[components.Transform](system.ecsWorld, entity)
		physics, _ := ecs.GetComponent[components.Physics](system.ecsWorld, entity)
		model, _ := ecs.GetComponent[components.Model](system.ecsWorld, entity)

		// Skip if no rigid body or model
		if !physics.HasRigidBody() {
			continue
		}

		// Create transformation matrix
		transformMatrix := mgl32.Translate3D(transform.Position.X(), transform.Position.Y(), transform.Position.Z()).
			Mul4(transform.Orientation.Mat4())

		// Add collision mesh for each part of the studiomodel using model ID from component
		if collisionMesh, exists := system.studiomodelCollisionMeshes[model.MeshPath]; exists {
			for _, r := range collisionMesh.vertices {
				verts := make([]mgl32.Vec3, 0, len(r))
				for _, v := range r {
					verts = append(verts, mgl32.Vec3{v[0], v[1], v[2]})
				}
				debugBuf.AddLines(verts, mgl32.Vec3{1, 0, 1}, transformMatrix)
			}
		}
	}

	// Phase 4: Draw player collision capsule (ECS player)
	if console.GetConvarBoolean("r_drawplayercollision") && system.playerEntity != nil {
		system.drawPlayerCapsule(debugBuf)
	}

	// Phase 4: Draw player collision hits (ECS player)
	if console.GetConvarBoolean("r_drawplayerhits") && system.playerEntity != nil {
		system.drawPlayerCollisions(debugBuf)
	}
}

// drawPlayerCapsule renders the player's collision capsule wireframe
// Phase 4: Uses ECS components instead of legacy player
func (system *PhysicsSystem) drawPlayerCapsule(debugBuf interface{}) {
	type debugBuffer interface {
		AddLines(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4)
	}

	buf := debugBuf.(debugBuffer)

	// Get player entity's components
	playerEntity := system.playerEntity.(ecs.Entity)

	// Get Transform for position
	transform, ok := ecs.GetComponent[components.Transform](system.ecsWorld, playerEntity)
	if !ok {
		return
	}

	// Get CharacterController for debug geometry
	charController, ok := ecs.GetComponent[components.CharacterController](system.ecsWorld, playerEntity)
	if !ok || !charController.HasController() {
		return
	}

	controller := charController.GetController().(*collision.CharacterController)

	// Generate capsule geometry
	vertices := controller.GetCapsuleDebugGeometry(transform.Position)

	// Draw in green
	buf.AddLines(vertices, mgl32.Vec3{0, 1, 0}, mgl32.Ident4())
}

// drawPlayerCollisions renders collision hit points and normals
// Phase 4: Uses ECS components instead of legacy player
func (system *PhysicsSystem) drawPlayerCollisions(debugBuf interface{}) {
	type debugBuffer interface {
		AddLines(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4)
	}

	buf := debugBuf.(debugBuffer)

	// Get player entity's CharacterController component
	playerEntity := system.playerEntity.(ecs.Entity)
	charController, ok := ecs.GetComponent[components.CharacterController](system.ecsWorld, playerEntity)
	if !ok || !charController.HasController() {
		return
	}

	controller := charController.GetController().(*collision.CharacterController)
	hits := controller.GetDebugHits()

	for _, hit := range hits {
		// Draw hit point as a small cross (red)
		size := float32(2.0)
		vertices := []mgl32.Vec3{
			// X axis
			hit.Point.Add(mgl32.Vec3{-size, 0, 0}),
			hit.Point.Add(mgl32.Vec3{size, 0, 0}),
			// Y axis
			hit.Point.Add(mgl32.Vec3{0, -size, 0}),
			hit.Point.Add(mgl32.Vec3{0, size, 0}),
			// Z axis
			hit.Point.Add(mgl32.Vec3{0, 0, -size}),
			hit.Point.Add(mgl32.Vec3{0, 0, size}),
		}
		buf.AddLines(vertices, mgl32.Vec3{1, 0, 0}, mgl32.Ident4()) // Red

		// Draw normal as a line from hit point (yellow)
		normalLength := float32(10.0)
		normalEnd := hit.Point.Add(hit.Normal.Mul(normalLength))
		normalVerts := []mgl32.Vec3{hit.Point, normalEnd}
		buf.AddLines(normalVerts, mgl32.Vec3{1, 1, 0}, mgl32.Ident4()) // Yellow
	}
}

func (system *PhysicsSystem) onChangeLevelTyped(e messages.ChangeLevelEvent) {
	if system.dataScene == nil {
		return
	}
	system.Cleanup()
}

func (system *PhysicsSystem) onLoadingLevelParsedTyped(e messages.LoadingLevelParsedEvent) {
	system.dataScene = e.Level

	// create an sdk handle
	system.sdk = bullet.BulletNewPhysicsSDK()
	// instance a world
	system.world = bullet.BulletNewDynamicWorld(system.sdk)
	bullet.BulletSetGravity(system.world, 0.0, 0.0, -100.0)

	console.PrintString(console.LevelInfo, "Generating collision structures....")

	// Generate BSP Rigidbody
	console.PrintString(console.LevelInfo, "BSP collision structure...")
	system.bspRigidBody = generateBspCollisionMesh(system.dataScene)
	bullet.BulletAddRigidBody(system.world, system.bspRigidBody.RigidBodyHandles)

	// Generate Displacement RigidBodies
	console.PrintString(console.LevelInfo, "Displacement collision structures...")
	system.displacementRigidBody = generateDisplacementCollisionMeshes(system.dataScene)
	if system.displacementRigidBody != nil {
		bullet.BulletAddRigidBody(system.world, system.displacementRigidBody.RigidBodyHandles)
	}

	// Generate Staticprop RigidBodies
	console.PrintString(console.LevelInfo, "Static prop collision structures...")
	for _, e := range system.dataScene.StaticProps {
		system.prepareModelInstanceRigidBody(e.Model(), e.Transform.TransformationMatrix(), true)
	}

	// Find entities that have a model
	console.PrintString(console.LevelInfo, "Physics prop collision structures...")

	entityCount := 0
	for _, e := range system.dataScene.Entities {
		if e.Model() != nil {
			entityCount++
			disableMotion := true
			// @TODO Once entity base types are implemented they can be detected better than this
			if strings.HasPrefix(e.Classname(), "prop_physics") {
				disableMotion = false
			}
			system.prepareModelInstanceRigidBody(e.Model(), e.Transform().TransformationMatrix(), disableMotion)

			// Phase 1: Create ECS entity directly with RigidBody handle
			system.createECSEntityFromLegacy(e)
		}
	}
	console.PrintString(console.LevelSuccess, "Collision structures ready!")

	// Phase 4: Initialize CharacterControllers for all entities (including player)
	system.initializeCharacterControllers()
}

// createECSEntityFromLegacy creates an ECS entity from a legacy entity with proper component setup.
// Phase 1-2: Directly stores RigidBody and ModelInstance handles in components.
func (system *PhysicsSystem) createECSEntityFromLegacy(legacyEntity interface{}) ecs.Entity {
	// Use the actual IEntity interface from framework/entity
	e, ok := legacyEntity.(entity.IEntity)
	if !ok {
		return ecs.NullEntity
	}

	// Create new ECS entity
	entity := system.ecsWorld.CreateEntity()

	// Add Transform component from legacy entity
	legacyTransform := e.Transform()
	transform := components.Transform{
		Position:    legacyTransform.Translation,
		Orientation: legacyTransform.Orientation,
		Scale:       legacyTransform.Scale,
	}
	ecs.AddComponent(system.ecsWorld, entity, transform)

	// Add Physics component with RigidBody handle
	if e.Model() != nil && e.Model().RigidBody != nil {
		mass := e.Model().Model.OriginalStudiomodel.Mdl.Header.Mass

		physics := components.Physics{
			Mass:            mass,
			Restitution:     0.5,
			Friction:        0.5,
			UseGravity:      mass > 0, // Only dynamic objects use gravity
			IsStatic:        mass == 0,
			RigidBodyHandle: e.Model().RigidBody, // ← Store actual handle!
		}
		ecs.AddComponent(system.ecsWorld, entity, physics)
	}

	// Add Model component (for renderer)
	if e.Model() != nil {
		model := components.Model{
			MeshPath:            e.Model().Model.Id,
			Visible:             true,
			ModelInstanceHandle: e.Model(), // Phase 2: Store ModelInstance handle
		}
		ecs.AddComponent(system.ecsWorld, entity, model)
	}

	// Phase 4: No bridge registration needed - all data in ECS components

	return entity
}

func (system *PhysicsSystem) prepareModelInstanceRigidBody(model *mesh.ModelInstance, initialTransformation mgl32.Mat4, isStatic bool) {
	mass := float32(0)
	if isStatic == false {
		mass = model.Model.OriginalStudiomodel.Mdl.Header.Mass
	}

	// Prepare Bullet environment for collision meshes
	if model.Model.OriginalStudiomodel.Phy != nil {
		// We have an actual source engine .phy collision model
		if _, ok := system.studiomodelCollisionMeshes[model.Model.Id]; !ok {
			system.studiomodelCollisionMeshes[model.Model.Id] = generateCollisionMeshFromStudiomodelPhy(model.Model.OriginalStudiomodel.Phy)
		}
		model.RigidBody = collision.NewConvexHullFromExistingShape(
			mass,
			system.studiomodelCollisionMeshes[model.Model.Id].compoundShapeHandle)
	} else {
		// Fall back to generating one
		model.RigidBody = collision.NewSphericalHull(4)
	}

	model.RigidBody.SetTransform(initialTransformation)

	// For dynamic objects (mass > 0), start in sleeping state to prevent
	// violent ejection if spawned slightly penetrating static geometry.
	// They will wake naturally when physics updates and settle via gravity.
	if !isStatic && mass > 0 {
		bullet.BulletForceActivationState(model.RigidBody.BulletHandle(), bullet.ActivationStateIslandSleeping)
	}

	bullet.BulletAddRigidBody(system.world, model.RigidBody.BulletHandle())
}

// RegisterPlayer stores the ECS player entity for physics initialization.
// Phase 4: Simplified - CharacterController is created via component query, not player-specific code.
// The ecsPlayer parameter should be an ecs.Entity with CharacterController component.
func (system *PhysicsSystem) RegisterPlayer(ecsPlayer ...interface{}) {
	// Phase 4: Store ECS player entity if provided
	if len(ecsPlayer) > 0 && ecsPlayer[0] != nil {
		system.playerEntity = ecsPlayer[0]
		console.PrintString(console.LevelInfo, "Player entity registered with physics system")
	}

	// If physics world already exists, initialize CharacterControllers immediately
	if system.dataScene != nil {
		system.initializeCharacterControllers()
	}
	// Otherwise, it will be initialized in onLoadingLevelParsedTyped
}

// initializeCharacterControllers creates character controllers for all entities with CharacterController components.
// Phase 4: This replaces the legacy player-specific physics initialization.
func (system *PhysicsSystem) initializeCharacterControllers() {
	if system.dataScene == nil {
		return
	}

	// Query for all entities with CharacterController component
	query := system.ecsWorld.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypeCharacterController).
		Build()

	entities := query.Entities()
	if len(entities) == 0 {
		return
	}

	console.PrintString(console.LevelInfo, fmt.Sprintf("Initializing CharacterControllers for %d entities", len(entities)))

	for _, entity := range entities {
		charController, ok := ecs.GetComponent[components.CharacterController](system.ecsWorld, entity)
		if !ok {
			continue
		}

		// Skip if already initialized
		if charController.HasController() {
			console.PrintString(console.LevelInfo, fmt.Sprintf("Entity %d already has CharacterController, skipping", entity))
			continue
		}

		// Create Bullet capsule shape
		capsuleHeight := float64(charController.Height) - (2 * float64(charController.Radius))

		console.PrintString(console.LevelInfo, fmt.Sprintf("Creating CharacterController for entity %d: radius=%.1f, capsuleHeight=%.1f, totalHeight=%.1f, stepHeight=%.1f",
			entity, charController.Radius, capsuleHeight, charController.Height, charController.StepHeight))

		capsuleShape := bullet.BulletNewCapsuleShapeZ(float64(charController.Radius), capsuleHeight)

		// Create CharacterController
		controller := collision.NewCharacterController(
			system.world,
			capsuleShape,
			float64(charController.Height),
			float64(charController.Radius),
			float64(charController.StepHeight))

		// Store in component
		charController.SetController(controller)

		console.PrintString(console.LevelSuccess, fmt.Sprintf("CharacterController initialized for entity %d", entity))
	}
}

func (system *PhysicsSystem) Cleanup() {
	if system.dataScene == nil {
		return
	}

	// Phase 4: Component-based cleanup - Delete rigid bodies from Physics components
	query := system.ecsWorld.Query().
		With(ecs.ComponentTypePhysics).
		Build()

	for _, entity := range query.Entities() {
		physics, ok := ecs.GetComponent[components.Physics](system.ecsWorld, entity)
		if !ok || !physics.HasRigidBody() {
			continue
		}

		rigidBody := physics.GetRigidBodyHandle().(collision.RigidBody)
		bullet.BulletDeleteRigidBody(rigidBody.BulletHandle())
	}

	// Cleanup CharacterControllers (including player)
	charQuery := system.ecsWorld.Query().
		With(ecs.ComponentTypeCharacterController).
		Build()

	for _, entity := range charQuery.Entities() {
		charController, ok := ecs.GetComponent[components.CharacterController](system.ecsWorld, entity)
		if !ok || !charController.HasController() {
			continue
		}

		// CharacterController cleanup (if needed in future)
		// Currently CharacterController doesn't allocate external resources
		// Bullet capsule shapes are cleaned up with the world
	}

	// Cleanup BSP and displacement collision meshes
	bullet.BulletDeleteRigidBody(system.bspRigidBody.RigidBodyHandles)

	if system.displacementRigidBody != nil {
		bullet.BulletDeleteRigidBody(system.displacementRigidBody.RigidBodyHandles)
	}

	// Cleanup Bullet world
	bullet.BulletDeleteDynamicWorld(system.world)
	bullet.BulletDeletePhysicsSDK(system.sdk)

	// Clear state
	system.playerCapsuleShape = bullet.BulletCollisionShapeHandle{}
	system.playerEntity = nil
	system.dataScene = nil
	system.bspRigidBody = nil
	system.displacementRigidBody = nil
}

// NewPhysicsSystem creates a new physics system with explicit dependencies
// Phase 4: No legacy bridge needed - pure ECS
func NewPhysicsSystem(eventBus *event.Dispatcher, sceneManager *scene.Manager, ecsWorld *ecs.World) *PhysicsSystem {
	return &PhysicsSystem{
		eventBus:                   eventBus,
		sceneManager:               sceneManager,
		ecsWorld:                   ecsWorld,
		studiomodelCollisionMeshes: map[string]studiomodelCollisionMesh{},
	}
}
