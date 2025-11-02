package menu

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/galaco/kero/framework/gui/theme"
	"github.com/galaco/kero/framework/metrics"
)

// Performance displays real-time performance metrics
type Performance struct {
	metricsCollector *metrics.Collector

	// Display settings
	graphHeight float32
	graphWidth  float32
}

// NewPerformance creates a new performance view
func NewPerformance(metricsCollector *metrics.Collector) *Performance {
	return &Performance{
		metricsCollector: metricsCollector,
		// Default graph size
		graphHeight:      200,
		graphWidth:       600,
	}
}

// SetGraphHeight sets the graph height
func (p *Performance) SetGraphHeight(height float32) {
	p.graphHeight = height
}

// SetGraphWidth sets the graph width
func (p *Performance) SetGraphWidth(width float32) {
	p.graphWidth = width
}

// Render draws the performance overlay in the bottom-left corner
func (p *Performance) Render() {
	// Get all metrics
	allMetrics := p.metricsCollector.GetAllMetrics()

	// Create a fixed window in the bottom-left corner
	// Set window position and size constraints
	windowFlags := imgui.WindowFlagsNoDecoration |
		imgui.WindowFlagsNoMove |
		imgui.WindowFlagsNoSavedSettings |
		imgui.WindowFlagsNoFocusOnAppearing |
		imgui.WindowFlagsNoNav |
		imgui.WindowFlagsAlwaysAutoResize

	// Position in bottom-left with padding
	padding := float32(10.0)
	viewport := imgui.MainViewport()
	workPos := viewport.WorkPos()
	workSize := viewport.WorkSize()

	windowPos := imgui.Vec2{
		X: workPos.X + padding,
		Y: workPos.Y + workSize.Y - padding,
	}

	imgui.SetNextWindowPosV(windowPos, imgui.CondAlways, imgui.Vec2{X: 0.0, Y: 1.0})
	imgui.SetNextWindowBgAlpha(0.35) // Transparent background

	if imgui.BeginV("Performance", nil, windowFlags) {
		// Plot system times on the same graph
		p.plotMultipleMetrics(allMetrics)
	}
	imgui.End()
}

// plotMultipleMetrics draws multiple metrics on overlapping graphs
func (p *Performance) plotMultipleMetrics(allMetrics map[string]*metrics.SystemMetrics) {
	// Define system names and colors
	type MetricConfig struct {
		name  string
		label string
		color imgui.Vec4
	}

	configs := []MetricConfig{
		{"input", "Input", theme.ColorGraphRed},
		{"physics_total", "Physics", theme.ColorGraphBlue},
		{"scene_update", "Scene", theme.ColorGraphOrange},
		{"render", "Render", theme.ColorGraphCyan},
		{"gui", "GUI", theme.ColorGraphMagenta},
	}

	// Calculate max value for consistent scaling
	maxValue := float32(0.0)
	for _, config := range configs {
		if metric, exists := allMetrics[config.name]; exists && metric.History != nil {
			for _, v := range metric.History {
				if float32(v) > maxValue {
					maxValue = float32(v)
				}
			}
		}
	}

	// Add some headroom
	maxValue *= 1.1
	if maxValue < 1.0 {
		maxValue = 1.0
	}

	// Plot each metric
	// We need to overlay them, so we use a child window with ImGuiWindowFlags_NoBackground
	firstPlot := true
	for _, config := range configs {
		metric, exists := allMetrics[config.name]
		if !exists || metric.History == nil || len(metric.History) == 0 {
			continue
		}

		// Convert to float32 for ImGui
		values := make([]float32, len(metric.History))
		for i, v := range metric.History {
			values[i] = float32(v)
		}

		// For the first plot, show the label. For subsequent plots, use overlay label
		var label string
		if firstPlot {
			label = "System Times (ms)"
			firstPlot = false
		} else {
			label = "##overlay" + config.name
		}

		// Plot the line with color
		imgui.PushStyleColorVec4(imgui.ColPlotLines, config.color)

		// Use SetCursorPos to overlay graphs
		if label[:2] == "##" {
			cursorPos := imgui.CursorPos()
			// Move cursor back up by graph height + spacing to overlay
			imgui.SetCursorPos(imgui.Vec2{X: cursorPos.X, Y: cursorPos.Y - p.graphHeight - imgui.CurrentStyle().ItemSpacing().Y})
		}

		imgui.PlotLinesFloatPtrV(
			label,
			&values[0],
			int32(len(values)),
			0,
			"",
			0,
			maxValue,
			imgui.Vec2{X: p.graphWidth, Y: p.graphHeight},
			4, // stride (sizeof(float32))
		)
		imgui.PopStyleColor()
	}

	// Legend with current time values
	imgui.Separator()
	imgui.Text("System Times (ms):")
	for _, config := range configs {
		metric, exists := allMetrics[config.name]
		if !exists {
			continue
		}

		imgui.PushStyleColorVec4(imgui.ColText, config.color)
		timeText := fmt.Sprintf("%s: %.2f ms", config.label, metric.CurrentValue)

		// Special case for physics to show steps
		if config.name == "physics_total" {
			if steps, stepsExist := allMetrics["physics_steps"]; stepsExist {
				timeText = fmt.Sprintf("%s: %.2f ms (%.0f steps)", config.label, metric.CurrentValue, steps.CurrentValue)
			}
		}

		imgui.Text(timeText)
		imgui.PopStyleColor()
	}

	// Display FPS and frame total
	imgui.Separator()
	if fps, exists := allMetrics["fps"]; exists {
		imgui.Text(fmt.Sprintf("FPS: %.1f", fps.CurrentValue))
	}

	if frameTotal, exists := allMetrics["frame_total"]; exists {
		imgui.Text(fmt.Sprintf("Frame Total: %.2f ms", frameTotal.CurrentValue))
	}
}
