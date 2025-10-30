package components

import "github.com/galaco/kero/framework/ecs"

// Animation component for skeletal animation.
// Entities with this component can play animated sequences.
type Animation struct {
	// Current animation state
	SequenceName string  // Name of the current animation sequence
	Frame        float32 // Current frame (can be fractional for interpolation)
	PlaybackRate float32 // Speed multiplier (1.0 = normal speed)

	// Animation control
	Looping      bool
	Playing      bool
	BlendWeight  float32 // For blending multiple animations (0-1)

	// Sequence info (populated by animation system)
	TotalFrames  int
	FrameRate    float32 // Frames per second
}

// IsComponent implements the ecs.Component marker interface
func (Animation) IsComponent() {}

// Register Animation component with the ECS system
func init() {
	ecs.RegisterComponent[Animation](ecs.ComponentTypeAnimation)
}

// NewAnimation creates an Animation component with default values
func NewAnimation(sequenceName string) Animation {
	return Animation{
		SequenceName: sequenceName,
		Frame:        0,
		PlaybackRate: 1.0,
		Looping:      true,
		Playing:      true,
		BlendWeight:  1.0,
		TotalFrames:  0,
		FrameRate:    30.0,
	}
}

// Play starts playing the animation
func (a *Animation) Play() {
	a.Playing = true
}

// Pause pauses the animation
func (a *Animation) Pause() {
	a.Playing = false
}

// Stop stops and resets the animation
func (a *Animation) Stop() {
	a.Playing = false
	a.Frame = 0
}

// IsFinished returns true if the animation has completed (for non-looping animations)
func (a *Animation) IsFinished() bool {
	if a.Looping {
		return false
	}
	return a.Frame >= float32(a.TotalFrames-1)
}

// GetNormalizedTime returns the current frame as a value between 0 and 1
func (a *Animation) GetNormalizedTime() float32 {
	if a.TotalFrames == 0 {
		return 0
	}
	return a.Frame / float32(a.TotalFrames)
}

// SetNormalizedTime sets the current frame from a normalized value (0-1)
func (a *Animation) SetNormalizedTime(normalizedTime float32) {
	a.Frame = normalizedTime * float32(a.TotalFrames)
}
