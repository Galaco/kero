package physics

import (
	"fmt"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
	"github.com/galaco/kero/framework/ecs/legacy"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/graphics/adapter"
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

	// Dynamic entities (includes prop_physics* & prop_dynamic*)
	// TODO: Phase 4 - deprecated, use ECS queries instead
	physicsEntities []entity.IEntity

	// Bullet
	sdk   bullet.BulletPhysicsSDKHandle
	world bullet.BulletDynamicWorldHandle

	bspRigidBody               *bspCollisionMesh
	displacementRigidBody      *displacementCollisionMesh
	studiomodelCollisionMeshes map[string]studiomodelCollisionMesh
}

func (system *PhysicsSystem) Initialize() {
	// Phase 4: Initialize ECS bridge if ECS World is set
	if system.ecsWorld != nil {
		system.legacyBridge = legacy.NewBridge(system.ecsWorld)
	}

	// Register typed event listeners (Phase 3)
	event.RegisterTypedEvent(system.eventBus, system.onChangeLevelTyped)
	event.RegisterTypedEvent(system.eventBus, system.onLoadingLevelParsedTyped)
	event.RegisterTypedEvent(system.eventBus, func(e messages.EngineDisconnectEvent) {
		system.Cleanup()
	})

	// Physics debug and monitoring console variables
	console.AddConvarBool("r_drawcollisionmodels", "Render collision mode vertices", false)
	console.AddConvarBool("physics_debug", "Show physics timing debug info", false)
	console.AddConvarBool("physics_use_ecs", "Use ECS for physics updates (Phase 4)", true)
}

// FixedUpdate runs the physics simulation at a fixed timestep.
// This method should be called with a constant dt value (typically 1.0/60.0 = 0.0166667 seconds).
// Fixed timestep ensures stable, deterministic physics simulation regardless of frame rate.
func (system *PhysicsSystem) FixedUpdate(dt float64) {
	// Phase 4: Use ECS path if enabled and available
	if console.GetConvarBoolean("physics_use_ecs") && system.ecsWorld != nil && system.legacyBridge != nil {
		system.fixedUpdateECS(dt)
		return
	}

	// Legacy path (backward compatibility)
	system.fixedUpdateLegacy(dt)
}

// fixedUpdateECS updates physics using ECS queries (Phase 4)
func (system *PhysicsSystem) fixedUpdateECS(dt float64) {
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
		console.PrintString(console.LevelInfo, fmt.Sprintf("Physics dt: %.6f (ECS mode)", dt))
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

	// Debug visualization
	if console.GetConvarBoolean("r_drawcollisionmodels") {
		system.drawDebug()
	}
}

// fixedUpdateLegacy updates physics using legacy entity list (backward compatibility)
func (system *PhysicsSystem) fixedUpdateLegacy(dt float64) {
	if len(system.physicsEntities) == 0 {
		// Nothing to simulate
		return
	}

	if !input.Keyboard().IsKeyPressed(input.KeyQ) {
		return
	}

	// Debug logging (useful to verify fixed timestep is working)
	if console.GetConvarBoolean("physics_debug") {
		console.PrintString(console.LevelInfo, fmt.Sprintf("Physics dt: %.6f (legacy mode)", dt))
	}

	// Update entity transforms to physics engine
	for _, n := range system.physicsEntities {
		if n.Model().RigidBody == nil {
			continue
		}
		n.Model().RigidBody.SetTransform(n.Transform().TransformationMatrix())
	}

	// Step physics simulation with fixed timestep
	bullet.BulletStepSimulation(system.world, dt)

	// Apply physics results back to entity transforms
	for idx, n := range system.physicsEntities {
		if n.Model().RigidBody == nil {
			continue
		}
		system.physicsEntities[idx].Transform().Translation = n.Model().RigidBody.GetTranslation()
		system.physicsEntities[idx].Transform().Orientation = n.Model().RigidBody.GetOrientation()
	}

	// Debug visualization
	if console.GetConvarBoolean("r_drawcollisionmodels") {
		system.drawDebug()
	}
}

// Update is a wrapper around FixedUpdate for backward compatibility.
// Deprecated: Use FixedUpdate for physics simulation with fixed timestep.
// This method delegates to FixedUpdate but should not be used directly.
func (system *PhysicsSystem) Update(dt float64) {
	system.FixedUpdate(dt)
}

func (system *PhysicsSystem) drawDebug() {
	if adapter.CurrentShader() == nil {
		return
	}
	adapter.EnableFrontFaceCulling()
	adapter.DisableDepthTesting()

	adapter.PushMat4(adapter.CurrentShader().GetUniform("model"), 1, false, mgl32.Ident4())
	verts := make([]float32, 0)
	for _, vert := range system.bspRigidBody.vertices {
		verts = append(verts, vert[0], vert[1], vert[2])
	}
	adapter.DrawDebugLines(verts, mgl32.Vec3{255, 0, 255})

	for _, n := range system.physicsEntities {
		if n.Model().RigidBody == nil {
			continue
		}
		adapter.PushMat4(adapter.CurrentShader().GetUniform("model"), 1, false, n.Transform().TransformationMatrix())
		for _, r := range system.studiomodelCollisionMeshes[n.Model().Model.Id].vertices {
			verts := make([]float32, 0)
			for _, v := range r {
				verts = append(verts, v[0], v[1], v[2])
			}
			adapter.DrawDebugLines(verts, mgl32.Vec3{255, 0, 255})
		}
	}
	adapter.EnableDepthTesting()
	adapter.EnableBackFaceCulling()
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
			system.physicsEntities = append(system.physicsEntities, e)

			// Phase 4: Create ECS entities from legacy entities
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

	for _, i := range system.physicsEntities {
		if i.Model() == nil || i.Model().RigidBody == nil {
			continue
		}
		bullet.BulletDeleteRigidBody(i.Model().RigidBody.BulletHandle())
		i.Model().RigidBody = nil
	}
	bullet.BulletDeleteRigidBody(system.bspRigidBody.RigidBodyHandles)

	if system.displacementRigidBody != nil {
		bullet.BulletDeleteRigidBody(system.displacementRigidBody.RigidBodyHandles)
	}
	system.physicsEntities = make([]entity.IEntity, 0)
	system.dataScene = nil
	system.bspRigidBody = nil
	system.displacementRigidBody = nil
}

// NewPhysicsSystem creates a new physics system with explicit dependencies
func NewPhysicsSystem(eventBus *event.Dispatcher, sceneManager *scene.Manager, ecsWorld *ecs.World) *PhysicsSystem {
	return &PhysicsSystem{
		eventBus:                   eventBus,
		sceneManager:               sceneManager,
		ecsWorld:                   ecsWorld,
		physicsEntities:            make([]entity.IEntity, 0),
		studiomodelCollisionMeshes: map[string]studiomodelCollisionMesh{},
	}
}
