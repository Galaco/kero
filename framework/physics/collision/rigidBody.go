package collision

import (
	"github.com/galaco/kero/framework/physics/collision/bullet"
	"github.com/galaco/studiomodel"
	"github.com/go-gl/mathgl/mgl32"
)

type CollisionBodyType int8

const (
	RigidBodyTypeConvexHull             = CollisionBodyType(0)
	RigidBodyTypeOrientedBoundingBox    = CollisionBodyType(1)
	RigidBodyTypeAxisAlignedBoundingBox = CollisionBodyType(2)
)

type RigidBody interface {
	CollisionBodyType() CollisionBodyType
	BulletHandle() bullet.BulletRigidBodyHandle

	GetTransform() mgl32.Mat4
	SetTransform(transform mgl32.Mat4)
	ApplyImpulse(impulse mgl32.Vec3, localPoint mgl32.Vec3)
}


type ConvexHull struct {
	handle bullet.BulletRigidBodyHandle
}

func (body *ConvexHull) CollisionBodyType() CollisionBodyType {
	return RigidBodyTypeConvexHull
}

func (body *ConvexHull) BulletHandle() bullet.BulletRigidBodyHandle {
	return body.handle
}

func (body *ConvexHull) GetTransform() mgl32.Mat4 {
	return bullet.BulletGetOpenGLMatrix(body.handle)
}

// SetTransform implements the core.RigidBody interface
func (body *ConvexHull) SetTransform(transform mgl32.Mat4) {
	bullet.BulletSetOpenGLMatrix(body.handle, transform)
}

// ApplyImpulse implements the core.RigidBody interface
func (body *ConvexHull) ApplyImpulse(impulse mgl32.Vec3, localPoint mgl32.Vec3) {
	bullet.BulletApplyImpulse(body.handle, impulse, localPoint)
}

func NewConvexHull() *ConvexHull {
	cbody := new(ConvexHull)

	h := bullet.BulletNewConvexHullShape()
	cbody.handle = bullet.NewRigidBody(1, h)

	return cbody
}

func NewSphericalHull(radius float64) *ConvexHull {
	cbody := new(ConvexHull)

	h := bullet.BulletNewSphericalHullShape(radius)
	cbody.handle = bullet.NewRigidBody(1, h)

	return cbody
}

type OrientedBoundingBox struct {

}

func (body *OrientedBoundingBox) CollisionBodyType() CollisionBodyType {
	return RigidBodyTypeOrientedBoundingBox
}

type AxisAlignedBoundingBox struct {
	Mins, Maxs mgl32.Vec3
}

func (body *AxisAlignedBoundingBox) CollisionBodyType() CollisionBodyType {
	return RigidBodyTypeAxisAlignedBoundingBox
}

func NewAxisAlignedBoundingBox(m *studiomodel.StudioModel) *AxisAlignedBoundingBox {
	return &AxisAlignedBoundingBox{
		Mins: m.Mdl.Header.ViewBBMin,
		Maxs: m.Mdl.Header.ViewBBMax,
	}
}