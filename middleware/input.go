package middleware

import (
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/messages"
)

var inputMiddleware *Input

type Input struct {
	event.Dispatcher

	shouldLockMouse bool
}

func (s *Input) Poll() {
	input.PollInput()
}

func (s *Input) frameworkKeyCallback(key input.Key, action input.KeyAction, mods input.ModifierKey) {
	switch action {
	case input.KeyPress:
		s.Dispatch(messages.NewKeyPress(key))
		if key == input.KeyEscape {
			s.shouldLockMouse = !s.shouldLockMouse
			if s.shouldLockMouse {
				input.Mouse().LockMousePosition()
			} else {
				input.Mouse().UnlockMousePosition()
			}
		}
	case input.KeyRelease:
		s.Dispatch(messages.NewKeyRelease(key))
	}
}

func (s *Input) frameworkMousePositionCallback(x, y float64) {
	s.Dispatch(messages.NewMouseMove(x, y))
}

func InitializeInput() *Input {
	inputMiddleware = &Input{}
	inputMiddleware.Dispatcher.Initialize()
	input.Keyboard().RegisterExternalKeyCallback(inputMiddleware.frameworkKeyCallback)
	input.Mouse().RegisterExternalMousePositionCallback(inputMiddleware.frameworkMousePositionCallback)
	return inputMiddleware
}

func InputMiddleware() *Input {
	return inputMiddleware
}