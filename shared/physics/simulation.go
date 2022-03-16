package physics

import (
	"github.com/galaco/kero/internal/framework/console"
	"github.com/galaco/kero/internal/framework/entity"
	"github.com/galaco/kero/internal/framework/event"
	"github.com/galaco/kero/internal/framework/graphics/mesh"
	"github.com/galaco/kero/internal/framework/physics/collision"
	"github.com/galaco/kero/internal/framework/physics/collision/bullet"
	"github.com/galaco/kero/internal/framework/scene"
	"github.com/galaco/kero/shared/messages"
	"github.com/go-gl/mathgl/mgl32"
	"strings"
)

const (
	updateRate = 1 / 12
)

type Simulation struct {
	timeSinceLastUpdate float64

	// Dynamic entities (includes prop_physics* & prop_dynamic*)
	physicsEntities []entity.IEntity

	// Bullet
	sdk   bullet.BulletPhysicsSDKHandle
	world bullet.BulletDynamicWorldHandle

	bspRigidBody               *bspCollisionMesh
	displacementRigidBody      *displacementCollisionMesh
	studiomodelCollisionMeshes map[string]studiomodelCollisionMesh
}

func (system *Simulation) Initialize() {
	event.Get().AddListener(messages.TypeChangeLevel, system.onChangeLevel)
	event.Get().AddListener(messages.TypeLoadingLevelParsed, system.onLoadingLevelParsed)

	event.Get().AddListener(messages.TypeEngineDisconnect, func(e interface{}) {
		system.Cleanup()
	})
}

func (system *Simulation) Update(dt float64) {
	// Collisions don't need to run every frame. Aim for 12 times per second
	if system.timeSinceLastUpdate <= updateRate {
		system.timeSinceLastUpdate += dt
		return
	}

	if len(system.physicsEntities) == 0 {
		// Nothing to simulate
		return
	}

	// Setup world frame
	for _, n := range system.physicsEntities {
		if n.Model().RigidBody == nil {
			continue
		}
		n.Model().RigidBody.SetTransform(n.Transform().TransformationMatrix())
	}
	// Run simulation
	bullet.BulletStepSimulation(system.world, system.timeSinceLastUpdate)

	// Apply changes
	for idx, n := range system.physicsEntities {
		if n.Model().RigidBody == nil {
			continue
		}
		system.physicsEntities[idx].Transform().Translation = n.Model().RigidBody.GetTranslation()
		system.physicsEntities[idx].Transform().Orientation = n.Model().RigidBody.GetOrientation()
	}

	system.timeSinceLastUpdate = 0
}

func (system *Simulation) onChangeLevel(message interface{}) {
	system.Cleanup()
}

func (system *Simulation) onLoadingLevelParsed(message interface{}) {
	dataScene := message.(*messages.LoadingLevelParsed).Level().(*scene.StaticScene)

	// create an sdk handle
	system.sdk = bullet.BulletNewPhysicsSDK()
	// instance a world
	system.world = bullet.BulletNewDynamicWorld(system.sdk)
	bullet.BulletSetGravity(system.world, 0.0, 0.0, -100.0)

	console.PrintString(console.LevelInfo, "Generating collision structures....")

	// Generate BSP Rigidbody
	console.PrintString(console.LevelInfo, "BSP collision structure...")
	system.bspRigidBody = generateBspCollisionMesh(dataScene)
	bullet.BulletAddRigidBody(system.world, system.bspRigidBody.RigidBodyHandles)

	// Generate Displacement RigidBodies
	console.PrintString(console.LevelInfo, "Displacement collision structures...")
	system.displacementRigidBody = generateDisplacementCollisionMeshes(dataScene)
	if system.displacementRigidBody != nil {
		bullet.BulletAddRigidBody(system.world, system.displacementRigidBody.RigidBodyHandles)
	}

	// Generate Staticprop RigidBodies
	console.PrintString(console.LevelInfo, "Static prop collision structures...")
	for _, e := range dataScene.StaticProps {
		system.prepareModelInstanceRigidBody(e.Model(), e.Transform.TransformationMatrix(), true)
	}

	// Find entities that have a model
	console.PrintString(console.LevelInfo, "Physics prop collision structures...")
	for _, e := range dataScene.Entities {
		if e.Model() != nil {
			disableMotion := true
			// @TODO Once entity base types are implemented they can be detected better than this
			if strings.HasPrefix(e.Classname(), "prop_physics") {
				disableMotion = false
			}
			system.prepareModelInstanceRigidBody(e.Model(), e.Transform().TransformationMatrix(), disableMotion)
			system.physicsEntities = append(system.physicsEntities, e)
		}
	}
	console.PrintString(console.LevelSuccess, "Collision structures ready!")
}

func (system *Simulation) prepareModelInstanceRigidBody(model *mesh.ModelInstance, initialTransformation mgl32.Mat4, isStatic bool) {
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

func (system *Simulation) Cleanup() {
	if system.bspRigidBody == nil {
		// Because we cannot have a collisionless BSP
		return
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
	system.bspRigidBody = nil
	system.displacementRigidBody = nil
}

func NewSimulation() *Simulation {
	return &Simulation{
		physicsEntities:            make([]entity.IEntity, 0),
		studiomodelCollisionMeshes: map[string]studiomodelCollisionMesh{},
	}
}
