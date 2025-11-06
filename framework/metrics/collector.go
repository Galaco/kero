package metrics

import (
	"sync"
	"time"
)

// SystemMetrics stores timing data for a specific system
type SystemMetrics struct {
	Name         string
	CurrentValue float64        // Current frame time in milliseconds
	History      []float64      // Ring buffer of historical values
	head         int            // Current write position in ring buffer
	size         int            // Current number of samples (up to capacity)
}

// Collector manages performance metrics for all engine systems
type Collector struct {
	mut           sync.RWMutex
	systems       map[string]*SystemMetrics
	activeTimings map[string]time.Time
	historySize   int // Number of samples to keep
	enabled       bool
}

// NewCollector creates a new metrics collector
func NewCollector(historySize int) *Collector {
	return &Collector{
		systems:       make(map[string]*SystemMetrics),
		activeTimings: make(map[string]time.Time),
		historySize:   historySize,
		enabled:       true,
	}
}

// SetEnabled enables or disables metrics collection
func (c *Collector) SetEnabled(enabled bool) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.enabled = enabled
}

// IsEnabled returns whether metrics collection is enabled
func (c *Collector) IsEnabled() bool {
	c.mut.RLock()
	defer c.mut.RUnlock()
	return c.enabled
}

// SetHistorySize changes the history buffer size
func (c *Collector) SetHistorySize(size int) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.historySize = size

	// Resize existing buffers
	for _, metrics := range c.systems {
		if len(metrics.History) != size {
			newHistory := make([]float64, size)
			metrics.History = newHistory
			metrics.head = 0
			metrics.size = 0
		}
	}
}

// StartTiming begins timing a system
func (c *Collector) StartTiming(systemName string) {
	if !c.enabled {
		return
	}

	c.mut.Lock()
	defer c.mut.Unlock()

	c.activeTimings[systemName] = time.Now()
}

// EndTiming ends timing a system and records the duration
func (c *Collector) EndTiming(systemName string) {
	if !c.enabled {
		return
	}

	endTime := time.Now()

	c.mut.Lock()
	defer c.mut.Unlock()

	startTime, exists := c.activeTimings[systemName]
	if !exists {
		return
	}

	// Calculate duration in milliseconds
	duration := float64(endTime.Sub(startTime).Microseconds()) / 1000.0

	// Get or create system metrics
	metrics, exists := c.systems[systemName]
	if !exists {
		metrics = &SystemMetrics{
			Name:    systemName,
			History: make([]float64, c.historySize),
			head:    0,
			size:    0,
		}
		c.systems[systemName] = metrics
	}

	// Update current value
	metrics.CurrentValue = duration

	// Add to ring buffer
	metrics.History[metrics.head] = duration
	metrics.head = (metrics.head + 1) % c.historySize
	if metrics.size < c.historySize {
		metrics.size++
	}

	// Clean up active timing
	delete(c.activeTimings, systemName)
}

// RecordValue records a value directly without timing (useful for counts, etc.)
func (c *Collector) RecordValue(systemName string, value float64) {
	if !c.enabled {
		return
	}

	c.mut.Lock()
	defer c.mut.Unlock()

	// Get or create system metrics
	metrics, exists := c.systems[systemName]
	if !exists {
		metrics = &SystemMetrics{
			Name:    systemName,
			History: make([]float64, c.historySize),
			head:    0,
			size:    0,
		}
		c.systems[systemName] = metrics
	}

	// Update current value
	metrics.CurrentValue = value

	// Add to ring buffer
	metrics.History[metrics.head] = value
	metrics.head = (metrics.head + 1) % c.historySize
	if metrics.size < c.historySize {
		metrics.size++
	}
}

// GetMetrics returns a copy of metrics for a system
func (c *Collector) GetMetrics(systemName string) *SystemMetrics {
	c.mut.RLock()
	defer c.mut.RUnlock()

	metrics, exists := c.systems[systemName]
	if !exists {
		return nil
	}

	// Return a copy to avoid external mutations
	copy := &SystemMetrics{
		Name:         metrics.Name,
		CurrentValue: metrics.CurrentValue,
		History:      make([]float64, metrics.size),
		head:         0,
		size:         metrics.size,
	}

	// Copy history in chronological order (oldest to newest)
	for i := 0; i < metrics.size; i++ {
		srcIdx := (metrics.head - metrics.size + i + c.historySize) % c.historySize
		copy.History[i] = metrics.History[srcIdx]
	}

	return copy
}

// GetAllMetrics returns copies of all system metrics
func (c *Collector) GetAllMetrics() map[string]*SystemMetrics {
	c.mut.RLock()
	defer c.mut.RUnlock()

	result := make(map[string]*SystemMetrics, len(c.systems))

	for name, metrics := range c.systems {
		// Return a copy to avoid external mutations
		copy := &SystemMetrics{
			Name:         metrics.Name,
			CurrentValue: metrics.CurrentValue,
			History:      make([]float64, metrics.size),
			head:         0,
			size:         metrics.size,
		}

		// Copy history in chronological order (oldest to newest)
		for i := 0; i < metrics.size; i++ {
			srcIdx := (metrics.head - metrics.size + i + c.historySize) % c.historySize
			copy.History[i] = metrics.History[srcIdx]
		}

		result[name] = copy
	}

	return result
}

// Reset clears all metrics data
func (c *Collector) Reset() {
	c.mut.Lock()
	defer c.mut.Unlock()

	for _, metrics := range c.systems {
		metrics.CurrentValue = 0
		metrics.head = 0
		metrics.size = 0
		for i := range metrics.History {
			metrics.History[i] = 0
		}
	}
}

// GetSystemNames returns a list of all tracked system names
func (c *Collector) GetSystemNames() []string {
	c.mut.RLock()
	defer c.mut.RUnlock()

	names := make([]string, 0, len(c.systems))
	for name := range c.systems {
		names = append(names, name)
	}
	return names
}
