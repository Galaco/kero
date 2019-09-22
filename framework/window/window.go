package window

import (
	"github.com/galaco/tinygametools"
)

var currentWindow *Window

// CurrentWindow
func CurrentWindow() *Window {
	return currentWindow
}

type Window struct {
	window *tinygametools.Window
}

func (win *Window) SetActive() {
	win.Handle().Handle().MakeContextCurrent()
}

func (win *Window) SwapBuffers() {
	win.Handle().Handle().SwapBuffers()
}

func (win *Window) Handle() *tinygametools.Window {
	return win.window
}

func CreateWindow(width, height int, title string) (*Window, error) {
	win, err := tinygametools.NewWindow(width, height, title)
	if err != nil {
		return nil, err
	}

	w := &Window{
		window: win,
	}
	currentWindow = w
	return w, nil
}
