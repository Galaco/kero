package menu

import (
	"fmt"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/galaco/kero/framework/metrics"
)

// Performance displays real-time performance metrics
type Performance struct {
	metricsCollector *metrics.Collector

	// Toggles for which metrics to display
	showInput        bool
	showPhysics      bool
	showScene        bool
	showRender       bool
	showGUI          bool
	showEvents       bool
	showFrameTotal   bool
	showFPS          bool

	// Display settings
	graphHeight      float32
	graphWidth       float32
}

// NewPerformance creates a new performance view
func NewPerformance(metricsCollector *metrics.Collector) *Performance {
	return &Performance{
		metricsCollector: metricsCollector,
		// Default toggles
		showInput:      true,
		showPhysics:    true,
		showScene:      true,
		showRender:     true,
		showGUI:        true,
		showEvents:     false, // Events are usually very fast
		showFrameTotal: true,
		showFPS:        true,
		// Default graph size
		graphHeight:    200,
		graphWidth:     600,
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

// Render draws the performance panel
func (p *Performance) Render() {
	if !imgui.CollapsingHeaderTreeNodeFlags("Performance Metrics") {
		return
	}

	// Display settings
	imgui.Text("Display Options:")
	imgui.Checkbox("Input", &p.showInput)
	imgui.SameLine()
	imgui.Checkbox("Physics", &p.showPhysics)
	imgui.SameLine()
	imgui.Checkbox("Scene", &p.showScene)
	imgui.SameLine()
	imgui.Checkbox("Render", &p.showRender)
	imgui.SameLine()
	imgui.Checkbox("GUI", &p.showGUI)

	imgui.Checkbox("Events", &p.showEvents)
	imgui.SameLine()
	imgui.Checkbox("Frame Total", &p.showFrameTotal)
	imgui.SameLine()
	imgui.Checkbox("FPS", &p.showFPS)

	imgui.Separator()

	// Get all metrics
	allMetrics := p.metricsCollector.GetAllMetrics()

	// Display current values
	imgui.Text("Current Frame Times (ms):")

	if fps, exists := allMetrics["fps"]; exists && p.showFPS {
		imgui.Text(fmt.Sprintf("FPS: %.1f", fps.CurrentValue))
	}

	if frameTotal, exists := allMetrics["frame_total"]; exists && p.showFrameTotal {
		imgui.Text(fmt.Sprintf("Frame Total: %.2f ms", frameTotal.CurrentValue))
	}

	if input, exists := allMetrics["input"]; exists && p.showInput {
		imgui.Text(fmt.Sprintf("Input: %.2f ms", input.CurrentValue))
	}

	if physics, exists := allMetrics["physics_total"]; exists && p.showPhysics {
		steps := allMetrics["physics_steps"]
		imgui.Text(fmt.Sprintf("Physics: %.2f ms (%.0f steps)", physics.CurrentValue, steps.CurrentValue))
	}

	if scene, exists := allMetrics["scene_update"]; exists && p.showScene {
		imgui.Text(fmt.Sprintf("Scene: %.2f ms", scene.CurrentValue))
	}

	if render, exists := allMetrics["render"]; exists && p.showRender {
		imgui.Text(fmt.Sprintf("Render: %.2f ms", render.CurrentValue))
	}

	if gui, exists := allMetrics["gui"]; exists && p.showGUI {
		imgui.Text(fmt.Sprintf("GUI: %.2f ms", gui.CurrentValue))
	}

	if p.showEvents {
		if preUpdate, exists := allMetrics["event_preupdate"]; exists {
			imgui.Text(fmt.Sprintf("Event PreUpdate: %.2f ms", preUpdate.CurrentValue))
		}
		if postUpdate, exists := allMetrics["event_postupdate"]; exists {
			imgui.Text(fmt.Sprintf("Event PostUpdate: %.2f ms", postUpdate.CurrentValue))
		}
		if preRender, exists := allMetrics["event_prerender"]; exists {
			imgui.Text(fmt.Sprintf("Event PreRender: %.2f ms", preRender.CurrentValue))
		}
		if postRender, exists := allMetrics["event_postrender"]; exists {
			imgui.Text(fmt.Sprintf("Event PostRender: %.2f ms", postRender.CurrentValue))
		}
	}

	imgui.Separator()
	imgui.Text("Performance History:")

	// Plot frame time graph
	p.plotMetric("Frame Time (ms)", "frame_total", allMetrics, imgui.Vec4{X: 0.0, Y: 1.0, Z: 0.0, W: 1.0})

	// Plot FPS graph
	if p.showFPS {
		p.plotMetric("FPS", "fps", allMetrics, imgui.Vec4{X: 1.0, Y: 1.0, Z: 0.0, W: 1.0})
	}

	// Plot system times on the same graph
	imgui.Text("System Times:")
	p.plotMultipleMetrics(allMetrics)
}

// plotMetric draws a single metric as a line graph
func (p *Performance) plotMetric(label string, metricName string, allMetrics map[string]*metrics.SystemMetrics, color imgui.Vec4) {
	metric, exists := allMetrics[metricName]
	if !exists || metric.History == nil || len(metric.History) == 0 {
		return
	}

	// Convert to float32 for ImGui
	values := make([]float32, len(metric.History))
	for i, v := range metric.History {
		values[i] = float32(v)
	}

	// Plot the line
	imgui.PushStyleColorVec4(imgui.ColPlotLines, color)
	imgui.PlotLinesFloatPtrV(
		label,
		&values[0],
		int32(len(values)),
		0,
		"",
		0,
		0, // Auto scale
		imgui.Vec2{X: p.graphWidth, Y: p.graphHeight},
		4, // stride (sizeof(float32))
	)
	imgui.PopStyleColor()
}

// plotMultipleMetrics draws multiple metrics on overlapping graphs
func (p *Performance) plotMultipleMetrics(allMetrics map[string]*metrics.SystemMetrics) {
	// Define system names and colors
	type MetricConfig struct {
		name    string
		label   string
		color   imgui.Vec4
		enabled *bool
	}

	configs := []MetricConfig{
		{"input", "Input", imgui.Vec4{X: 1.0, Y: 0.0, Z: 0.0, W: 1.0}, &p.showInput},
		{"physics_total", "Physics", imgui.Vec4{X: 0.0, Y: 0.0, Z: 1.0, W: 1.0}, &p.showPhysics},
		{"scene_update", "Scene", imgui.Vec4{X: 1.0, Y: 0.5, Z: 0.0, W: 1.0}, &p.showScene},
		{"render", "Render", imgui.Vec4{X: 0.0, Y: 1.0, Z: 1.0, W: 1.0}, &p.showRender},
		{"gui", "GUI", imgui.Vec4{X: 1.0, Y: 0.0, Z: 1.0, W: 1.0}, &p.showGUI},
	}

	// Calculate max value for consistent scaling
	maxValue := float32(0.0)
	for _, config := range configs {
		if !*config.enabled {
			continue
		}
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

	// Plot each enabled metric
	// We need to overlay them, so we use a child window with ImGuiWindowFlags_NoBackground
	firstPlot := true
	for _, config := range configs {
		if !*config.enabled {
			continue
		}

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

	// Legend
	imgui.Separator()
	imgui.Text("Legend:")
	for _, config := range configs {
		if !*config.enabled {
			continue
		}
		imgui.PushStyleColorVec4(imgui.ColText, config.color)
		imgui.Text(config.label)
		imgui.PopStyleColor()
		imgui.SameLine()
	}
	imgui.Text("") // End line after legend
}
