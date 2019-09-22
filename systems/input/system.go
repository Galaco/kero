package input

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
)

type Input struct {
	systems.System
}

func (s *Input) Update(dt float64) {
	input.PollInput()
}

func (s *Input) frameworkKeyCallback(key input.Key, action input.KeyAction, mods input.ModifierKey) {
	switch action {
	case input.KeyPress:
		event.Singleton().Dispatch(messages.NewKeyPress(key))
	case input.KeyRelease:
		event.Singleton().Dispatch(messages.NewKeyRelease(key))
	}
}

func NewInput() *Input {
	i := &Input{}
	input.Keyboard().RegisterExternalKeyCallback(i.frameworkKeyCallback)
	return i
}
