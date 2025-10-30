package graphics

import (
	"math"

	"github.com/galaco/kero/framework/console"
	"github.com/go-gl/mathgl/mgl32"
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

	// Movement properties for smooth acceleration/deceleration
	velocity     mgl32.Vec3 // Current movement velocity (units/second)
	acceleration float32    // How quickly camera accelerates to max speed
	deceleration float32    // How quickly camera decelerates when no input
	maxSpeed     float32    // Maximum movement speed
}

// Fov
func (camera *Camera) Fov() float32 {
	return camera.fov
}

// AspectRatio
func (camera *Camera) AspectRatio() float32 {
	return camera.aspectRatio
}

// Transform Returns this entity's transform component
func (camera *Camera) Transform() *Transform {
	return &camera.transform
}

// SetMovementInput sets the desired movement direction based on input.
// The direction vector should be normalized or zero.
// This method applies acceleration toward the desired direction.
func (camera *Camera) SetMovementInput(forward, right float32, dt float64) {
	if dt <= 0 {
		return
	}

	// Build desired velocity from input direction
	desiredDirection := camera.direction.Mul(forward).Add(camera.right.Mul(right))

	// Normalize if we have input
	if desiredDirection.Len() > 0.01 {
		desiredDirection = desiredDirection.Normalize()
		desiredVelocity := desiredDirection.Mul(camera.maxSpeed)

		// Accelerate toward desired velocity
		velocityDiff := desiredVelocity.Sub(camera.velocity)
		accelerationAmount := camera.acceleration * float32(dt)

		// If the velocity difference is smaller than what we'd add, just set it
		if velocityDiff.Len() < accelerationAmount {
			camera.velocity = desiredVelocity
		} else {
			camera.velocity = camera.velocity.Add(velocityDiff.Normalize().Mul(accelerationAmount))
		}
	} else {
		// No input - apply deceleration
		decelAmount := camera.deceleration * float32(dt)
		currentSpeed := camera.velocity.Len()

		if currentSpeed < decelAmount {
			// Close enough to zero, just stop
			camera.velocity = mgl32.Vec3{0, 0, 0}
		} else {
			// Decelerate
			camera.velocity = camera.velocity.Mul(1.0 - (decelAmount / currentSpeed))
		}
	}

	// Clamp velocity to max speed
	if camera.velocity.Len() > camera.maxSpeed {
		camera.velocity = camera.velocity.Normalize().Mul(camera.maxSpeed)
	}
}

// Forwards
func (camera *Camera) Forwards(dt float64) {
	camera.Transform().Translation = camera.Transform().Translation.Add(camera.direction.Mul(float32(cameraSpeed * dt)))
}

// Backwards
func (camera *Camera) Backwards(dt float64) {
	camera.Transform().Translation = camera.Transform().Translation.Sub(camera.direction.Mul(float32(cameraSpeed * dt)))
}

// Left
func (camera *Camera) Left(dt float64) {
	camera.Transform().Translation = camera.Transform().Translation.Sub(camera.right.Mul(float32(cameraSpeed * dt)))
}

// Right
func (camera *Camera) Right(dt float64) {
	camera.Transform().Translation = camera.Transform().Translation.Add(camera.right.Mul(float32(cameraSpeed * dt)))
}

// Rotate applies mouse delta to camera rotation.
// Uses m_sensitivity ConVar as a multiplier for runtime adjustment.
func (camera *Camera) Rotate(x, y, z float32) {
	// Get sensitivity multiplier from ConVar (default 1.0)
	sens := console.GetConvarFloat("m_sensitivity")
	if sens <= 0 {
		sens = 1.0
	}
	effectiveSens := sensitivity * sens

	camera.Transform().Orientation.V[0] = camera.Transform().Orientation.V[0] + (x * effectiveSens)
	camera.Transform().Orientation.V[1] = camera.Transform().Orientation.V[1] + (y * effectiveSens)
	camera.Transform().Orientation.V[2] = camera.Transform().Orientation.V[2] + (z * effectiveSens)

	// Lock vertical rotation to prevent over-rotation
	if camera.Transform().Orientation.V[2] > maxVerticalRotation {
		camera.Transform().Orientation.V[2] = maxVerticalRotation
	}
	if camera.Transform().Orientation.V[2] < minVerticalRotation {
		camera.Transform().Orientation.V[2] = minVerticalRotation
	}
}

// Update updates the camera position
func (camera *Camera) Update(dt float64) {
	camera.updateVectors()

	// Update movement properties from console variables if available
	if accel := console.GetConvarFloat("cam_acceleration"); accel > 0 {
		camera.acceleration = accel
	}
	if decel := console.GetConvarFloat("cam_deceleration"); decel > 0 {
		camera.deceleration = decel
	}
	if maxSpeed := console.GetConvarFloat("cam_maxspeed"); maxSpeed > 0 {
		camera.maxSpeed = maxSpeed
	}

	// Integrate velocity into position
	if dt > 0 {
		camera.Transform().Translation = camera.Transform().Translation.Add(camera.velocity.Mul(float32(dt)))
	}
}

// updateVectors Updates the camera directional properties with any changes
func (camera *Camera) updateVectors() {
	rot := camera.Transform().Orientation

	// Calculate the new Front vector
	camera.direction = mgl32.Vec3{
		float32(math.Cos(float64(rot.V[2])) * math.Sin(float64(rot.V[0]))),
		float32(math.Cos(float64(rot.V[2])) * math.Cos(float64(rot.V[0]))),
		float32(math.Sin(float64(rot.V[2]))),
	}
	// Also re-calculate the right and up vector
	camera.right = mgl32.Vec3{
		float32(math.Sin(float64(rot.V[0]) - math.Pi/2)),
		float32(math.Cos(float64(rot.V[0]) - math.Pi/2)),
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
		camera.Transform().Translation,
		camera.Transform().Translation.Add(camera.direction),
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
		fov:          fov,
		aspectRatio:  aspectRatio,
		up:           mgl32.Vec3{0, 1, 0},
		worldUp:      mgl32.Vec3{0, 1, 0},
		direction:    mgl32.Vec3{0, 0, -1},
		velocity:     mgl32.Vec3{0, 0, 0},
		acceleration: 2000.0, // Units per second squared
		deceleration: 8.0,    // Multiplier for deceleration
		maxSpeed:     640.0,  // Units per second
	}
}
