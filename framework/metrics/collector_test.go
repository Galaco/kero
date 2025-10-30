package metrics

import (
	"testing"
	"time"
)

func TestCollectorBasicTiming(t *testing.T) {
	collector := NewCollector(10)

	collector.StartTiming("test_system")
	time.Sleep(1 * time.Millisecond)
	collector.EndTiming("test_system")

	metrics := collector.GetMetrics("test_system")
	if metrics == nil {
		t.Fatal("Expected metrics to exist")
	}

	if metrics.CurrentValue < 1.0 {
		t.Errorf("Expected at least 1ms, got %.2f", metrics.CurrentValue)
	}

	if metrics.size != 1 {
		t.Errorf("Expected size 1, got %d", metrics.size)
	}
}

func TestCollectorRingBuffer(t *testing.T) {
	historySize := 5
	collector := NewCollector(historySize)

	// Add more samples than buffer size
	for i := 0; i < 10; i++ {
		collector.RecordValue("test_system", float64(i))
	}

	metrics := collector.GetMetrics("test_system")
	if metrics == nil {
		t.Fatal("Expected metrics to exist")
	}

	// Should only have the last 5 values
	if metrics.size != historySize {
		t.Errorf("Expected size %d, got %d", historySize, metrics.size)
	}

	// History should contain values 5-9 in order
	expectedValues := []float64{5, 6, 7, 8, 9}
	for i, expected := range expectedValues {
		if metrics.History[i] != expected {
			t.Errorf("History[%d]: expected %.0f, got %.0f", i, expected, metrics.History[i])
		}
	}
}

func TestCollectorDisabled(t *testing.T) {
	collector := NewCollector(10)
	collector.SetEnabled(false)

	collector.StartTiming("test_system")
	time.Sleep(1 * time.Millisecond)
	collector.EndTiming("test_system")

	metrics := collector.GetMetrics("test_system")
	if metrics != nil {
		t.Error("Expected no metrics when disabled")
	}
}

func TestCollectorMultipleSystems(t *testing.T) {
	collector := NewCollector(10)

	systems := []string{"input", "physics", "render"}
	for _, system := range systems {
		collector.RecordValue(system, float64(len(system)))
	}

	allMetrics := collector.GetAllMetrics()
	if len(allMetrics) != len(systems) {
		t.Errorf("Expected %d systems, got %d", len(systems), len(allMetrics))
	}

	for _, system := range systems {
		if _, exists := allMetrics[system]; !exists {
			t.Errorf("Expected system %s to exist", system)
		}
	}
}

func TestCollectorReset(t *testing.T) {
	collector := NewCollector(10)

	collector.RecordValue("test_system", 42.0)
	collector.Reset()

	metrics := collector.GetMetrics("test_system")
	if metrics == nil {
		t.Fatal("Expected metrics to exist")
	}

	if metrics.CurrentValue != 0 {
		t.Errorf("Expected current value 0 after reset, got %.2f", metrics.CurrentValue)
	}

	if metrics.size != 0 {
		t.Errorf("Expected size 0 after reset, got %d", metrics.size)
	}
}

func TestCollectorResizeHistory(t *testing.T) {
	collector := NewCollector(5)

	// Add 5 samples
	for i := 0; i < 5; i++ {
		collector.RecordValue("test_system", float64(i))
	}

	// Resize to larger
	collector.SetHistorySize(10)
	metrics := collector.GetMetrics("test_system")

	// After resize, history should be empty (size reset)
	if metrics.size != 0 {
		t.Errorf("Expected size 0 after resize, got %d", metrics.size)
	}
}
