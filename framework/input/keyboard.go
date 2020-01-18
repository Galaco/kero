package input

import (
	"github.com/galaco/tinygametools"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type keyboard struct {
	keyboard         *tinygametools.Keyboard
	currentKeyStates [1024]bool
}

func (kb *keyboard) RegisterExternalKeyCallback(callback func(key Key, action KeyAction, mods ModifierKey)) {
	kb.keyboard.AddKeyCallback(func(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		callback(Key(key), KeyAction(action), ModifierKey(mods))
	})
}

func (kb *keyboard) keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		kb.currentKeyStates[key] = true
	case glfw.Repeat:
		return
	case glfw.Release:
		kb.currentKeyStates[key] = false
	}
}

func (kb *keyboard) KeyStates() [1024]bool {
	return kb.currentKeyStates
}

func (kb *keyboard) IsKeyPressed(key Key) bool {
	return kb.currentKeyStates[key]
}
