package physics

import (
	"github.com/galaco/kero/framework/console"
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
	dataScene *scene.StaticScene

	// Dynamic entities (includes prop_physics* & prop_dynamic*)
	physicsEntities []entity.IEntity

	// Bullet
	sdk   bullet.BulletPhysicsSDKHandle
	world bullet.BulletDynamicWorldHandle

	bspRigidBody               *bspCollisionMesh
	displacementRigidBody      *displacementCollisionMesh
	studiomodelCollisionMeshes map[string]studiomodelCollisionMesh
}

func (system *PhysicsSystem) Initialize() {
	event.Get().AddListener(messages.TypeChangeLevel, system.onChangeLevel)
	event.Get().AddListener(messages.TypeLoadingLevelParsed, system.onLoadingLevelParsed)

	event.Get().AddListener(messages.TypeEngineDisconnect, func(e interface{}) {
		system.Cleanup()
	})

	console.AddConvarBool("r_drawcollisionmodels", "Render collision mode vertices", false)
}

func (system *PhysicsSystem) Update(dt float64) {
	if len(system.physicsEntities) == 0 {
		// Nothing to simulate
		return
	}

	if !input.Keyboard().IsKeyPressed(input.KeyQ) {
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
		system.physicsEntities[idx].Transform().Translation = n.Model().RigidBody.GetTranslation()
		system.physicsEntities[idx].Transform().Orientation = n.Model().RigidBody.GetOrientation()
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

func (system *PhysicsSystem) onChangeLevel(message interface{}) {
	if system.dataScene == nil {
		return
	}
	system.Cleanup()
}

func (system *PhysicsSystem) onLoadingLevelParsed(message interface{}) {
	system.dataScene = message.(*messages.LoadingLevelParsed).Level().(*scene.StaticScene)

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

func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{
		physicsEntities:            make([]entity.IEntity, 0),
		studiomodelCollisionMeshes: map[string]studiomodelCollisionMesh{},
	}
}
