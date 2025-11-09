package components

import (
	"github.com/galaco/kero/framework/ecs"
	"github.com/go-gl/mathgl/mgl32"
)

// CharacterMovement handles character-specific movement physics.
// This component stores movement state and parameters for character controllers.
type CharacterMovement struct {
	// Current motion state
	Velocity mgl32.Vec3 // Current velocity vector (units/second)

	// Movement parameters
	GroundSpeed  float32 // Maximum ground movement speed (units/second)
	AirSpeed     float32 // Air control speed (units/second)
	JumpVelocity float32 // Initial velocity when jumping (units/second)
	Friction     float32 // Ground friction coefficient
	Acceleration float32 // Acceleration rate (units/second²)

	// State flags
	OnGround bool // True if character is on ground

	// Collision state
	LastWallNormal mgl32.Vec3 // Normal of last wall hit (for slide prevention)
}

// IsComponent implements the ecs.Component marker interface
func (CharacterMovement) IsComponent() {}

// Register CharacterMovement component with the ECS system
func init() {
	ecs.RegisterComponent[CharacterMovement](ecs.ComponentTypeCharacterMovement)
}

// NewCharacterMovement creates a CharacterMovement component with default values
func NewCharacterMovement() CharacterMovement {
	return CharacterMovement{
		Velocity:       mgl32.Vec3{0, 0, 0},
		GroundSpeed:    320.0,
		AirSpeed:       30.0,
		JumpVelocity:   270.0,
		Friction:       4.0,
		Acceleration:   10.0,
		OnGround:       false,
		LastWallNormal: mgl32.Vec3{0, 0, 0},
	}
}

// IsMoving returns true if the character has significant velocity
func (cm *CharacterMovement) IsMoving() bool {
	const threshold = 0.01
	return cm.Velocity.Len() > threshold
}

// ClearVelocity sets velocity to zero
func (cm *CharacterMovement) ClearVelocity() {
	cm.Velocity = mgl32.Vec3{0, 0, 0}
}
