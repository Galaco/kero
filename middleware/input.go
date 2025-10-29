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
		s.eventBus.Dispatch(messages.TypeKeyPress, key)
		if key == input.KeyEscape {
			s.shouldLockMouse = !s.shouldLockMouse
			if s.shouldLockMouse {
				input.Mouse().LockMousePosition()
			} else {
				input.Mouse().UnlockMousePosition()
			}
		}
	case input.KeyRelease:
		s.eventBus.Dispatch(messages.TypeKeyRelease, key)
	}
}

func (s *Input) frameworkMousePositionCallback(x, y float64) {
	s.eventBus.Dispatch(messages.TypeMouseMove, mgl32.Vec2{float32(x), float32(y)})
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
