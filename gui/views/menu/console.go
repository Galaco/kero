package menu

import (
	"sort"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/gui"
	"github.com/galaco/kero/framework/gui/theme"
)

type consoleMessage struct {
	Color imgui.Vec4
	Text  *gui.Text
	Level console.LogLevel
}

func newConsoleMessage(logLevel console.LogLevel, message string) consoleMessage {
	var color imgui.Vec4

	switch logLevel {
	case console.LevelUnknown:
		color = theme.ColorLogUnknown
	case console.LevelFatal:
		color = theme.ColorLogFatal
	case console.LevelError:
		color = theme.ColorLogError
	case console.LevelWarning:
		color = theme.ColorLogWarning
	case console.LevelInfo:
		color = theme.ColorLogInfo
	case console.LevelSuccess:
		color = theme.ColorLogSuccess
	}

	return consoleMessage{
		Color: color,
		Text:  gui.NewText(message),
		Level: logLevel,
	}
}

const maxConsoleMessages = 1000

// Special constant for "All" tab
const allTabValue = console.LogLevel(-1)

type Console struct {
	messages []consoleMessage

	// Tab management
	activeTab             console.LogLevel
	messageIndicesByLevel map[console.LogLevel][]int

	commandInput              string
	autocompleteSelectedIndex int
	shouldRefocus             bool
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
	// StartPanel always requires a matching EndPanel, regardless of return value
	// Use NoCollapse flag to prevent the panel from being collapsed
	if gui.StartPanelV("Console", nil, imgui.WindowFlagsNoCollapse) {
		// Render tab bar
		if imgui.BeginTabBar("ConsoleTabs") {
			// "All" tab (always shown)
			if imgui.BeginTabItem("All") {
				view.activeTab = allTabValue
				imgui.EndTabItem()
			}

			// Tab for each log level (always shown)
			if imgui.BeginTabItem("Fatal") {
				view.activeTab = console.LevelFatal
				imgui.EndTabItem()
			}

			if imgui.BeginTabItem("Error") {
				view.activeTab = console.LevelError
				imgui.EndTabItem()
			}

			if imgui.BeginTabItem("Warning") {
				view.activeTab = console.LevelWarning
				imgui.EndTabItem()
			}

			if imgui.BeginTabItem("Info") {
				view.activeTab = console.LevelInfo
				imgui.EndTabItem()
			}

			if imgui.BeginTabItem("Success") {
				view.activeTab = console.LevelSuccess
				imgui.EndTabItem()
			}

			imgui.EndTabBar()
		}

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

		// Get messages for the currently active tab
		visibleMessages := view.getVisibleMessages()

		// Messages area (subtract space for input box and autocomplete)
		imgui.BeginChildStrV("ConsoleMessages", imgui.Vec2{X: -1, Y: -(24 + autocompleteHeight)}, 0, 0)
		for _, s := range visibleMessages {
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
					imgui.PushStyleColorVec4(imgui.ColText, theme.ColorHighlight) // Highlighted
				} else {
					imgui.PushStyleColorVec4(imgui.ColText, theme.ColorMuted) // Muted
				}
				imgui.Text(option)
				imgui.PopStyleColor()
			}
			imgui.EndChild()
		}

		// Input box - standard InputTextWithHint
		imgui.PushItemWidth(-1)

		// Set focus if needed (must be called before the input field)
		if view.shouldRefocus {
			imgui.SetKeyboardFocusHere()
			view.shouldRefocus = false
		}

		// Use standard input text (Enter returns true to submit)
		if imgui.InputTextWithHint("##console_input", "", &view.commandInput, imgui.InputTextFlagsEnterReturnsTrue, nil) {
			// Check if input exactly matches a command/convar
			isExactMatch := false
			for _, option := range autocompleteOptions {
				if view.commandInput == option {
					isExactMatch = true
					break
				}
			}

			// Enter was pressed - either submit command or accept autocomplete
			if isExactMatch || len(autocompleteOptions) == 0 {
				// Exact match or no autocomplete - submit the command
				err := console.ExecuteCommand(view.commandInput)
				if err != nil {
					console.PrintString(console.LevelError, err.Error())
				}
				view.commandInput = ""
				view.shouldRefocus = true
			} else {
				// Partial match - autocomplete to the highlighted option
				view.commandInput = autocompleteOptions[view.autocompleteSelectedIndex]
				view.autocompleteSelectedIndex = -1
				view.shouldRefocus = true
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
	}
	// ALWAYS call EndPanel after StartPanel, even if StartPanel returns false
	gui.EndPanel()
}

// getVisibleMessages returns messages for the currently active tab
func (view *Console) getVisibleMessages() []consoleMessage {
	// If showing "All" tab or not initialized, return all messages
	if view.activeTab == allTabValue || view.messageIndicesByLevel == nil {
		return view.messages
	}

	// Get indices for active tab
	indices := view.messageIndicesByLevel[view.activeTab]

	// Build filtered slice
	filtered := make([]consoleMessage, 0, len(indices))
	for _, idx := range indices {
		if idx < len(view.messages) {
			filtered = append(filtered, view.messages[idx])
		}
	}
	return filtered
}

// rebuildIndices reconstructs the message indices map after messages have been shifted
func (view *Console) rebuildIndices() {
	// Clear existing indices
	view.messageIndicesByLevel = make(map[console.LogLevel][]int)

	// Rebuild from current messages
	for i, msg := range view.messages {
		// Add to "All" tab
		view.messageIndicesByLevel[allTabValue] = append(
			view.messageIndicesByLevel[allTabValue], i)

		// Add to specific level tab
		view.messageIndicesByLevel[msg.Level] = append(
			view.messageIndicesByLevel[msg.Level], i)
	}
}

func (view *Console) AddMessage(level console.LogLevel, message string) {
	// Initialize with pre-allocated capacity to avoid early reallocations
	if view.messages == nil {
		view.messages = make([]consoleMessage, 0, maxConsoleMessages)
		view.messageIndicesByLevel = make(map[console.LogLevel][]int)
		view.activeTab = allTabValue // Default to "All" tab
	}

	// If at capacity, remove oldest message(s) to maintain limit
	if len(view.messages) >= maxConsoleMessages {
		// Shift slice to remove oldest message (index 0)
		// This is more efficient than multiple individual removals
		copy(view.messages, view.messages[1:])
		view.messages = view.messages[:len(view.messages)-1]

		// Rebuild indices since all indices have shifted
		view.rebuildIndices()
	}

	// Add new message
	newIndex := len(view.messages)
	view.messages = append(view.messages, newConsoleMessage(level, message))

	// Add index to "All" tab
	view.messageIndicesByLevel[allTabValue] = append(
		view.messageIndicesByLevel[allTabValue], newIndex)

	// Add index to specific level tab
	view.messageIndicesByLevel[level] = append(
		view.messageIndicesByLevel[level], newIndex)
}
