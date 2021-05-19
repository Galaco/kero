package bullet

// #cgo pkg-config: bullet
// #cgo windows LDFLAGS: -Wl,--allow-multiple-definition
// #include "bulletglue.h"
import "C"
import (
	"github.com/go-gl/mathgl/mgl32"
)

type BulletPhysicsSDKHandle C.plPhysicsSdkHandle
type BulletDynamicWorldHandle C.plDynamicsWorldHandle

func BulletNewPhysicsSDK() BulletPhysicsSDKHandle {
	return BulletPhysicsSDKHandle(C.plNewBulletSdk())
}
func BulletNewDynamicWorld(sdk BulletPhysicsSDKHandle) BulletDynamicWorldHandle {
	return BulletDynamicWorldHandle(C.plCreateDynamicsWorld(sdk))
}

func BulletSetGravity(world BulletDynamicWorldHandle, x,y,z float32) {
	vec3 := vec3ToBullet(mgl32.Vec3{x,y,z})
	C.plSetGravity(world, vec3[0], vec3[1], vec3[2])
}

func BulletStepSimulation(world BulletDynamicWorldHandle, dt float64) {
	C.plStepSimulation(world, C.plReal(dt))
}

func BulletDeleteDynamicWorld(world BulletDynamicWorldHandle) {
	C.plDeleteDynamicsWorld(world)
}

func BulletDeletePhysicsSDK(sdk BulletPhysicsSDKHandle) {
	C.plDeletePhysicsSdk(sdk)
}

// Math
type BulletVec3 C.plVector3
type BulletQuat C.plQuaternion

func vec3ToBullet(vec mgl32.Vec3) (out BulletVec3) {
	out[0] = C.plReal(float64(vec.X()))
	out[1] = C.plReal(float64(vec.Y()))
	out[2] = C.plReal(float64(vec.Z()))

	return out
}

func quatToBullet(quat mgl32.Quat) (out BulletQuat) {
	out[0] = C.plReal(float64(quat.X()))
	out[1] = C.plReal(float64(quat.Y()))
	out[2] = C.plReal(float64(quat.Z()))
	out[3] = C.plReal(float64(quat.W))

	return out
}

func mat4ToBullet(mat mgl32.Mat4) (out [16]C.plReal) {
	for x := 0; x < 16; x++ {
		out[x] = C.plReal(float64(mat[x]))
	}
	return out
}

func mat4FromBullet(mat [16]C.plReal) (out mgl32.Mat4) {
	for x := 0; x < 16; x++ {
		out[x] = float32(mat[x])
	}
	return out
}

// RigidBody
type BulletRigidBodyHandle struct {
	handle C.plRigidBodyHandle
}

func NewRigidBody(mass float32, shape BulletCollisionShapeHandle) BulletRigidBodyHandle {
	body := C.plCreateRigidBody(nil, C.float(mass), shape.handle)
	r := BulletRigidBodyHandle{
		handle: body,
	}
	return r
}

type BulletCollisionShapeHandle struct {
	handle C.plCollisionShapeHandle
}

// AddVertex implements the core.CollisionShape interface
func (c BulletCollisionShapeHandle) AddVertex(v mgl32.Vec3) {
	C.plAddVertex(c.handle, C.plReal(float64(v.X())), C.plReal(float64(v.Y())), C.plReal(float64(v.Z())))
}

func BulletNewConvexHullShape() BulletCollisionShapeHandle {
	return BulletCollisionShapeHandle{
		handle: C.plNewConvexHullShape(),
	}
}

func BulletNewSphericalHullShape(radius float64) BulletCollisionShapeHandle {
	return BulletCollisionShapeHandle{
		handle: C.plNewSphereShape(C.plReal(radius)),
	}
}



func BulletAddRigidBody(world BulletDynamicWorldHandle, handle BulletRigidBodyHandle) {
	C.plAddRigidBody(world, handle.handle)
}

func BulletRemoveRigidBody(world BulletDynamicWorldHandle, handle BulletRigidBodyHandle) {
	C.plRemoveRigidBody(world, handle.handle)
}

func BulletDeleteRigidBody(handle BulletRigidBodyHandle) {
	C.plDeleteRigidBody(handle.handle)
}

func BulletGetOpenGLMatrix(handle BulletRigidBodyHandle) mgl32.Mat4{
	mat := mat4ToBullet(mgl32.Ident4())
	C.plGetOpenGLMatrix(handle.handle, &mat[0])
	return mat4FromBullet(mat)
}

func BulletSetOpenGLMatrix(handle BulletRigidBodyHandle, transform mgl32.Mat4) {
	mat := mat4ToBullet(transform)
	C.plSetOpenGLMatrix(handle.handle, &mat[0])
}

func BulletApplyImpulse(handle BulletRigidBodyHandle, impulse, localPoint mgl32.Vec3) {
	i := vec3ToBullet(impulse)
	p := vec3ToBullet(localPoint)
	C.plApplyImpulse(handle.handle, &i[0], &p[0])
}