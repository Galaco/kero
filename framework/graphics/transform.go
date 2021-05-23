package graphics

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Transform Represents the transformation of an entity in
// a 3-dimensional space: position, rotation and scale.
// Note: Orientation is measured in degrees
type Transform struct {
	Translation mgl32.Vec3
	Orientation mgl32.Quat
	Scale       mgl32.Vec3
	Velocity    mgl32.Vec3

	prevPosition mgl32.Vec3
	prevRotation mgl32.Quat
	prevScale    mgl32.Vec3
	matrix       mgl32.Mat4
}

// TransformationMatrix computes object transformation matrix
func (transform *Transform) TransformationMatrix() mgl32.Mat4 {
	if !transform.Translation.ApproxEqual(transform.prevPosition) ||
		!transform.Orientation.ApproxEqual(transform.prevRotation) ||
		!transform.Scale.ApproxEqual(transform.prevScale) {

		// Scale of 0 is invalid
		if transform.Scale.X() == 0 ||
			transform.Scale.Y() == 0 ||
			transform.Scale.Z() == 0 {
			transform.Scale = mgl32.Vec3{1, 1, 1}
		}

		//Translate
		matrix := mgl32.Translate3D(transform.Translation.X(), transform.Translation.Y(), transform.Translation.Z())

		// rotate
		// IMPORTANT. Source engine has X and Z axis switched
		rotation := transform.Orientation.Mat4()

		//rotation := transform.rotateAroundAxis(mgl32.Ident4(), mgl32.Vec3{1, 0, 0}, mgl32.DegToRad(transform.Orientation.V[0]))
		//rotation = transform.rotateAroundAxis(rotation, mgl32.Vec3{0, 0, 1}, mgl32.DegToRad(transform.Orientation.V[1]))
		//rotation = transform.rotateAroundAxis(rotation, mgl32.Vec3{0, 1, 0}, mgl32.DegToRad(transform.Orientation.V[2]))


		//@TODO ROTATIONS

		// scale
		scale := mgl32.Scale3D(transform.Scale.X(), transform.Scale.Y(), transform.Scale.Z())

		transform.prevPosition = transform.Translation
		transform.prevRotation = transform.Orientation
		transform.prevScale = transform.Scale

		transform.matrix = matrix.Mul4(rotation).Mul4(scale)
	}

	return transform.matrix
}