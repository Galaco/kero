package window

import (
	"github.com/galaco/tinygametools"
	"github.com/go-gl/glfw/v3.3/glfw"
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
	w, _ := win.Handle().Handle().GetFramebufferSize()
	return w
}

// Height
func (win *Window) Height() int {
	_, h := win.Handle().Handle().GetFramebufferSize()
	return h
}

// ShouldClose
func (win *Window) ShouldClose() bool {
	return win.Handle().Handle().ShouldClose()
}

// Close
func (win *Window) Close() {
	win.Handle().Handle().SetShouldClose(true)
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
	_, _, maxWidth, maxHeight := glfw.GetPrimaryMonitor().GetWorkarea()
	if width > maxWidth || height > maxHeight {
		width = maxWidth
		height = maxHeight
		win.Handle().SetSize(width, height)
	}

	currentWindow = &Window{
		window: win,
	}
	return currentWindow, nil
}
