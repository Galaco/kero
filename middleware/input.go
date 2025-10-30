package middleware

import (
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/messages"
	"github.com/go-gl/mathgl/mgl32"
)

var inputMiddleware *Input

type Input struct {
	eventBus        *event.Dispatcher
	shouldLockMouse bool
}

func (s *Input) Poll() {
	input.PollInput()
}

func (s *Input) frameworkKeyCallback(key input.Key, action input.KeyAction, mods input.ModifierKey) {
	switch action {
	case input.KeyPress:
		// Use typed event dispatch (Phase 3)
		event.DispatchTyped(s.eventBus, messages.KeyPressEvent{Key: key})
		if key == input.KeyEscape {
			s.shouldLockMouse = !s.shouldLockMouse
			if s.shouldLockMouse {
				input.Mouse().LockMousePosition()
			} else {
				input.Mouse().UnlockMousePosition()
			}
		}
	case input.KeyRelease:
		// Use typed event dispatch (Phase 3)
		event.DispatchTyped(s.eventBus, messages.KeyReleaseEvent{Key: key})
	}
}

func (s *Input) frameworkMousePositionCallback(x, y float64) {
	// Use typed event dispatch (Phase 3)
	// Note: x, y are already DELTAS from mouse.go (not absolute position)
	event.DispatchTyped(s.eventBus, messages.MouseMoveEvent{
		Position: mgl32.Vec2{0, 0},                     // Not used for camera rotation
		Delta:    mgl32.Vec2{float32(x), float32(y)}, // Mouse movement delta
	})
}

// NewInput creates a new Input middleware with explicit dependencies.
// This is the preferred way to create Input instances.
func NewInput(eventBus *event.Dispatcher) *Input {
	inputMiddleware := &Input{
		eventBus: eventBus,
	}
	input.Keyboard().RegisterExternalKeyCallback(inputMiddleware.frameworkKeyCallback)
	input.Mouse().RegisterExternalMousePositionCallback(inputMiddleware.frameworkMousePositionCallback)
	return inputMiddleware
}

// Deprecated: Use NewInput with explicit EventBus dependency instead
func InitializeInput() *Input {
	if inputMiddleware == nil {
		inputMiddleware = &Input{}
		inputMiddleware.eventBus = event.Get()
		input.Keyboard().RegisterExternalKeyCallback(inputMiddleware.frameworkKeyCallback)
		input.Mouse().RegisterExternalMousePositionCallback(inputMiddleware.frameworkMousePositionCallback)
	}
	return inputMiddleware
}

// Deprecated: Use Engine.Input() instead
func InputMiddleware() *Input {
	return inputMiddleware
}

// EventBus returns the event bus for external access
func (s *Input) EventBus() *event.Dispatcher {
	return s.eventBus
}
