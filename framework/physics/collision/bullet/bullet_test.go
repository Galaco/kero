package bullet

import (
	"math"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

// TestWorldLifecycle builds and destroys a world the way changing level does, several times over
func TestWorldLifecycle(t *testing.T) {
	for level := 0; level < 3; level++ {
		sdk := BulletNewPhysicsSDK()
		world := BulletNewDynamicWorld(sdk)
		BulletSetGravity(world, 0, 0, -100)

		// A floor at z=0, made from a mesh Bullet must keep its own copy of
		indices := []BulletPhysicsIndice{0, 1, 2, 0, 2, 3}
		vertices := []mgl32.Vec3{{-100, -100, 0}, {100, -100, 0}, {100, 100, 0}, {-100, 100, 0}}
		floor := BulletNewStaticTriangleShape(indices, vertices, 2, 4)
		for i := range vertices {
			vertices[i] = mgl32.Vec3{}
		}
		for i := range indices {
			indices[i] = 0
		}

		brush := BulletNewConvexHullShape()
		brush.AddVertices([]mgl32.Vec3{{50, 50, 0}, {60, 50, 0}, {50, 60, 0}, {50, 50, 10}})
		brush.AddVertices(nil)
		compound := BulletNewCompoundShape()
		BulletAddChildToCompoundShape(compound, brush, mgl32.Vec3{}, mgl32.QuatIdent())
		sphere := BulletNewSphericalHullShape(4)
		capsule := BulletNewCapsuleShapeZ(16, 40)
		shapes := []BulletCollisionShapeHandle{floor, compound, brush, sphere, capsule}

		bodies := []BulletRigidBodyHandle{NewRigidBody(0, floor), NewRigidBody(0, compound), NewRigidBody(1, sphere)}
		BulletSetOpenGLMatrix(bodies[2], mgl32.Translate3D(-50, -50, 20))
		for _, body := range bodies {
			BulletAddRigidBody(world, body)
		}
		for i := 0; i < 10; i++ {
			BulletStepSimulation(world, 1.0/60)
		}

		ray := BulletRayTest(world, mgl32.Vec3{10, -20, 10}, mgl32.Vec3{10, -20, -10})
		if !ray.HasHit || math.Abs(float64(ray.HitPoint.Z())) > 0.01 {
			t.Errorf("level %d: expected ray to hit the floor at z=0, got %+v", level, ray)
		}
		sweep := BulletConvexSweepTest(world, capsule, mgl32.Vec3{10, -20, 100}, mgl32.Vec3{10, -20, 0})
		if !sweep.HasHit || sweep.HitNormal.Z() < 0.99 {
			t.Errorf("level %d: expected capsule to land on the floor, got %+v", level, sweep)
		}

		for _, body := range bodies {
			BulletRemoveRigidBody(world, body)
		}
		BulletDeleteDynamicWorld(world)
		for _, body := range bodies {
			BulletDeleteRigidBody(body)
		}
		for _, shape := range shapes {
			BulletDeleteShape(shape)
		}
		BulletDeletePhysicsSDK(sdk)
	}
}
