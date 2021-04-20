package graphics

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Transform Represents the transformation of an entity in
// a 3-dimensional space: position, rotation and scale.
// Note: Rotation is measured in degrees
type Transform struct {
	Position mgl32.Vec3
	Rotation mgl32.Vec3
	Scale    mgl32.Vec3

	prevPosition mgl32.Vec3
	prevRotation mgl32.Vec3
	prevScale    mgl32.Vec3
	matrix       mgl32.Mat4
	quat         mgl32.Quat
}

// TransformationMatrix computes object transformation matrix
func (transform *Transform) TransformationMatrix() mgl32.Mat4 {
	if !transform.Position.ApproxEqual(transform.prevPosition) ||
		!transform.Rotation.ApproxEqual(transform.prevRotation) ||
		!transform.Scale.ApproxEqual(transform.prevScale) {

		transform.quat = mgl32.QuatIdent()

		// Scale of 0 is invalid
		if transform.Scale.X() == 0 ||
			transform.Scale.Y() == 0 ||
			transform.Scale.Z() == 0 {
			transform.Scale = mgl32.Vec3{1, 1, 1}
		}

		//Translate
		translation := mgl32.Translate3D(transform.Position.X(), transform.Position.Y(), transform.Position.Z())

		// rotate
		// IMPORTANT. Source engine has X and Z axis switched
		rotation := mgl32.Ident4()
		rotation = transform.rotateAroundAxis(rotation, mgl32.Vec3{1, 0, 0}, mgl32.DegToRad(transform.Rotation.X()))
		rotation = transform.rotateAroundAxis(rotation, mgl32.Vec3{0, 0, 1}, mgl32.DegToRad(transform.Rotation.Y()))
		rotation = transform.rotateAroundAxis(rotation, mgl32.Vec3{0, 1, 0}, mgl32.DegToRad(transform.Rotation.Z()))

		//@TODO ROTATIONS

		// scale
		scale := mgl32.Scale3D(transform.Scale.X(), transform.Scale.Y(), transform.Scale.Z())

		transform.prevPosition = transform.Position
		transform.prevRotation = transform.Rotation
		transform.prevScale = transform.Scale

		transform.matrix = translation.Mul4(rotation).Mul4(scale)
	}

	return transform.matrix
}

// rotateAroundAxis rotates a matrix around a given axis
func (transform *Transform) rotateAroundAxis(matrix mgl32.Mat4, axis mgl32.Vec3, angle float32) mgl32.Mat4 {
	q1 := mgl32.QuatRotate(angle, axis)
	transform.quat = transform.quat.Mul(q1)

	return matrix.Mul4(q1.Mat4())

}
