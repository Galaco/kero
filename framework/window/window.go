package window

import (
	"github.com/galaco/tinygametools"
)

var currentWindow *Window

// CurrentWindow
func CurrentWindow() *Window {
	return currentWindow
}

// Window hold references to GUI Window
type Window struct {
	window *tinygametools.Window
}

// Width
func (win *Window) Width() int {
	w,_ :=win.Handle().Handle().GetFramebufferSize()
	return w
}

// Height
func (win *Window) Height() int {
	_,h :=win.Handle().Handle().GetFramebufferSize()
	return h
}

// SetActive marks this windows reference as current
func (win *Window) SetActive() {
	win.Handle().Handle().MakeContextCurrent()
}

// SetTitle changes this window's title
func (win *Window) SetTitle(title string) {
	win.Handle().Handle().SetTitle(title)
}

// SwapBuffers will swap this windows buffers (aka finish the current frame)
func (win *Window) SwapBuffers() {
	win.Handle().Handle().SwapBuffers()
}

// Handle returns the underlying window
func (win *Window) Handle() *tinygametools.Window {
	return win.window
}

// CreateWindow returns a new window of provided size and title.
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
