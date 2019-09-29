package input

import (
	"github.com/galaco/tinygametools"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type mouse struct {
	window 	   *tinygametools.Window
	mouse      *tinygametools.Mouse
	xOld, yOld float64
}

func (m *mouse) LockMousePosition() {
	m.window.Handle().SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
}

func (m *mouse) UnlockMousePosition() {
	m.window.Handle().SetInputMode(glfw.CursorMode, glfw.CursorNormal)
}

func (m *mouse) SetBoundWindow(win *tinygametools.Window) {
	m.window = win
}

func (m *mouse) RegisterExternalMousePositionCallback(callback func(x, y float64)) {
	m.mouse.AddMousePosCallback(func(window *glfw.Window, xpos float64, ypos float64) {
		if m.xOld == 0 {
			m.xOld = xpos
		}
		if m.yOld == 0 {
			m.yOld = ypos
		}

		callback(xpos-m.xOld, ypos-m.yOld)
		m.xOld = xpos
		m.yOld = ypos
	})
}
