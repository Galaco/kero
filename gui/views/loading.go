package views

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/messages"
	"math"
)

type Loading struct {
	state         int
	statusText    *gui.Text
	cancelButton  *gui.Button
	animationTime float32
	onCancel      func()
}

func NewLoading(onCancel func()) *Loading {
	return &Loading{
		state:         messages.LoadingProgressStateStarted,
		statusText:    gui.NewText("Initializing..."),
		cancelButton:  gui.NewButton("cancel_loading", "Cancel Loading", onCancel),
		animationTime: 0,
		onCancel:      onCancel,
	}
}

func (view *Loading) UpdateProgress(state int) {
	view.state = state

	// Update status message based on state
	statusMessage := getLoadingStateMessage(state)
	view.statusText.SetText(statusMessage)
}

func (view *Loading) Update(dt float32) {
	// Update animation time for progress bar
	view.animationTime += dt
}

func (view *Loading) Render() {
	// Center the loading panel on screen
	viewport := imgui.MainViewport()
	displaySize := viewport.WorkSize()
	windowWidth := float32(400)
	windowHeight := float32(200)

	imgui.SetNextWindowPosV(
		imgui.Vec2{X: (displaySize.X - windowWidth) / 2, Y: (displaySize.Y - windowHeight) / 2},
		imgui.CondAlways,
		imgui.Vec2{X: 0, Y: 0},
	)
	imgui.SetNextWindowSizeV(
		imgui.Vec2{X: windowWidth, Y: windowHeight},
		imgui.CondAlways,
	)

	// Make window non-movable, non-resizable
	flags := imgui.WindowFlagsNoResize |
	         imgui.WindowFlagsNoMove |
	         imgui.WindowFlagsNoCollapse |
	         imgui.WindowFlagsNoTitleBar

	// Begin always requires a matching End, regardless of return value
	if imgui.BeginV("Loading Map", nil, flags) {
		// Title
		titleSize := imgui.CalcTextSize("Loading Map...")
		imgui.SetCursorPosX((windowWidth - titleSize.X) / 2)
		imgui.Text("Loading Map...")
		imgui.Spacing()
		imgui.Spacing()

		// Animated progress bar
		// Create a cycling animation effect (0 -> 1 -> 0)
		progress := float32(math.Sin(float64(view.animationTime)*2.0)*0.5 + 0.5)
		imgui.ProgressBarV(progress, imgui.Vec2{X: -1, Y: 0}, "")

		imgui.Spacing()
		imgui.Spacing()

		// Status text
		view.statusText.Render()

		imgui.Spacing()
		imgui.Spacing()
		imgui.Spacing()

		// Center cancel button
		buttonWidth := float32(120)
		imgui.SetCursorPosX((windowWidth - buttonWidth) / 2)
		view.cancelButton.Draw()
	}
	// ALWAYS call End after Begin, even if Begin returns false
	imgui.End()
}

// getLoadingStateMessage returns a user-friendly message for each loading state
func getLoadingStateMessage(state int) string {
	switch state {
	case messages.LoadingProgressStateStarted:
		return "Initializing map loading..."
	case messages.LoadingProgressStateBSPParsed:
		return "Parsing BSP file..."
	case messages.LoadingProgressStateGeometryLoaded:
		return "Loading BSP geometry..."
	case messages.LoadingProgressStateStaticPropsLoaded:
		return "Loading static props..."
	case messages.LoadingProgressStateEntitiesLoaded:
		return "Loading entities..."
	case messages.LoadingProgressStateDynamicPropsLoaded:
		return "Preparing dynamic props..."
	case messages.LoadingProgressStateFinished:
		return "Finalizing..."
	case messages.LoadingProgressStateError:
		return "Loading failed or cancelled"
	default:
		return "Loading..."
	}
}
