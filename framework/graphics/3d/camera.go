package graphics3d

import (
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

const cameraSpeed = float64(320) * 2
const sensitivity = float32(0.03)

var minVerticalRotation = mgl32.DegToRad(90)
var maxVerticalRotation = mgl32.DegToRad(270)

// Camera
type Camera struct {
	transform Transform

	fov         float32
	aspectRatio float32
	up          mgl32.Vec3
	right       mgl32.Vec3
	direction   mgl32.Vec3
	worldUp     mgl32.Vec3
}

// Transform Returns this entity's transform component
func (camera *Camera) Transform() *Transform {
	return &camera.transform
}

// Forwards
func (camera *Camera) Forwards(dt float64) {
	camera.Transform().Position = camera.Transform().Position.Add(camera.direction.Mul(float32(cameraSpeed * dt)))
}

// Backwards
func (camera *Camera) Backwards(dt float64) {
	camera.Transform().Position = camera.Transform().Position.Sub(camera.direction.Mul(float32(cameraSpeed * dt)))
}

// Left
func (camera *Camera) Left(dt float64) {
	camera.Transform().Position = camera.Transform().Position.Sub(camera.right.Mul(float32(cameraSpeed * dt)))
}

// Right
func (camera *Camera) Right(dt float64) {
	camera.Transform().Position = camera.Transform().Position.Add(camera.right.Mul(float32(cameraSpeed * dt)))
}

// Rotate
func (camera *Camera) Rotate(x, y, z float32) {
	camera.Transform().Rotation[0] = camera.Transform().Rotation[0] + (x * sensitivity)
	camera.Transform().Rotation[1] = camera.Transform().Rotation[1] + (y * sensitivity)
	camera.Transform().Rotation[2] = camera.Transform().Rotation[2] + (z * sensitivity)

	// Lock vertical rotation
	if camera.Transform().Rotation[2] > maxVerticalRotation {
		camera.Transform().Rotation[2] = maxVerticalRotation
	}
	if camera.Transform().Rotation[2] < minVerticalRotation {
		camera.Transform().Rotation[2] = minVerticalRotation
	}
}

// Update updates the camera position
func (camera *Camera) Update(dt float64) {
	camera.updateVectors()
}

// updateVectors Updates the camera directional properties with any changes
func (camera *Camera) updateVectors() {
	rot := camera.Transform().Rotation

	// Calculate the new Front vector
	camera.direction = mgl32.Vec3{
		float32(math.Cos(float64(rot[2])) * math.Sin(float64(rot[0]))),
		float32(math.Cos(float64(rot[2])) * math.Cos(float64(rot[0]))),
		float32(math.Sin(float64(rot[2]))),
	}
	// Also re-calculate the right and up vector
	camera.right = mgl32.Vec3{
		float32(math.Sin(float64(rot[0]) - math.Pi/2)),
		float32(math.Cos(float64(rot[0]) - math.Pi/2)),
		0,
	}
	camera.up = camera.right.Cross(camera.direction)
}

// ModelMatrix returns identity matrix (camera model is our position!)
func (camera *Camera) ModelMatrix() mgl32.Mat4 {
	return mgl32.Ident4()
}

// ViewMatrix calculates the cameras View matrix
func (camera *Camera) ViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(
		camera.Transform().Position,
		camera.Transform().Position.Add(camera.direction),
		camera.up)
}

// ProjectionMatrix calculates projection matrix.
// This is unlikely to change throughout program lifetime, but could do
func (camera *Camera) ProjectionMatrix() mgl32.Mat4 {
	return mgl32.Perspective(camera.fov, camera.aspectRatio, 0.2, 16384)
}

// NewCamera returns a new camera
// fov should be provided in radians
func NewCamera(fov float32, aspectRatio float32) *Camera {
	return &Camera{
		fov:         fov,
		aspectRatio: aspectRatio,
		up:          mgl32.Vec3{0, 1, 0},
		worldUp:     mgl32.Vec3{0, 1, 0},
		direction:   mgl32.Vec3{0, 0, -1},
	}
}
