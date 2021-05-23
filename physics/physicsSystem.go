package physics

import (
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/graphics/adapter"
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

	bspRigidBody               *bspCollisionMesh
	studiomodelCollisionMeshes map[string]studiomodelCollisionMesh
}

func (system *PhysicsSystem) Initialize() {
	event.Get().AddListener(messages.TypeChangeLevel, system.onChangeLevel)
	event.Get().AddListener(messages.TypeLoadingLevelParsed, system.onLoadingLevelParsed)

	console.AddConvarBool("r_drawcollisionmodels", "Render collision mode vertices", false)

	// create an sdk handle
	system.sdk = bullet.BulletNewPhysicsSDK()

	// instance a world
	system.world = bullet.BulletNewDynamicWorld(system.sdk)
	bullet.BulletSetGravity(system.world, 0.0, 0.0, -100.0)
}

func (system *PhysicsSystem) Cleanup() {
	for _, i := range system.physicsEntities {
		bullet.BulletDeleteRigidBody(i.Model().RigidBody.BulletHandle())
		i.Model().RigidBody = nil
	}
	if system.bspRigidBody != nil {
		for _, r := range system.bspRigidBody.RigidBodyHandles {
			bullet.BulletDeleteRigidBody(r)
		}
	}

	bullet.BulletDeleteDynamicWorld(system.world)
	bullet.BulletDeletePhysicsSDK(system.sdk)
}

func (system *PhysicsSystem) Update(dt float64) {
	if len(system.physicsEntities) == 0 {
		// Nothing to simulate
		return
	}

	for _, n := range system.physicsEntities {
		if n.Model().RigidBody == nil {
			continue
		}
		n.Model().RigidBody.SetTransform(n.Transform().TransformationMatrix())
	}
	bullet.BulletStepSimulation(system.world, dt)
	for idx, n := range system.physicsEntities {
		if n.Model().RigidBody == nil {
			continue
		}
		trans := n.Model().RigidBody.GetTransform()
		system.physicsEntities[idx].Transform().Position = mgl32.Vec3{
			trans[12],
			trans[13],
			trans[14],
		}
	}

	if console.GetConvarBoolean("r_drawcollisionmodels") == true {
		system.drawDebug()
	}
}

func (system *PhysicsSystem) drawDebug() {
	if adapter.CurrentShader() == nil {
		return
	}
	adapter.EnableFrontFaceCulling()
	adapter.DisableDepthTesting()

	adapter.PushMat4(adapter.CurrentShader().GetUniform("model"), 1, false, mgl32.Ident4())
	verts := make([]float32, 0)
	for _, vs := range system.bspRigidBody.vertices {
		for _, vert := range vs {
			verts = append(verts, vert[0], vert[1], vert[2])
		}
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

func (system *PhysicsSystem) onChangeLevel(message interface{}) {
	if system.dataScene == nil {
		return
	}

	for _, i := range system.physicsEntities {
		bullet.BulletDeleteRigidBody(i.Model().RigidBody.BulletHandle())
		i.Model().RigidBody = nil
	}
	for _, r := range system.bspRigidBody.RigidBodyHandles {
		bullet.BulletDeleteRigidBody(r)
	}
	system.dataScene = nil
	system.bspRigidBody = nil
}

func (system *PhysicsSystem) onLoadingLevelParsed(message interface{}) {
	system.dataScene = message.(*messages.LoadingLevelParsed).Level().(*scene.StaticScene)

	console.PrintString(console.LevelInfo, "Generating collision structures....")

	// Generate BSP Rigidbody
	system.bspRigidBody = generateBspCollisionMesh(system.dataScene)
	for _, r := range system.bspRigidBody.RigidBodyHandles {
		bullet.BulletAddRigidBody(system.world, r)
	}
	console.PrintString(console.LevelInfo, "BSP collision structure ready!")

	// Find entities that have a model
	for idx, e := range system.dataScene.Entities {
		if e.Model() != nil {
			// Prepare Bullet environment for collision meshes
			if e.Model() != nil && e.Model().Model.OriginalStudiomodel.Phy != nil {
				// We have an actual source engine .phy collision model
				if _, ok := system.studiomodelCollisionMeshes[e.Model().Model.Id]; !ok {
					system.studiomodelCollisionMeshes[e.Model().Model.Id] = generateCollisionMeshFromStudiomodelPhy(e.Model().Model.OriginalStudiomodel.Phy)
				}
				system.dataScene.Entities[idx].Model().RigidBody = collision.NewConvexHullFromExistingShape(
					1,
					system.studiomodelCollisionMeshes[e.Model().Model.Id].compoundShapeHandle)
			} else {
				// Fall back to generating one
				system.dataScene.Entities[idx].Model().RigidBody = collision.NewSphericalHull(4)
			}

			system.dataScene.Entities[idx].Model().RigidBody.SetTransform(e.Transform().TransformationMatrix())
			bullet.BulletAddRigidBody(system.world, system.dataScene.Entities[idx].Model().RigidBody.BulletHandle())
			system.physicsEntities = append(system.physicsEntities, system.dataScene.Entities[idx])
		}
	}
	console.PrintString(console.LevelInfo, "Physics prop collision structures ready!")
	console.PrintString(console.LevelInfo, "Collision structure ready!")
}

func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{
		physicsEntities:            make([]entity.IEntity, 0),
		studiomodelCollisionMeshes: map[string]studiomodelCollisionMesh{},
	}
}
