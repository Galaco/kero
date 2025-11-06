package components

import (
	"math"

	"github.com/galaco/kero/framework/ecs"
	"github.com/go-gl/mathgl/mgl32"
)

// Transform component for spatial positioning in 3D space.
// Every entity with a position in the world should have this component.
type Transform struct {
	Position    mgl32.Vec3
	Orientation mgl32.Quat
	Scale       mgl32.Vec3
}

// IsComponent implements the ecs.Component marker interface
func (Transform) IsComponent() {}

// Register Transform component with the ECS system
func init() {
	ecs.RegisterComponent[Transform](ecs.ComponentTypeTransform)
}

// Forward returns the forward direction vector (local Z axis)
func (t *Transform) Forward() mgl32.Vec3 {
	return t.Orientation.Rotate(mgl32.Vec3{0, 0, -1})
}

// Right returns the right direction vector (local X axis)
func (t *Transform) Right() mgl32.Vec3 {
	return t.Orientation.Rotate(mgl32.Vec3{1, 0, 0})
}

// Up returns the up direction vector (local Y axis)
func (t *Transform) Up() mgl32.Vec3 {
	return t.Orientation.Rotate(mgl32.Vec3{0, 1, 0})
}

// SetEulerAngles sets the orientation from Euler angles (in radians)
func (t *Transform) SetEulerAngles(pitch, yaw, roll float32) {
	t.Orientation = mgl32.AnglesToQuat(pitch, yaw, roll, mgl32.XYZ)
}

// GetEulerAngles returns the orientation as Euler angles (in radians)
// Note: This is an approximation and may have gimbal lock issues
func (t *Transform) GetEulerAngles() (pitch, yaw, roll float32) {
	// Convert quaternion to Euler angles
	// Based on: https://en.wikipedia.org/wiki/Conversion_between_quaternions_and_Euler_angles
	q := t.Orientation

	// Roll (x-axis rotation)
	sinr_cosp := 2.0 * (q.W*q.V[0] + q.V[1]*q.V[2])
	cosr_cosp := 1.0 - 2.0*(q.V[0]*q.V[0]+q.V[1]*q.V[1])
	roll = float32(math.Atan2(float64(sinr_cosp), float64(cosr_cosp)))

	// Pitch (y-axis rotation)
	sinp := 2.0 * (q.W*q.V[1] - q.V[2]*q.V[0])
	if mgl32.Abs(sinp) >= 1 {
		pitch = float32(math.Copysign(math.Pi/2, float64(sinp))) // Use 90 degrees if out of range
	} else {
		pitch = float32(math.Asin(float64(sinp)))
	}

	// Yaw (z-axis rotation)
	siny_cosp := 2.0 * (q.W*q.V[2] + q.V[0]*q.V[1])
	cosy_cosp := 1.0 - 2.0*(q.V[1]*q.V[1]+q.V[2]*q.V[2])
	yaw = float32(math.Atan2(float64(siny_cosp), float64(cosy_cosp)))

	return pitch, yaw, roll
}

// TransformPoint transforms a point from local space to world space
func (t *Transform) TransformPoint(localPoint mgl32.Vec3) mgl32.Vec3 {
	// Scale, rotate, then translate
	scaled := localPoint.Mul(t.Scale.X())  // Uniform scale for simplicity
	rotated := t.Orientation.Rotate(scaled)
	return rotated.Add(t.Position)
}

// InverseTransformPoint transforms a point from world space to local space
func (t *Transform) InverseTransformPoint(worldPoint mgl32.Vec3) mgl32.Vec3 {
	// Inverse: untranslate, unrotate, unscale
	translated := worldPoint.Sub(t.Position)
	unrotated := t.Orientation.Conjugate().Rotate(translated)
	return unrotated.Mul(1.0 / t.Scale.X())  // Uniform scale
}

// NewTransform creates a Transform with default values
func NewTransform() Transform {
	return Transform{
		Position:    mgl32.Vec3{0, 0, 0},
		Orientation: mgl32.QuatIdent(),
		Scale:       mgl32.Vec3{1, 1, 1},
	}
}

// NewTransformAt creates a Transform at a specific position
func NewTransformAt(position mgl32.Vec3) Transform {
	return Transform{
		Position:    position,
		Orientation: mgl32.QuatIdent(),
		Scale:       mgl32.Vec3{1, 1, 1},
	}
}
