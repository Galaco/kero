package components

import (
	"github.com/galaco/kero/framework/ecs"
	"github.com/go-gl/mathgl/mgl32"
)

// CameraController attaches a camera to an entity and controls its behavior.
// This component allows an entity to control camera position and orientation.
type CameraController struct {
	// Camera positioning
	EyeOffset mgl32.Vec3 // Offset from entity origin to camera position

	// Camera settings
	FOV         float32 // Field of view in degrees
	AspectRatio float32 // Aspect ratio (width/height)
	Near        float32 // Near clipping plane
	Far         float32 // Far clipping plane

	// State
	Active bool // Is this the active rendering camera?

	// Phase 3: Direct camera reference
	// Using interface{} to avoid graphics package dependency in components.
	// Camera system casts this to *graphics.Camera when needed.
	CameraHandle interface{} // graphics.Camera handle (nil = not initialized)
}

// IsComponent implements the ecs.Component marker interface
func (CameraController) IsComponent() {}

// Register CameraController component with the ECS system
func init() {
	ecs.RegisterComponent[CameraController](ecs.ComponentTypeCameraController)
}

// ============================================================================
// Camera Handle Management
// ============================================================================

// SetCamera stores the graphics camera handle for this entity.
// This should be called during entity creation or camera attachment.
func (cc *CameraController) SetCamera(camera interface{}) {
	cc.CameraHandle = camera
}

// GetCamera retrieves the graphics camera handle.
// Returns nil if no camera is attached.
// Camera system should cast this to *graphics.Camera.
func (cc *CameraController) GetCamera() interface{} {
	return cc.CameraHandle
}

// HasCamera checks if a camera is attached to this component.
// Returns true if the handle is non-nil.
func (cc *CameraController) HasCamera() bool {
	return cc.CameraHandle != nil
}

// NewCameraController creates a CameraController component with default values
func NewCameraController(eyeOffset mgl32.Vec3, fov float32, active bool) CameraController {
	return CameraController{
		EyeOffset:    eyeOffset,
		FOV:          fov,
		AspectRatio:  16.0 / 9.0, // Default aspect ratio
		Near:         0.1,
		Far:          10000.0,
		Active:       active,
		CameraHandle: nil,
	}
}
