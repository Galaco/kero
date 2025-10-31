package menu

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/gui"
	"sort"
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

// getAutocompleteOptions returns up to 5 commands/convars that match the current input
func (view *Console) getAutocompleteOptions() []string {
	if view.commandInput == "" {
		return nil
	}

	// Get matching commands and convars
	commands := console.GetCommandList(view.commandInput)
	convars := console.GetConvarList(view.commandInput)

	// Combine and sort
	combined := append(commands, convars...)
	sort.Strings(combined)

	// Limit to 5 results
	if len(combined) > 5 {
		combined = combined[:5]
	}

	return combined
}

func (view *Console) Render() {
	if gui.StartPanel("Console") {
		// Get autocomplete options
		autocompleteOptions := view.getAutocompleteOptions()

		// Calculate height for autocomplete area (each option is ~20px, plus some padding)
		autocompleteHeight := float32(0)
		if len(autocompleteOptions) > 0 {
			autocompleteHeight = float32(len(autocompleteOptions)*20 + 5)
		}

		// Messages area (subtract space for input box and autocomplete)
		imgui.BeginChildStrV("ConsoleMessages", imgui.Vec2{X: -1, Y: -(24 + autocompleteHeight)}, 0, 0)
		for _, s := range view.messages {
			imgui.PushStyleColorVec4(imgui.ColText, s.Color)
			s.Text.Render()
			imgui.PopStyleColor()
		}
		imgui.EndChild()

		// Render autocomplete suggestions above input box
		if len(autocompleteOptions) > 0 {
			imgui.BeginChildStrV("AutocompleteArea", imgui.Vec2{X: -1, Y: autocompleteHeight}, 0, 0)
			imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.7, Y: 0.7, Z: 0.7, W: 1})
			for _, option := range autocompleteOptions {
				imgui.Text(option)
			}
			imgui.PopStyleColor()
			imgui.EndChild()
		}

		// Input box - using InputTextWithHint with no callback (Enter key returns true)
		imgui.PushItemWidth(-1)
		if imgui.InputTextWithHint("##console_input", "", &view.commandInput, imgui.InputTextFlagsEnterReturnsTrue, nil) {
			err := console.ExecuteCommand(view.commandInput)
			if err != nil {
				console.PrintString(console.LevelError, err.Error())
			}
			view.commandInput = ""
		}
		imgui.PopItemWidth()

		gui.EndPanel()
	}
}

func (view *Console) AddMessage(level console.LogLevel, message string) {
	if view.messages == nil {
		view.messages = make([]consoleMessage, 0)
	}
	view.messages = append(view.messages, newConsoleMessage(level, message))
}
