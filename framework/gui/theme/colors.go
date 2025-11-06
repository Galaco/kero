package theme

import "github.com/AllenDang/cimgui-go/imgui"

// Semantic colors for consistent use across the application
// These colors are designed to work with the modern dark theme

var (
	// Log level colors
	ColorLogUnknown = imgui.Vec4{X: 0.85, Y: 0.85, Z: 0.85, W: 1.0} // Light gray
	ColorLogFatal   = imgui.Vec4{X: 1.0, Y: 0.2, Z: 0.2, W: 1.0}    // Bright red
	ColorLogError   = imgui.Vec4{X: 0.95, Y: 0.35, Z: 0.35, W: 1.0} // Red
	ColorLogWarning = imgui.Vec4{X: 1.0, Y: 0.75, Z: 0.2, W: 1.0}   // Amber/Orange
	ColorLogInfo    = imgui.Vec4{X: 0.85, Y: 0.85, Z: 0.85, W: 1.0} // Light gray
	ColorLogSuccess = imgui.Vec4{X: 0.3, Y: 0.9, Z: 0.4, W: 1.0}    // Green

	// Chart/Graph colors - vibrant and distinct for overlaying metrics
	ColorGraphRed     = imgui.Vec4{X: 0.95, Y: 0.3, Z: 0.3, W: 1.0}   // Red
	ColorGraphBlue    = imgui.Vec4{X: 0.3, Y: 0.6, Z: 1.0, W: 1.0}    // Blue
	ColorGraphOrange  = imgui.Vec4{X: 1.0, Y: 0.6, Z: 0.2, W: 1.0}    // Orange
	ColorGraphCyan    = imgui.Vec4{X: 0.2, Y: 0.9, Z: 0.9, W: 1.0}    // Cyan
	ColorGraphMagenta = imgui.Vec4{X: 0.95, Y: 0.3, Z: 0.95, W: 1.0}  // Magenta
	ColorGraphGreen   = imgui.Vec4{X: 0.3, Y: 0.95, Z: 0.5, W: 1.0}   // Green
	ColorGraphYellow  = imgui.Vec4{X: 1.0, Y: 0.95, Z: 0.3, W: 1.0}   // Yellow

	// UI state colors
	ColorHighlight = imgui.Vec4{X: 1.0, Y: 0.75, Z: 0.2, W: 1.0}   // Amber - for selected/highlighted items
	ColorMuted     = imgui.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1.0}    // Gray - for disabled/unselected items
	ColorAccent    = imgui.Vec4{X: 0.2, Y: 0.7, Z: 1.0, W: 1.0}    // Bright blue - primary accent
)
