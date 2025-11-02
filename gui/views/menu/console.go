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

	commandInput              string
	autocompleteSelectedIndex int
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

		// Manage autocomplete selection index
		if len(autocompleteOptions) == 0 {
			view.autocompleteSelectedIndex = -1
		} else if view.autocompleteSelectedIndex == -1 {
			// First time showing autocomplete, select first option
			view.autocompleteSelectedIndex = 0
		} else if view.autocompleteSelectedIndex >= len(autocompleteOptions) {
			// Options changed, clamp to valid range
			view.autocompleteSelectedIndex = len(autocompleteOptions) - 1
		}

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
			for i, option := range autocompleteOptions {
				// Highlight the selected option
				if i == view.autocompleteSelectedIndex {
					imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 1, Y: 1, Z: 0, W: 1}) // Yellow for selected
				} else {
					imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 0.7, Y: 0.7, Z: 0.7, W: 1}) // Gray for unselected
				}
				imgui.Text(option)
				imgui.PopStyleColor()
			}
			imgui.EndChild()
		}

		// Input box - standard InputTextWithHint
		imgui.PushItemWidth(-1)

		// Use standard input text (Enter returns true to submit)
		if imgui.InputTextWithHint("##console_input", "", &view.commandInput, imgui.InputTextFlagsEnterReturnsTrue, nil) {
			// Enter was pressed - either submit command or accept autocomplete
			if len(autocompleteOptions) > 0 && view.autocompleteSelectedIndex >= 0 {
				// Autocomplete is active - select the highlighted option
				view.commandInput = autocompleteOptions[view.autocompleteSelectedIndex]
				view.autocompleteSelectedIndex = -1
			} else {
				// No autocomplete - submit the command
				err := console.ExecuteCommand(view.commandInput)
				if err != nil {
					console.PrintString(console.LevelError, err.Error())
				}
				view.commandInput = ""
			}
		}

		// Check for autocomplete navigation keys AFTER rendering the input
		// This only works when the input text is active/focused
		if imgui.IsItemActive() && len(autocompleteOptions) > 0 {
			// Handle Up arrow - navigate up in autocomplete
			if imgui.IsKeyPressedBool(imgui.KeyUpArrow) {
				if view.autocompleteSelectedIndex > 0 {
					view.autocompleteSelectedIndex--
				} else {
					view.autocompleteSelectedIndex = len(autocompleteOptions) - 1
				}
			}

			// Handle Down arrow - navigate down in autocomplete
			if imgui.IsKeyPressedBool(imgui.KeyDownArrow) {
				if view.autocompleteSelectedIndex < len(autocompleteOptions)-1 {
					view.autocompleteSelectedIndex++
				} else {
					view.autocompleteSelectedIndex = 0
				}
			}
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
