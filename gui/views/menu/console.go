package menu

import (
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/gui"
	"github.com/inkyblackness/imgui-go/v2"
	"log"
)

type consoleMessage struct {
	Color imgui.Vec4
	Text  *gui.Text
}

func newConsoleMessage(logLevel console.LogLevel, message string) consoleMessage {
	var color imgui.Vec4

	switch logLevel {
	case console.LevelUnknown:
		color = imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}
	case console.LevelFatal:
		color = imgui.Vec4{X: 1, Y: 0, Z: 0, W: 1}
	case console.LevelError:
		color = imgui.Vec4{X: 1, Y: 0, Z: 0, W: 1}
	case console.LevelWarning:
		color = imgui.Vec4{X: 1, Y: 1, Z: 0, W: 1}
	case console.LevelInfo:
		color = imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}
	case console.LevelSuccess:
		color = imgui.Vec4{X: 0, Y: 1, Z: 0, W: 1}
	}

	return consoleMessage{
		Color: color,
		Text:  gui.NewText(message),
	}
}

type Console struct {
	messages []consoleMessage

	commandInput string
}

func (view *Console) commandInputCallback(data imgui.InputTextCallbackData) int32 {
	log.Println(data.EventKey())
	if data.EventKey() == imgui.KeyEnter {
		console.PrintString(console.LevelInfo, view.commandInput)
	}
	if data.EventFlag() & imgui.InputTextFlagsEnterReturnsTrue != 0 {
		console.PrintString(console.LevelInfo, view.commandInput)
	}

	return 0
}

func (view *Console) Render() {
	if gui.StartPanel("Console") {
		for _, s := range view.messages {
			imgui.PushStyleColor(imgui.StyleColorText, s.Color)
			s.Text.Render()
			imgui.PopStyleColor()
		}

		if imgui.InputTextV("CommandInput", &view.commandInput, imgui.InputTextFlagsEnterReturnsTrue, view.commandInputCallback) {
			err := console.ExecuteCommand(view.commandInput)
			if err != nil {
				console.PrintString(console.LevelError, err.Error())
			}
			view.commandInput = ""
		}

		gui.EndPanel()
	}
}

func (view *Console) AddMessage(level console.LogLevel, message string) {
	view.messages = append(view.messages, newConsoleMessage(level, message))
}
