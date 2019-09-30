package menu

import (
	"github.com/galaco/kero/framework/console"
	"github.com/inkyblackness/imgui-go"
)

type Console struct {
	messages []string
}

func (view *Console) Render() {
	if imgui.BeginV("Console", nil, imgui.WindowFlagsNoResize|
		imgui.WindowFlagsNoMove|
		imgui.WindowFlagsNoBringToFrontOnFocus|
		imgui.WindowFlagsNoScrollbar|
		imgui.WindowFlagsNoScrollWithMouse|
		imgui.WindowFlagsNoNav|
		imgui.WindowFlagsNoInputs) {

		for _, s := range view.messages {
			imgui.BeginChild("ConsoleScrolling")
			imgui.Text(s)
			imgui.EndChild()
		}

		imgui.End()
	}
}

func (view *Console) AddMessage(level console.LogLevel, message string) {
	view.messages = append(view.messages, message)
}
