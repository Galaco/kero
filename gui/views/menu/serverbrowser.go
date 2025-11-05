package menu

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/galaco/kero/browser"
	"github.com/galaco/kero/framework/gui"
)

// ServerBrowser is a GUI component for browsing and querying Source Engine servers
// It provides a tabbed interface for different browser modes (Query, Favorites, etc.)
type ServerBrowser struct {
	// Window state
	isOpen bool

	// Tab state
	selectedTab int // 0 = Query tab (future: 1 = Favorites, 2 = History, etc.)

	// Query tab state
	addressInput string // User's input for server address
	resultText   string // Display text for query results

	// UI state flags
	queryButtonEnabled bool

	// Dependencies
	browserState *browser.ServerBrowser // Business logic layer
}

// NewServerBrowser creates a new server browser view
func NewServerBrowser(browserState *browser.ServerBrowser) *ServerBrowser {
	return &ServerBrowser{
		isOpen:             false,
		selectedTab:        0,
		addressInput:       "localhost:27015", // Sensible default for testing
		resultText:         "Enter a server address and click Query to retrieve server information.",
		queryButtonEnabled: true,
		browserState:       browserState,
	}
}

// Render draws the server browser window
// This should be called every frame from the menu view
func (sb *ServerBrowser) Render() {
	if !sb.isOpen {
		return
	}

	// Create window with close button
	// ImGui will set isOpen to false when user clicks X
	if gui.StartPanelV("Server Browser", &sb.isOpen, imgui.WindowFlagsNone) {
		sb.renderContent()
	}
	gui.EndPanel()
}

// renderContent renders the window contents (tabs and content)
func (sb *ServerBrowser) renderContent() {
	// Render tab bar
	if gui.BeginTabBar("BrowserTabs") {
		// Query tab (the only tab for now)
		if gui.BeginTabItem("Query") {
			sb.renderQueryTab()
			gui.EndTabItem()
		}

		// Future tabs can be added here:
		// if gui.BeginTabItem("Favorites") {
		//     sb.renderFavoritesTab()
		//     gui.EndTabItem()
		// }
		//
		// if gui.BeginTabItem("History") {
		//     sb.renderHistoryTab()
		//     gui.EndTabItem()
		// }

		gui.EndTabBar()
	}
}

// renderQueryTab renders the Query tab content
func (sb *ServerBrowser) renderQueryTab() {
	// Instructions
	imgui.Text("Enter a server address (IP:PORT) to query server information:")
	imgui.Spacing()

	// Address input field
	imgui.Text("Server Address:")
	imgui.SameLine()
	imgui.InputTextWithHint("##address", "e.g., 192.168.1.100:27015", &sb.addressInput, 0, nil)

	// Check if query is in progress
	isQuerying := sb.browserState.IsQuerying()

	// Query button (disabled during query)
	if isQuerying {
		// Show disabled button with "Querying..." text
		imgui.TextDisabled("Querying...")
	} else {
		// Show active button
		gui.NewButton("query_button", "Query", func() {
			sb.onQueryButtonPressed()
		}).Draw()
	}

	// Visual separator
	imgui.Spacing()
	imgui.Separator()
	imgui.Spacing()

	// Results section
	imgui.Text("Results:")

	// Multi-line text box for results (read-only)
	// Use -1 for width to fill available space, 300 for fixed height
	resultSize := imgui.Vec2{X: -1, Y: 300}
	imgui.InputTextMultiline("##results", &sb.resultText, resultSize, imgui.InputTextFlagsReadOnly, nil)

	// Show query status if in progress
	if isQuerying {
		imgui.Spacing()
		currentAddress := sb.browserState.GetCurrentAddress()
		imgui.TextColored(
			imgui.Vec4{X: 0.3, Y: 0.7, Z: 1.0, W: 1.0}, // Blue color
			fmt.Sprintf("Querying %s...", currentAddress),
		)
	}
}

// onQueryButtonPressed handles the Query button click
func (sb *ServerBrowser) onQueryButtonPressed() {
	address := sb.addressInput

	// Validate input
	if address == "" {
		sb.resultText = "Error: Please enter a server address.\n\nExample: 192.168.1.100:27015"
		return
	}

	// Clear previous results and show "querying" message
	sb.resultText = fmt.Sprintf("Querying %s...\n\nPlease wait, this may take a few seconds.", address)

	// Trigger async query (business logic layer)
	err := sb.browserState.Query(address)
	if err != nil {
		// Validation error (e.g., concurrent query)
		sb.resultText = fmt.Sprintf("Error: %v\n\nPlease try again.", err)
	}

	// Note: Results will be updated via UpdateFromBrowserState()
	// when the query completes and events are processed
}

// UpdateFromBrowserState updates the view from browser state changes
// This should be called when browser events are received (QueryCompleted/QueryFailed)
func (sb *ServerBrowser) UpdateFromBrowserState() {
	// Get the latest result from browser state
	result, err := sb.browserState.GetLastResult()

	if err != nil {
		// Query failed
		sb.resultText = fmt.Sprintf("Query Failed\n\n%v\n\n", err)
		sb.resultText += "Possible causes:\n"
		sb.resultText += "- Server is offline or unreachable\n"
		sb.resultText += "- Incorrect IP address or port\n"
		sb.resultText += "- Network firewall blocking UDP traffic\n"
		sb.resultText += "- Server is not a Source Engine game server"
		return
	}

	if result != nil {
		// Query succeeded - format the result for display
		sb.resultText = result.FormatForDisplay()
	}
}

// Open shows the server browser window
func (sb *ServerBrowser) Open() {
	sb.isOpen = true
}

// Close hides the server browser window
func (sb *ServerBrowser) Close() {
	sb.isOpen = false
}

// IsOpen returns true if the window is currently visible
func (sb *ServerBrowser) IsOpen() bool {
	return sb.isOpen
}

// SetDefaultAddress sets the default address in the input field
func (sb *ServerBrowser) SetDefaultAddress(address string) {
	sb.addressInput = address
}

// GetCurrentAddress returns the current address in the input field
func (sb *ServerBrowser) GetCurrentAddress() string {
	return sb.addressInput
}
