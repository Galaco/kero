package context

import (
	"github.com/AllenDang/cimgui-go/imgui"
	gl3_impl "github.com/AllenDang/cimgui-go/impl/opengl3"
	"github.com/galaco/kero/framework/input"
	"github.com/go-gl/glfw/v3.3/glfw"
	"math"
)

type imguiGlfw3 struct {
	window           *glfw.Window
	time             float64
	mouseJustPressed [3]bool
}

func imguiGlfw3Init(window *glfw.Window) *imguiGlfw3 {
	impl := &imguiGlfw3{
		window: window,
	}

	// Initialize OpenGL3 backend only
	gl3_impl.InitV("#version 150")

	// Set up ImGui key mappings
	io := imgui.CurrentIO()
	io.SetBackendFlags(io.BackendFlags() | imgui.BackendFlagsHasMouseCursors)

	// Map keyboard keys - using AddKeyEvent for the new API
	// Note: We'll handle actual key events in callbacks, this is just for ImGui internal mapping

	// Install callbacks on the actual GLFW window
	impl.installCallbacks()

	return impl
}

func (impl *imguiGlfw3) NewFrame() {
	io := imgui.CurrentIO()

	// Setup display size (every frame to accommodate for window resizing)
	windowWidth, windowHeight := impl.window.GetSize()
	framebufferWidth, framebufferHeight := impl.window.GetFramebufferSize()

	io.SetDisplaySize(imgui.Vec2{X: float32(windowWidth), Y: float32(windowHeight)})

	// Set framebuffer scale for high-DPI displays (Retina, etc.)
	if windowWidth > 0 && windowHeight > 0 {
		io.SetDisplayFramebufferScale(imgui.Vec2{
			X: float32(framebufferWidth) / float32(windowWidth),
			Y: float32(framebufferHeight) / float32(windowHeight),
		})
	}

	// Setup time step
	currentTime := glfw.GetTime()
	if impl.time > 0 {
		io.SetDeltaTime(float32(currentTime - impl.time))
	} else {
		io.SetDeltaTime(1.0 / 60.0) // Assume 60 FPS on first frame
	}
	impl.time = currentTime

	// Setup inputs
	if impl.window.GetAttrib(glfw.Focused) != 0 {
		x, y := impl.window.GetCursorPos()
		io.AddMousePosEvent(float32(x), float32(y))
	} else {
		io.AddMousePosEvent(-math.MaxFloat32, -math.MaxFloat32)
	}

	// Update mouse buttons
	for i := 0; i < len(impl.mouseJustPressed); i++ {
		down := impl.mouseJustPressed[i] || (impl.window.GetMouseButton(buttonIDByIndex[i]) == glfw.Press)
		io.AddMouseButtonEvent(int32(i), down)
		impl.mouseJustPressed[i] = false
	}

	// OpenGL3 backend prepares rendering state
	gl3_impl.NewFrame()

	// Start ImGui frame
	imgui.NewFrame()
}

func (impl *imguiGlfw3) Render() {
	// Render ImGui and get the draw data
	imgui.Render()
	drawData := imgui.CurrentDrawData()
	gl3_impl.RenderDrawData(drawData)
}

func (impl *imguiGlfw3) Shutdown() {
	gl3_impl.Shutdown()
}

func (impl *imguiGlfw3) installCallbacks() {
	impl.window.SetMouseButtonCallback(impl.mouseButtonChange)
	impl.window.SetScrollCallback(impl.mouseScrollChange)
	input.Keyboard().RegisterExternalKeyCallback(impl.keyChange)
	impl.window.SetCharCallback(impl.charChange)
}

var buttonIndexByID = map[glfw.MouseButton]int{
	glfw.MouseButton1: 0,
	glfw.MouseButton2: 1,
	glfw.MouseButton3: 2,
}

var buttonIDByIndex = map[int]glfw.MouseButton{
	0: glfw.MouseButton1,
	1: glfw.MouseButton2,
	2: glfw.MouseButton3,
}

func (impl *imguiGlfw3) mouseButtonChange(window *glfw.Window, rawButton glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	buttonIndex, known := buttonIndexByID[rawButton]

	if known && (action == glfw.Press) {
		impl.mouseJustPressed[buttonIndex] = true
	}
}

func (impl *imguiGlfw3) mouseScrollChange(window *glfw.Window, x, y float64) {
	io := imgui.CurrentIO()
	io.AddMouseWheelEvent(float32(x), float32(y))
}

func (impl *imguiGlfw3) keyChange(key input.Key, action input.KeyAction, mods input.ModifierKey) {
	io := imgui.CurrentIO()

	// Map kero input.Key to ImGui key
	imguiKey := mapKeyToImGuiKey(key)

	// Handle key press/release
	if action == input.KeyPress {
		io.AddKeyEvent(imguiKey, true)
	}
	if action == input.KeyRelease {
		io.AddKeyEvent(imguiKey, false)
	}

	// Update modifier keys
	io.AddKeyEvent(imgui.KeyLeftCtrl, impl.window.GetKey(glfw.KeyLeftControl) == glfw.Press)
	io.AddKeyEvent(imgui.KeyRightCtrl, impl.window.GetKey(glfw.KeyRightControl) == glfw.Press)
	io.AddKeyEvent(imgui.KeyLeftShift, impl.window.GetKey(glfw.KeyLeftShift) == glfw.Press)
	io.AddKeyEvent(imgui.KeyRightShift, impl.window.GetKey(glfw.KeyRightShift) == glfw.Press)
	io.AddKeyEvent(imgui.KeyLeftAlt, impl.window.GetKey(glfw.KeyLeftAlt) == glfw.Press)
	io.AddKeyEvent(imgui.KeyRightAlt, impl.window.GetKey(glfw.KeyRightAlt) == glfw.Press)
	io.AddKeyEvent(imgui.KeyLeftSuper, impl.window.GetKey(glfw.KeyLeftSuper) == glfw.Press)
	io.AddKeyEvent(imgui.KeyRightSuper, impl.window.GetKey(glfw.KeyRightSuper) == glfw.Press)
}

func (impl *imguiGlfw3) charChange(window *glfw.Window, char rune) {
	io := imgui.CurrentIO()
	io.AddInputCharacter(uint32(char))
}

// mapKeyToImGuiKey maps GLFW/kero keys to ImGui keys
func mapKeyToImGuiKey(key input.Key) imgui.Key {
	// Map common keys - kero's input.Key values should match GLFW key codes
	switch glfw.Key(key) {
	case glfw.KeyTab:
		return imgui.KeyTab
	case glfw.KeyLeft:
		return imgui.KeyLeftArrow
	case glfw.KeyRight:
		return imgui.KeyRightArrow
	case glfw.KeyUp:
		return imgui.KeyUpArrow
	case glfw.KeyDown:
		return imgui.KeyDownArrow
	case glfw.KeyPageUp:
		return imgui.KeyPageUp
	case glfw.KeyPageDown:
		return imgui.KeyPageDown
	case glfw.KeyHome:
		return imgui.KeyHome
	case glfw.KeyEnd:
		return imgui.KeyEnd
	case glfw.KeyInsert:
		return imgui.KeyInsert
	case glfw.KeyDelete:
		return imgui.KeyDelete
	case glfw.KeyBackspace:
		return imgui.KeyBackspace
	case glfw.KeySpace:
		return imgui.KeySpace
	case glfw.KeyEnter:
		return imgui.KeyEnter
	case glfw.KeyEscape:
		return imgui.KeyEscape
	case glfw.KeyA:
		return imgui.KeyA
	case glfw.KeyC:
		return imgui.KeyC
	case glfw.KeyV:
		return imgui.KeyV
	case glfw.KeyX:
		return imgui.KeyX
	case glfw.KeyY:
		return imgui.KeyY
	case glfw.KeyZ:
		return imgui.KeyZ
	default:
		return imgui.KeyNone
	}
}
