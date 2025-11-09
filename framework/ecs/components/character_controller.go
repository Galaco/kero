package components

import "github.com/galaco/kero/framework/ecs"

// CharacterController provides character-specific physics collision.
// This component stores the Bullet CharacterController handle for kinematic character movement.
type CharacterController struct {
	// Capsule dimensions
	Height     float32 // Total capsule height (units)
	Radius     float32 // Capsule radius (units)
	StepHeight float32 // Maximum step-up height for stairs (units)

	// Phase 3: Direct CharacterController reference
	// Using interface{} to avoid physics package dependency in components.
	// Physics system casts this to *collision.CharacterController when needed.
	ControllerHandle interface{} // Bullet CharacterController handle (nil = not initialized)
}

// IsComponent implements the ecs.Component marker interface
func (CharacterController) IsComponent() {}

// Register CharacterController component with the ECS system
func init() {
	ecs.RegisterComponent[CharacterController](ecs.ComponentTypeCharacterController)
}

// ============================================================================
// CharacterController Handle Management
// ============================================================================

// SetController stores the Bullet CharacterController handle for this entity.
// This should be called by the physics system after creating the controller.
func (cc *CharacterController) SetController(controller interface{}) {
	cc.ControllerHandle = controller
}

// GetController retrieves the Bullet CharacterController handle.
// Returns nil if no controller is attached.
// Physics/movement systems should cast this to *collision.CharacterController.
func (cc *CharacterController) GetController() interface{} {
	return cc.ControllerHandle
}

// HasController checks if a CharacterController is attached to this component.
// Returns true if the handle is non-nil.
func (cc *CharacterController) HasController() bool {
	return cc.ControllerHandle != nil
}

// NewCharacterController creates a CharacterController component with default player dimensions
func NewCharacterController(height, radius, stepHeight float32) CharacterController {
	return CharacterController{
		Height:           height,
		Radius:           radius,
		StepHeight:       stepHeight,
		ControllerHandle: nil,
	}
}
