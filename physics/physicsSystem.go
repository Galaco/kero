package physics

import (
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/physics/collision"
	"github.com/galaco/kero/framework/physics/collision/bullet"
	"github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/messages"
	"github.com/go-gl/mathgl/mgl32"
)

type PhysicsSystem struct {
	dataScene *scene.StaticScene

	physicsEntities []entity.IEntity

	// Bullet
	sdk   bullet.BulletPhysicsSDKHandle
	world bullet.BulletDynamicWorldHandle
}

func (system *PhysicsSystem) Initialize() {
	event.Get().AddListener(messages.TypeLoadingLevelParsed, system.onLoadingLevelParsed)

	// create an sdk handle
	system.sdk = bullet.BulletNewPhysicsSDK()

	// instance a world
	system.world = bullet.BulletNewDynamicWorld(system.sdk)
	bullet.BulletSetGravity(system.world, 0.0, 0.0, -100.0)
}

func (system *PhysicsSystem) Cleanup() {
	bullet.BulletDeleteDynamicWorld(system.world)
	bullet.BulletDeletePhysicsSDK(system.sdk)
}

func (system *PhysicsSystem) Update(dt float64) {
	if len(system.physicsEntities) == 0 {
		// Nothing to simulate
		return
	}


	for _, n := range system.physicsEntities {
		if n.Model().RigidBody() != nil {
			n.Model().RigidBody().SetTransform(n.Transform().TransformationMatrix())
		}
	}
	bullet.BulletStepSimulation(system.world, dt)
	for _, n := range system.physicsEntities {
		if n.Model().RigidBody() == nil {
			continue
		}
		trans := n.Model().RigidBody().GetTransform()
		n.Transform().Position = mgl32.Vec3{
			trans[12],
			trans[13],
			trans[14],
		}
	}
}

func (system *PhysicsSystem) onLoadingLevelParsed(message interface{}) {
	system.dataScene = message.(*messages.LoadingLevelParsed).Level().(*scene.StaticScene)

	// Find entities that have a model
	for idx,e := range system.dataScene.Entities {
		if e.Model() != nil {
			// Prepare Bullet environment for collision meshes
			system.dataScene.Entities[idx].Model().AddRigidBody(collision.NewSphericalHull(32))
			bullet.BulletAddRigidBody(system.world, system.dataScene.Entities[idx].Model().RigidBody().BulletHandle())
			system.physicsEntities = append(system.physicsEntities, system.dataScene.Entities[idx])
		}
	}
}

func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{
		physicsEntities: make([]entity.IEntity, 0),
	}
}