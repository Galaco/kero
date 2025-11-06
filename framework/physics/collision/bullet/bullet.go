package bullet

// #cgo pkg-config: bullet
// #cgo windows LDFLAGS: -Wl,--allow-multiple-definition
// #include "bulletglue.h"
import "C"
import (
	"github.com/go-gl/mathgl/mgl32"
	"unsafe"
)

// Bullet activation states
const (
	ActivationStateActiveTag          = 1
	ActivationStateIslandSleeping     = 2
	ActivationStateWantsDeactivation  = 3
	ActivationStateDisableDeactivation = 4
	ActivationStateDisableSimulation   = 5
)

type BulletPhysicsIndice C.int
type BulletPhysicsSDKHandle C.plPhysicsSdkHandle
type BulletDynamicWorldHandle C.plDynamicsWorldHandle

func BulletNewPhysicsSDK() BulletPhysicsSDKHandle {
	return BulletPhysicsSDKHandle(C.plNewBulletSdk())
}
func BulletNewDynamicWorld(sdk BulletPhysicsSDKHandle) BulletDynamicWorldHandle {
	return BulletDynamicWorldHandle(C.plCreateDynamicsWorld(sdk))
}

func BulletSetGravity(world BulletDynamicWorldHandle, x, y, z float32) {
	vec3 := Vec3ToBullet(mgl32.Vec3{x, y, z})
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

func Vec3ToBullet(vec mgl32.Vec3) (out BulletVec3) {
	out[0] = C.plReal(float64(vec.X()))
	out[1] = C.plReal(float64(vec.Y()))
	out[2] = C.plReal(float64(vec.Z()))

	return out
}

func vec3FromBullet(vec BulletVec3) (out mgl32.Vec3) {
	out[0] = float32(vec[0])
	out[1] = float32(vec[1])
	out[2] = float32(vec[2])

	return out
}

func quatToBullet(quat mgl32.Quat) (out BulletQuat) {
	out[0] = C.plReal(float64(quat.X()))
	out[1] = C.plReal(float64(quat.Y()))
	out[2] = C.plReal(float64(quat.Z()))
	out[3] = C.plReal(float64(quat.W))

	return out
}

func quatFromBullet(quat BulletQuat) (out mgl32.Quat) {
	out.V[0] = float32(quat[0])
	out.V[1] = float32(quat[1])
	out.V[2] = float32(quat[2])
	out.W = float32(quat[3])

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
	holder       interface{} // Stores vertices to prevent GC
	indexHolder  interface{} // Stores indices to prevent GC
	triangleArrayHandle interface{} // Stores btTriangleIndexVertexArray handle
	handle       C.plCollisionShapeHandle
}

func (c BulletCollisionShapeHandle) AddVertex(v mgl32.Vec3) {
	C.plAddVertex(c.handle, C.plReal(float64(v.X())), C.plReal(float64(v.Y())), C.plReal(float64(v.Z())))
}

func (c BulletCollisionShapeHandle) AddVertices(verts []mgl32.Vec3) {
	v := make([]BulletVec3, len(verts))
	for idx, r := range verts {
		v[idx] = Vec3ToBullet(r)
	}

	C.plAddVertices(c.handle, (*C.plVector3)(unsafe.Pointer(&v[0])), C.int(len(v)))
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

func BulletNewCapsuleShapeZ(radius, height float64) BulletCollisionShapeHandle {
	return BulletCollisionShapeHandle{
		handle: C.plNewCapsuleShapeZ(C.plReal(radius), C.plReal(height)),
	}
}

func BulletNewCompoundShape() BulletCollisionShapeHandle {
	return BulletCollisionShapeHandle{
		handle: C.plNewCompoundShape(),
	}
}

func BulletNewStaticPlaneShape(plane mgl32.Vec3, constant float64) BulletCollisionShapeHandle {
	p := Vec3ToBullet(plane)
	return BulletCollisionShapeHandle{
		handle: C.plNewStaticPlaneShape(&p[0], C.float(constant)),
	}
}

func BulletNewStaticTriangleShape(indices []BulletPhysicsIndice, vertices []mgl32.Vec3, totalTriangles, totalVerts int) BulletCollisionShapeHandle {
	v := make([]BulletVec3, len(vertices))
	for idx, r := range vertices {
		v[idx] = Vec3ToBullet(r)
	}

	m := C.btNewBvhTriangleIndexVertexArray((*C.int)(unsafe.Pointer(&indices[0])), (*C.plVector3)(unsafe.Pointer(&v[0])), C.int(int64(totalTriangles)), C.int(int64(totalVerts)))

	return BulletCollisionShapeHandle{
		holder:              v,       // Keep vertices alive
		indexHolder:         indices, // Keep indices alive
		triangleArrayHandle: m,       // Keep triangle array handle alive
		handle:              C.btNewBvhTriangleMeshShape(m),
	}
}

func BulletNewBrushShape(vertices []mgl32.Vec3) BulletCollisionShapeHandle {
	shape := BulletNewConvexHullShape()
	shape.AddVertices(vertices)

	return BulletCollisionShapeHandle{
		handle: shape.handle,
	}
}

func BulletAddChildToCompoundShape(parent BulletCollisionShapeHandle, s BulletCollisionShapeHandle, p mgl32.Vec3, o mgl32.Quat) {
	vec := Vec3ToBullet(p)
	quat := quatToBullet(o)
	C.plAddChildShape(parent.handle, s.handle, &vec[0], &quat[0])
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

func BulletGetOpenGLMatrix(handle BulletRigidBodyHandle) mgl32.Mat4 {
	mat := mat4ToBullet(mgl32.Ident4())
	C.plGetOpenGLMatrix(handle.handle, &mat[0])
	return mat4FromBullet(mat)
}

func BulletSetOpenGLMatrix(handle BulletRigidBodyHandle, transform mgl32.Mat4) {
	mat := mat4ToBullet(transform)
	C.plSetOpenGLMatrix(handle.handle, &mat[0])
}

func BulletGetTranslation(handle BulletRigidBodyHandle) mgl32.Vec3 {
	vec := Vec3ToBullet(mgl32.Vec3{})
	C.plGetPosition(handle.handle, &vec[0])
	return vec3FromBullet(vec)
}

func BulletGetOrientation(handle BulletRigidBodyHandle) mgl32.Quat {
	quat := quatToBullet(mgl32.Quat{})
	C.plGetOrientation(handle.handle, &quat[0])
	return quatFromBullet(quat)
}

func BulletApplyImpulse(handle BulletRigidBodyHandle, impulse, localPoint mgl32.Vec3) {
	i := Vec3ToBullet(impulse)
	p := Vec3ToBullet(localPoint)
	C.plApplyImpulse(handle.handle, &i[0], &p[0])
}

func BulletSetActivationState(handle BulletRigidBodyHandle, state int) {
	C.plSetActivationState(handle.handle, C.int(state))
}

func BulletForceActivationState(handle BulletRigidBodyHandle, state int) {
	C.plForceActivationState(handle.handle, C.int(state))
}

// SweepResult represents the result of a convex sweep test
type SweepResult struct {
	HasHit      bool
	HitPoint    mgl32.Vec3
	HitNormal   mgl32.Vec3
	HitFraction float32
}

// RaycastResult represents the result of a raycast
type RaycastResult struct {
	HasHit      bool
	HitPoint    mgl32.Vec3
	HitNormal   mgl32.Vec3
	HitFraction float32
}

// BulletConvexSweepTest sweeps a convex shape from one position to another
func BulletConvexSweepTest(world BulletDynamicWorldHandle, shape BulletCollisionShapeHandle, from, to mgl32.Vec3) SweepResult {
	fromVec := Vec3ToBullet(from)
	toVec := Vec3ToBullet(to)

	var result C.plSweepResult
	C.plConvexSweepTest(world, shape.handle, &fromVec[0], &toVec[0], &result)

	// Debug: log raw C result
	// fmt.Printf("C sweep result: hasHit=%d, fraction=%.3f\n", result.hasHit, result.hitFraction)

	return SweepResult{
		HasHit:      result.hasHit != 0,
		HitPoint:    vec3FromBullet(BulletVec3(result.hitPoint)),
		HitNormal:   vec3FromBullet(BulletVec3(result.hitNormal)),
		HitFraction: float32(result.hitFraction),
	}
}

// BulletRayTest performs a raycast test
func BulletRayTest(world BulletDynamicWorldHandle, from, to mgl32.Vec3) RaycastResult {
	fromVec := Vec3ToBullet(from)
	toVec := Vec3ToBullet(to)

	var result C.plRaycastResult
	C.plRayTest(world, &fromVec[0], &toVec[0], &result)

	return RaycastResult{
		HasHit:      result.hasHit != 0,
		HitPoint:    vec3FromBullet(BulletVec3(result.hitPoint)),
		HitNormal:   vec3FromBullet(BulletVec3(result.hitNormal)),
		HitFraction: float32(result.hitFraction),
	}
}
