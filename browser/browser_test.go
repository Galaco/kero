package browser_test

import (
	"errors"
	"testing"
	"time"

	"github.com/galaco/kero/browser"
	"github.com/galaco/kero/messages"
)

// MockEventDispatcher is a test double for EventDispatcher
type MockEventDispatcher struct {
	dispatchedEvents []interface{}
}

func NewMockEventDispatcher() *MockEventDispatcher {
	return &MockEventDispatcher{
		dispatchedEvents: make([]interface{}, 0),
	}
}

func (m *MockEventDispatcher) DispatchTyped(event interface{}) {
	m.dispatchedEvents = append(m.dispatchedEvents, event)
}

func (m *MockEventDispatcher) GetEvents() []interface{} {
	return m.dispatchedEvents
}

func (m *MockEventDispatcher) GetEventCount() int {
	return len(m.dispatchedEvents)
}

func (m *MockEventDispatcher) Reset() {
	m.dispatchedEvents = make([]interface{}, 0)
}

// TestServerBrowser_QuerySuccess tests a successful query
func TestServerBrowser_QuerySuccess(t *testing.T) {
	// Arrange
	mockNetwork := browser.NewMockNetworkAdapter()
	mockEvents := NewMockEventDispatcher()
	b := browser.NewServerBrowser(mockNetwork, mockEvents)

	address := "192.168.1.100:27015"

	// Act
	err := b.Query(address)
	if err != nil {
		t.Fatalf("Query() returned error: %v", err)
	}

	// Wait for async query to complete
	time.Sleep(100 * time.Millisecond)

	// Assert - Check state
	if b.IsQuerying() {
		t.Error("Expected query to be complete")
	}

	result, queryErr := b.GetLastResult()
	if queryErr != nil {
		t.Errorf("Expected successful query, got error: %v", queryErr)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Address != address {
		t.Errorf("Expected address %s, got %s", address, result.Address)
	}

	// Assert - Check events
	if mockEvents.GetEventCount() != 2 {
		t.Errorf("Expected 2 events (started, completed), got %d", mockEvents.GetEventCount())
	}

	// Verify event order
	events := mockEvents.GetEvents()
	if _, ok := events[0].(messages.BrowserQueryStartedEvent); !ok {
		t.Error("First event should be BrowserQueryStartedEvent")
	}
	if _, ok := events[1].(messages.BrowserQueryCompletedEvent); !ok {
		t.Error("Second event should be BrowserQueryCompletedEvent")
	}
}

// TestServerBrowser_QueryFailure tests a failed query
func TestServerBrowser_QueryFailure(t *testing.T) {
	// Arrange
	mockNetwork := browser.NewMockNetworkAdapter()
	mockNetwork.SetMockError(errors.New("connection timeout"))
	mockEvents := NewMockEventDispatcher()
	b := browser.NewServerBrowser(mockNetwork, mockEvents)

	address := "invalid:99999"

	// Act
	err := b.Query(address)
	if err != nil {
		t.Fatalf("Query() returned error: %v", err)
	}

	// Wait for async query to complete
	time.Sleep(100 * time.Millisecond)

	// Assert - Check state
	if b.IsQuerying() {
		t.Error("Expected query to be complete")
	}

	result, queryErr := b.GetLastResult()
	if queryErr == nil {
		t.Error("Expected error, got nil")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}

	// Assert - Check events
	if mockEvents.GetEventCount() != 2 {
		t.Errorf("Expected 2 events (started, failed), got %d", mockEvents.GetEventCount())
	}

	// Verify event types
	events := mockEvents.GetEvents()
	if _, ok := events[0].(messages.BrowserQueryStartedEvent); !ok {
		t.Error("First event should be BrowserQueryStartedEvent")
	}
	if _, ok := events[1].(messages.BrowserQueryFailedEvent); !ok {
		t.Error("Second event should be BrowserQueryFailedEvent")
	}
}

// TestServerBrowser_QueryEmptyAddress tests validation
func TestServerBrowser_QueryEmptyAddress(t *testing.T) {
	// Arrange
	mockNetwork := browser.NewMockNetworkAdapter()
	mockEvents := NewMockEventDispatcher()
	b := browser.NewServerBrowser(mockNetwork, mockEvents)

	// Act
	err := b.Query("")

	// Assert
	if err == nil {
		t.Error("Expected error for empty address")
	}
	if b.IsQuerying() {
		t.Error("Query should not have started")
	}
	if mockEvents.GetEventCount() != 0 {
		t.Error("No events should be dispatched for invalid input")
	}
}

// TestServerBrowser_QueryWhileInProgress tests concurrent query prevention
func TestServerBrowser_QueryWhileInProgress(t *testing.T) {
	// Arrange
	mockNetwork := browser.NewMockNetworkAdapter()
	mockEvents := NewMockEventDispatcher()
	b := browser.NewServerBrowser(mockNetwork, mockEvents)

	// Act - Start first query
	err1 := b.Query("server1:27015")
	if err1 != nil {
		t.Fatalf("First query failed: %v", err1)
	}

	// Immediately try second query (first is still in progress)
	err2 := b.Query("server2:27015")

	// Assert
	if err2 == nil {
		t.Error("Expected error when querying while another query is in progress")
	}

	// Wait for first query to complete
	time.Sleep(100 * time.Millisecond)

	// Should be able to query again now
	err3 := b.Query("server3:27015")
	if err3 != nil {
		t.Errorf("Third query should succeed, got error: %v", err3)
	}

	// Wait for third query to complete
	time.Sleep(100 * time.Millisecond)
}

// TestServerBrowser_GetQueryHistory tests history tracking
func TestServerBrowser_GetQueryHistory(t *testing.T) {
	// Arrange
	mockNetwork := browser.NewMockNetworkAdapter()
	mockEvents := NewMockEventDispatcher()
	b := browser.NewServerBrowser(mockNetwork, mockEvents)

	// Act - Perform multiple queries
	addresses := []string{
		"server1:27015",
		"server2:27015",
		"server3:27015",
	}

	for _, addr := range addresses {
		err := b.Query(addr)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		time.Sleep(100 * time.Millisecond) // Wait for completion
	}

	// Assert
	history := b.GetQueryHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 history entries, got %d", len(history))
	}

	// Verify addresses match
	for i, addr := range addresses {
		if history[i].Address != addr {
			t.Errorf("History[%d]: expected %s, got %s", i, addr, history[i].Address)
		}
	}
}

// TestServerBrowser_ClearHistory tests history clearing
func TestServerBrowser_ClearHistory(t *testing.T) {
	// Arrange
	mockNetwork := browser.NewMockNetworkAdapter()
	mockEvents := NewMockEventDispatcher()
	b := browser.NewServerBrowser(mockNetwork, mockEvents)

	// Add some history
	b.Query("server1:27015")
	time.Sleep(100 * time.Millisecond)

	// Act
	b.ClearHistory()

	// Assert
	history := b.GetQueryHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history after clear, got %d entries", len(history))
	}
}

// TestServerBrowser_ThreadSafety tests concurrent access
func TestServerBrowser_ThreadSafety(t *testing.T) {
	// Arrange
	mockNetwork := browser.NewMockNetworkAdapter()
	mockEvents := NewMockEventDispatcher()
	b := browser.NewServerBrowser(mockNetwork, mockEvents)

	// Act - Access from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			// These should not race
			_ = b.IsQuerying()
			_ = b.GetCurrentAddress()
			_, _ = b.GetLastResult()
			_ = b.GetQueryHistory()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without a race condition, test passes
}

// TestMockNetworkAdapter tests the mock adapter
func TestMockNetworkAdapter(t *testing.T) {
	// Arrange
	mock := browser.NewMockNetworkAdapter()

	// Act - Query with default response
	info1, err1 := mock.QueryServerInfo("test:27015")

	// Assert - Default response
	if err1 != nil {
		t.Errorf("Expected success, got error: %v", err1)
	}
	if info1.Name != "Mock Server" {
		t.Errorf("Expected 'Mock Server', got '%s'", info1.Name)
	}
	if mock.CallCount != 1 {
		t.Errorf("Expected 1 call, got %d", mock.CallCount)
	}

	// Act - Set custom error
	customError := errors.New("custom error")
	mock.SetMockError(customError)
	info2, err2 := mock.QueryServerInfo("test:27015")

	// Assert - Custom error
	if err2 != customError {
		t.Errorf("Expected custom error, got %v", err2)
	}
	if info2 != nil {
		t.Error("Expected nil result on error")
	}
	if mock.CallCount != 2 {
		t.Errorf("Expected 2 calls, got %d", mock.CallCount)
	}
}
