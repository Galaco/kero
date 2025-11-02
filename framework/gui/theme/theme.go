package theme

import "github.com/AllenDang/cimgui-go/imgui"

// Apply configures the ImGui style with a modern dark theme
// This is called once during UI context initialization
func Apply() {
	style := imgui.CurrentStyle()

	// ========================================
	// Style Variables - Spacing, Rounding, Sizes
	// ========================================

	// Rounding for modern appearance
	style.SetWindowRounding(6.0)
	style.SetChildRounding(4.0)
	style.SetFrameRounding(3.0)
	style.SetPopupRounding(4.0)
	style.SetScrollbarRounding(6.0)
	style.SetGrabRounding(3.0)
	style.SetTabRounding(4.0)

	// Spacing and padding
	style.SetWindowPadding(imgui.Vec2{X: 12, Y: 12})
	style.SetFramePadding(imgui.Vec2{X: 8, Y: 4})
	style.SetItemSpacing(imgui.Vec2{X: 8, Y: 6})
	style.SetItemInnerSpacing(imgui.Vec2{X: 6, Y: 4})
	style.SetIndentSpacing(22.0)

	// Borders
	style.SetWindowBorderSize(1.0)
	style.SetChildBorderSize(1.0)
	style.SetPopupBorderSize(1.0)
	style.SetFrameBorderSize(0.0)
	style.SetTabBorderSize(0.0)

	// Scrollbar
	style.SetScrollbarSize(14.0)

	// Grab (sliders, scrollbars)
	style.SetGrabMinSize(10.0)

	// ========================================
	// Color Scheme - Modern Dark Theme
	// ========================================

	colors := style.Colors()

	// Background colors - Dark slate/charcoal base
	colors[imgui.ColWindowBg] = imgui.Vec4{X: 0.12, Y: 0.12, Z: 0.14, W: 0.95}           // Main window background
	colors[imgui.ColChildBg] = imgui.Vec4{X: 0.15, Y: 0.15, Z: 0.17, W: 1.0}             // Child window background
	colors[imgui.ColPopupBg] = imgui.Vec4{X: 0.10, Y: 0.10, Z: 0.12, W: 0.95}            // Popup background
	colors[imgui.ColBorder] = imgui.Vec4{X: 0.25, Y: 0.25, Z: 0.28, W: 0.6}              // Border color
	colors[imgui.ColBorderShadow] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.0}           // Border shadow (disabled)

	// Frame backgrounds (input boxes, etc.)
	colors[imgui.ColFrameBg] = imgui.Vec4{X: 0.18, Y: 0.18, Z: 0.20, W: 1.0}             // Frame background
	colors[imgui.ColFrameBgHovered] = imgui.Vec4{X: 0.22, Y: 0.22, Z: 0.25, W: 1.0}      // Frame hovered
	colors[imgui.ColFrameBgActive] = imgui.Vec4{X: 0.25, Y: 0.25, Z: 0.28, W: 1.0}       // Frame active/focused

	// Title bar
	colors[imgui.ColTitleBg] = imgui.Vec4{X: 0.10, Y: 0.10, Z: 0.12, W: 1.0}             // Title background
	colors[imgui.ColTitleBgActive] = imgui.Vec4{X: 0.15, Y: 0.15, Z: 0.18, W: 1.0}       // Title active
	colors[imgui.ColTitleBgCollapsed] = imgui.Vec4{X: 0.08, Y: 0.08, Z: 0.10, W: 0.75}   // Title collapsed

	// Menu bar
	colors[imgui.ColMenuBarBg] = imgui.Vec4{X: 0.15, Y: 0.15, Z: 0.17, W: 1.0}           // Menu bar background

	// Scrollbar
	colors[imgui.ColScrollbarBg] = imgui.Vec4{X: 0.10, Y: 0.10, Z: 0.12, W: 1.0}         // Scrollbar background
	colors[imgui.ColScrollbarGrab] = imgui.Vec4{X: 0.30, Y: 0.30, Z: 0.35, W: 1.0}       // Scrollbar grab
	colors[imgui.ColScrollbarGrabHovered] = imgui.Vec4{X: 0.40, Y: 0.40, Z: 0.45, W: 1.0} // Scrollbar grab hovered
	colors[imgui.ColScrollbarGrabActive] = imgui.Vec4{X: 0.50, Y: 0.50, Z: 0.55, W: 1.0}  // Scrollbar grab active

	// Checkbox/Radio
	colors[imgui.ColCheckMark] = imgui.Vec4{X: 0.2, Y: 0.7, Z: 1.0, W: 1.0}              // Checkmark (accent blue)

	// Slider/Grab
	colors[imgui.ColSliderGrab] = imgui.Vec4{X: 0.25, Y: 0.65, Z: 0.95, W: 1.0}          // Slider grab (accent)
	colors[imgui.ColSliderGrabActive] = imgui.Vec4{X: 0.3, Y: 0.75, Z: 1.0, W: 1.0}      // Slider grab active

	// Button - Modern blue accent
	colors[imgui.ColButton] = imgui.Vec4{X: 0.20, Y: 0.55, Z: 0.85, W: 1.0}              // Button normal
	colors[imgui.ColButtonHovered] = imgui.Vec4{X: 0.25, Y: 0.65, Z: 0.95, W: 1.0}       // Button hovered
	colors[imgui.ColButtonActive] = imgui.Vec4{X: 0.15, Y: 0.50, Z: 0.80, W: 1.0}        // Button clicked

	// Header (collapsing header, tree node)
	colors[imgui.ColHeader] = imgui.Vec4{X: 0.20, Y: 0.55, Z: 0.85, W: 0.7}              // Header normal
	colors[imgui.ColHeaderHovered] = imgui.Vec4{X: 0.25, Y: 0.65, Z: 0.95, W: 0.8}       // Header hovered
	colors[imgui.ColHeaderActive] = imgui.Vec4{X: 0.20, Y: 0.60, Z: 0.90, W: 1.0}        // Header active

	// Separator
	colors[imgui.ColSeparator] = imgui.Vec4{X: 0.25, Y: 0.25, Z: 0.28, W: 1.0}           // Separator line
	colors[imgui.ColSeparatorHovered] = imgui.Vec4{X: 0.30, Y: 0.60, Z: 0.90, W: 1.0}    // Separator hovered
	colors[imgui.ColSeparatorActive] = imgui.Vec4{X: 0.35, Y: 0.70, Z: 1.0, W: 1.0}      // Separator active

	// Resize grip
	colors[imgui.ColResizeGrip] = imgui.Vec4{X: 0.20, Y: 0.55, Z: 0.85, W: 0.5}          // Resize grip
	colors[imgui.ColResizeGripHovered] = imgui.Vec4{X: 0.25, Y: 0.65, Z: 0.95, W: 0.7}   // Resize grip hovered
	colors[imgui.ColResizeGripActive] = imgui.Vec4{X: 0.30, Y: 0.75, Z: 1.0, W: 0.9}     // Resize grip active

	// Tab
	colors[imgui.ColTab] = imgui.Vec4{X: 0.15, Y: 0.45, Z: 0.75, W: 0.8}                 // Tab inactive
	colors[imgui.ColTabHovered] = imgui.Vec4{X: 0.25, Y: 0.65, Z: 0.95, W: 1.0}          // Tab hovered
	// Note: ColTabActive, ColTabUnfocused, ColTabUnfocusedActive may not be available in all ImGui versions

	// Plot/Graph colors
	colors[imgui.ColPlotLines] = imgui.Vec4{X: 0.3, Y: 0.95, Z: 0.5, W: 1.0}             // Plot lines (green)
	colors[imgui.ColPlotLinesHovered] = imgui.Vec4{X: 0.4, Y: 1.0, Z: 0.6, W: 1.0}       // Plot lines hovered
	colors[imgui.ColPlotHistogram] = imgui.Vec4{X: 0.2, Y: 0.7, Z: 1.0, W: 1.0}          // Plot histogram (blue)
	colors[imgui.ColPlotHistogramHovered] = imgui.Vec4{X: 0.3, Y: 0.8, Z: 1.0, W: 1.0}   // Plot histogram hovered

	// Text
	colors[imgui.ColText] = imgui.Vec4{X: 0.95, Y: 0.95, Z: 0.95, W: 1.0}                // Primary text (bright)
	colors[imgui.ColTextDisabled] = imgui.Vec4{X: 0.50, Y: 0.50, Z: 0.50, W: 1.0}        // Disabled text (gray)

	// Text selection
	colors[imgui.ColTextSelectedBg] = imgui.Vec4{X: 0.20, Y: 0.55, Z: 0.85, W: 0.45}     // Text selection background

	// Drag and drop
	colors[imgui.ColDragDropTarget] = imgui.Vec4{X: 0.2, Y: 0.7, Z: 1.0, W: 0.9}         // Drag drop target

	// Navigation
	// Note: ColNavHighlight, ColNavWindowingHighlight, ColNavWindowingDimBg may not be available in all ImGui versions

	// Modal window dimming
	colors[imgui.ColModalWindowDimBg] = imgui.Vec4{X: 0.10, Y: 0.10, Z: 0.10, W: 0.6}    // Modal dim background

	// Table colors
	colors[imgui.ColTableHeaderBg] = imgui.Vec4{X: 0.18, Y: 0.18, Z: 0.20, W: 1.0}       // Table header background
	colors[imgui.ColTableBorderStrong] = imgui.Vec4{X: 0.30, Y: 0.30, Z: 0.35, W: 1.0}   // Table outer border
	colors[imgui.ColTableBorderLight] = imgui.Vec4{X: 0.22, Y: 0.22, Z: 0.25, W: 1.0}    // Table inner border
	colors[imgui.ColTableRowBg] = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.0}             // Table row background (transparent)
	colors[imgui.ColTableRowBgAlt] = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.06}         // Table row alternate background
}
