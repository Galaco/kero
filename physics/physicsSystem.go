package physics

import (
	"fmt"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
	"github.com/galaco/kero/framework/ecs/legacy"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/physics/collision"
	"github.com/galaco/kero/framework/physics/collision/bullet"
	"github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/messages"
	"github.com/go-gl/mathgl/mgl32"
	"strings"
)

type PhysicsSystem struct {
	eventBus      *event.Dispatcher
	sceneManager  *scene.Manager
	dataScene     *scene.StaticScene

	// Phase 4: ECS integration
	ecsWorld      *ecs.World
	legacyBridge  *legacy.Bridge

	// Bullet
	sdk   bullet.BulletPhysicsSDKHandle
	world bullet.BulletDynamicWorldHandle

	bspRigidBody               *bspCollisionMesh
	displacementRigidBody      *displacementCollisionMesh
	studiomodelCollisionMeshes map[string]studiomodelCollisionMesh
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

	if !input.Keyboard().IsKeyPressed(input.KeyQ) {
		return
	}

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

	// Update entity transforms to physics engine
	for _, entity := range entities {
		transform, _ := ecs.GetComponent[components.Transform](system.ecsWorld, entity)

		// Get legacy entity to access RigidBody (migration phase)
		legacyEntity, exists := system.legacyBridge.GetLegacyEntity(entity)
		if !exists || legacyEntity.Model() == nil || legacyEntity.Model().RigidBody == nil {
			continue
		}

		// Convert transform to matrix and update Bullet
		transformMatrix := mgl32.Translate3D(transform.Position.X(), transform.Position.Y(), transform.Position.Z()).
			Mul4(transform.Orientation.Mat4())

		// Update Bullet rigid body transform
		legacyEntity.Model().RigidBody.SetTransform(transformMatrix)
	}

	// Step physics simulation with fixed timestep
	bullet.BulletStepSimulation(system.world, dt)

	// Apply physics results back to ECS components
	for _, entity := range entities {
		transform, _ := ecs.GetComponent[components.Transform](system.ecsWorld, entity)

		// Get legacy entity to access RigidBody (migration phase)
		legacyEntity, exists := system.legacyBridge.GetLegacyEntity(entity)
		if !exists || legacyEntity.Model() == nil || legacyEntity.Model().RigidBody == nil {
			continue
		}

		// Read physics results from Bullet
		transform.Position = legacyEntity.Model().RigidBody.GetTranslation()
		transform.Orientation = legacyEntity.Model().RigidBody.GetOrientation()
	}

	// Sync ECS changes back to legacy entities for systems that still use them
	system.legacyBridge.SyncAllECSToLegacy()
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
		Build()

	for _, entity := range query.Entities() {
		transform, _ := ecs.GetComponent[components.Transform](system.ecsWorld, entity)

		// Get legacy entity to access model (migration phase)
		legacyEntity, exists := system.legacyBridge.GetLegacyEntity(entity)
		if !exists || legacyEntity.Model() == nil || legacyEntity.Model().RigidBody == nil {
			continue
		}

		// Create transformation matrix
		transformMatrix := mgl32.Translate3D(transform.Position.X(), transform.Position.Y(), transform.Position.Z()).
			Mul4(transform.Orientation.Mat4())

		// Add collision mesh for each part of the studiomodel
		for _, r := range system.studiomodelCollisionMeshes[legacyEntity.Model().Model.Id].vertices {
			verts := make([]mgl32.Vec3, 0, len(r))
			for _, v := range r {
				verts = append(verts, mgl32.Vec3{v[0], v[1], v[2]})
			}
			debugBuf.AddLines(verts, mgl32.Vec3{1, 0, 1}, transformMatrix)
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
	for _, e := range system.dataScene.Entities {
		if e.Model() != nil {
			disableMotion := true
			// @TODO Once entity base types are implemented they can be detected better than this
			if strings.HasPrefix(e.Classname(), "prop_physics") {
				disableMotion = false
			}
			system.prepareModelInstanceRigidBody(e.Model(), e.Transform().TransformationMatrix(), disableMotion)

			// Create ECS entities from legacy entities
			if system.legacyBridge != nil {
				system.legacyBridge.CreateECSEntityFromLegacy(e)
			}
		}
	}
	console.PrintString(console.LevelSuccess, "Collision structures ready!")
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

func (system *PhysicsSystem) Cleanup() {
	if system.dataScene == nil {
		return
	}

	// Phase 4: Clear ECS bridge
	if system.legacyBridge != nil {
		system.legacyBridge.Clear()
	}

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
