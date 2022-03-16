package input

import (
	"github.com/galaco/tinygametools"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type ContextBindable interface {
	Handle() *tinygametools.Window
}

var internalKeyboard *keyboard
var internalMouse *mouse

func Keyboard() *keyboard {
	return internalKeyboard
}

func Mouse() *mouse {
	return internalMouse
}

func PollInput() {
	glfw.PollEvents()
}

func SetBoundWindow(win ContextBindable) {
	Keyboard().keyboard.AddKeyCallback(Keyboard().keyCallback)
	Keyboard().keyboard.RegisterCallbacks(win.Handle())
	Mouse().mouse.RegisterCallbacks(win.Handle())
	Mouse().SetBoundWindow(win.Handle())
}

func init() {
	internalKeyboard = &keyboard{
		keyboard: tinygametools.NewKeyboard(),
	}
	internalMouse = &mouse{
		mouse: tinygametools.NewMouse(),
	}
}
