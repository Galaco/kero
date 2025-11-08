package physics

import (
	"fmt"
	"strings"

	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
	"github.com/galaco/kero/framework/ecs/legacy"
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

	// Phase 4: ECS integration
	ecsWorld     *ecs.World
	legacyBridge *legacy.Bridge

	// Bullet
	sdk   bullet.BulletPhysicsSDKHandle
	world bullet.BulletDynamicWorldHandle

	bspRigidBody               *bspCollisionMesh
	displacementRigidBody      *displacementCollisionMesh
	studiomodelCollisionMeshes map[string]studiomodelCollisionMesh

	// Player physics
	player             interface{} // Stored as interface{} to avoid import cycle
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

	// NOTE: Phase 1 Migration - No longer need to sync to legacy entities
	// Once Phase 2 (Model component) is complete, we can remove legacyBridge entirely
	if system.legacyBridge != nil {
		system.legacyBridge.SyncAllECSToLegacy()
	}
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

	// Draw player collision capsule
	if console.GetConvarBoolean("r_drawplayercollision") && system.player != nil {
		system.drawPlayerCapsule(debugBuf)
	}

	// Draw player collision hits
	if console.GetConvarBoolean("r_drawplayerhits") && system.player != nil {
		system.drawPlayerCollisions(debugBuf)
	}
}

// drawPlayerCapsule renders the player's collision capsule wireframe
func (system *PhysicsSystem) drawPlayerCapsule(debugBuf interface{}) {
	type debugBuffer interface {
		AddLines(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4)
	}

	buf := debugBuf.(debugBuffer)

	// Get player's CharacterController via type assertion
	type playerWithController interface {
		GetCharacterController() *collision.CharacterController
		GetPosition() mgl32.Vec3
	}

	if playerCtrl, ok := system.player.(playerWithController); ok {
		controller := playerCtrl.GetCharacterController()
		if controller != nil {
			// Get current position
			pos := playerCtrl.GetPosition()

			// Generate capsule geometry
			vertices := controller.GetCapsuleDebugGeometry(pos)

			// Draw in green
			buf.AddLines(vertices, mgl32.Vec3{0, 1, 0}, mgl32.Ident4())
		}
	}
}

// drawPlayerCollisions renders collision hit points and normals
func (system *PhysicsSystem) drawPlayerCollisions(debugBuf interface{}) {
	type debugBuffer interface {
		AddLines(vertices []mgl32.Vec3, color mgl32.Vec3, transform mgl32.Mat4)
	}

	buf := debugBuf.(debugBuffer)

	// Get player's CharacterController
	type playerWithController interface {
		GetCharacterController() *collision.CharacterController
	}

	if playerCtrl, ok := system.player.(playerWithController); ok {
		controller := playerCtrl.GetCharacterController()
		if controller != nil {
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

	// Initialize player physics if player was registered
	if system.player != nil {
		system.initializePlayerPhysics()
	}
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

	// Still register with legacy bridge for Phase 2 (entity creation still uses bridge)
	// This will be removed in Phase 4
	if system.legacyBridge != nil {
		system.legacyBridge.Register(entity, legacyEntity)
	}

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

// RegisterPlayer adds the player to the physics world and sets up collision.
// The player parameter should be a *gameEntity.Player (from game/entity package).
// We use interface{} to avoid import cycles.
func (system *PhysicsSystem) RegisterPlayer(player interface{}) {
	// Store the player reference
	system.player = player

	// If physics world already exists, initialize player physics immediately
	if system.dataScene != nil {
		system.initializePlayerPhysics()
	}
	// Otherwise, it will be initialized in onLoadingLevelParsedTyped
}

// initializePlayerPhysics creates the character controller for the player.
// This is called after the physics world is created.
func (system *PhysicsSystem) initializePlayerPhysics() {
	if system.player == nil || system.dataScene == nil {
		return
	}

	// Create capsule shape for player character controller
	// Player dimensions: radius=16, total height=72
	// Capsule height = total height - (2 * radius) = 72 - 32 = 40
	playerRadius := float64(16)
	playerHeight := float64(72)
	capsuleHeight := playerHeight - (2 * playerRadius)

	console.PrintString(console.LevelInfo, fmt.Sprintf("Creating player capsule: radius=%.1f, height=%.1f, totalHeight=%.1f",
		playerRadius, capsuleHeight, playerHeight))

	system.playerCapsuleShape = bullet.BulletNewCapsuleShapeZ(playerRadius, capsuleHeight)

	// Call InitializePhysics on the player via reflection (interface{} method call)
	// We expect the player to have a method: InitializePhysics(world interface{}, capsuleShape interface{})
	type physicsInitializer interface {
		InitializePhysics(world interface{}, capsuleShape interface{})
	}

	if playerWithPhysics, ok := system.player.(physicsInitializer); ok {
		playerWithPhysics.InitializePhysics(system.world, system.playerCapsuleShape)
		console.PrintString(console.LevelSuccess, "Player physics initialized!")
	} else {
		console.PrintString(console.LevelWarning, "Player does not implement InitializePhysics method")
	}
}

func (system *PhysicsSystem) Cleanup() {
	if system.dataScene == nil {
		return
	}

	// Phase 4: Clear ECS bridge
	if system.legacyBridge != nil {
		system.legacyBridge.Clear()
	}

	// Cleanup player physics
	// Note: Collision shapes are owned by Bullet and cleaned up when the world is deleted
	system.playerCapsuleShape = bullet.BulletCollisionShapeHandle{}
	system.player = nil

	bullet.BulletDeleteDynamicWorld(system.world)
	bullet.BulletDeletePhysicsSDK(system.sdk)

	// Delete rigid bodies for all physics entities
	query := system.ecsWorld.Query().
		With(ecs.ComponentTypePhysics).
		Build()

	for _, entity := range query.Entities() {
		legacyEntity, exists := system.legacyBridge.GetLegacyEntity(entity)
		if exists && legacyEntity.Model() != nil && legacyEntity.Model().RigidBody != nil {
			bullet.BulletDeleteRigidBody(legacyEntity.Model().RigidBody.BulletHandle())
			legacyEntity.Model().RigidBody = nil
		}
	}

	bullet.BulletDeleteRigidBody(system.bspRigidBody.RigidBodyHandles)

	if system.displacementRigidBody != nil {
		bullet.BulletDeleteRigidBody(system.displacementRigidBody.RigidBodyHandles)
	}
	system.dataScene = nil
	system.bspRigidBody = nil
	system.displacementRigidBody = nil
}

// NewPhysicsSystem creates a new physics system with explicit dependencies
func NewPhysicsSystem(eventBus *event.Dispatcher, sceneManager *scene.Manager, ecsWorld *ecs.World, legacyBridge *legacy.Bridge) *PhysicsSystem {
	return &PhysicsSystem{
		eventBus:                   eventBus,
		sceneManager:               sceneManager,
		ecsWorld:                   ecsWorld,
		legacyBridge:               legacyBridge,
		studiomodelCollisionMeshes: map[string]studiomodelCollisionMesh{},
	}
}
