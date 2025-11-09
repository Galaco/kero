package components

import "github.com/galaco/kero/framework/ecs"

// PlayerController marks an entity as player-controlled and stores rotation state.
// This component handles player-specific state like yaw/pitch rotation and player index.
type PlayerController struct {
	// Player identification
	PlayerIndex int // For multiplayer support (0 = first player)

	// Rotation state (in radians)
	Yaw   float32 // Horizontal rotation (unbounded)
	Pitch float32 // Vertical rotation (clamped to [-π/2, π/2])

	// Input sensitivity
	Sensitivity float32 // Mouse sensitivity multiplier
}

// IsComponent implements the ecs.Component marker interface
func (PlayerController) IsComponent() {}

// Register PlayerController component with the ECS system
func init() {
	ecs.RegisterComponent[PlayerController](ecs.ComponentTypePlayerController)
}

// NewPlayerController creates a PlayerController component with default values
func NewPlayerController(playerIndex int, yaw float32, sensitivity float32) PlayerController {
	return PlayerController{
		PlayerIndex: playerIndex,
		Yaw:         yaw,
		Pitch:       0,
		Sensitivity: sensitivity,
	}
}
