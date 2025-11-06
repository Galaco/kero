package browser

import (
	"fmt"
	"sync"

	"github.com/galaco/kero/messages"
)

// EventDispatcher interface defines how the browser communicates state changes
// This decouples the browser from the event system implementation
type EventDispatcher interface {
	// DispatchTyped sends a typed event to all registered listeners
	DispatchTyped(event interface{})
}

// ServerBrowser manages server browser state and operations
// It provides thread-safe access to query results and orchestrates async queries
type ServerBrowser struct {
	mu sync.RWMutex

	// Current query state
	queryInProgress bool
	currentAddress  string

	// Results storage
	lastResult *ServerInfo
	lastError  error

	// Query history (for future features like recent servers)
	queryHistory []*ServerInfo

	// Dependencies (injected for testability)
	networkClient NetworkClient
	eventBus      EventDispatcher
}

// NewServerBrowser creates a new server browser with the given dependencies
func NewServerBrowser(networkClient NetworkClient, eventBus EventDispatcher) *ServerBrowser {
	return &ServerBrowser{
		networkClient: networkClient,
		eventBus:      eventBus,
		queryHistory:  make([]*ServerInfo, 0),
	}
}

// Query initiates an asynchronous server query
// This method returns immediately and dispatches events when the query completes
func (b *ServerBrowser) Query(address string) error {
	// Validate input
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	b.mu.Lock()
	// Check if a query is already in progress
	if b.queryInProgress {
		b.mu.Unlock()
		return fmt.Errorf("query already in progress for %s", b.currentAddress)
	}

	// Mark query as started
	b.queryInProgress = true
	b.currentAddress = address

	// Clear previous results
	b.lastResult = nil
	b.lastError = nil
	b.mu.Unlock()

	// Dispatch query started event
	if b.eventBus != nil {
		b.eventBus.DispatchTyped(messages.BrowserQueryStartedEvent{Address: address})
	}

	// Execute query asynchronously
	go b.executeQuery(address)

	return nil
}

// executeQuery performs the actual query in a goroutine
// This is internal and should not be called directly
func (b *ServerBrowser) executeQuery(address string) {
	// Query the server
	info, err := b.networkClient.QueryServerInfo(address)

	// Update state (thread-safe)
	b.mu.Lock()
	b.queryInProgress = false
	b.currentAddress = ""
	b.lastResult = info
	b.lastError = err

	// Add to history if successful
	if err == nil && info != nil {
		b.addToHistory(info)
	}
	b.mu.Unlock()

	// Dispatch appropriate event
	if b.eventBus != nil {
		if err != nil {
			b.eventBus.DispatchTyped(messages.BrowserQueryFailedEvent{
				Address: address,
			})
		} else {
			b.eventBus.DispatchTyped(messages.BrowserQueryCompletedEvent{
				Address: address,
			})
		}
	}
}

// addToHistory adds a query result to the history
// Must be called with lock held
func (b *ServerBrowser) addToHistory(info *ServerInfo) {
	// Limit history to 100 entries
	const maxHistory = 100

	b.queryHistory = append(b.queryHistory, info)
	if len(b.queryHistory) > maxHistory {
		// Remove oldest entry
		b.queryHistory = b.queryHistory[1:]
	}
}

// GetLastResult returns the most recent query result
// Returns (result, error) - one will be nil depending on whether query succeeded
// Thread-safe: can be called from any goroutine
func (b *ServerBrowser) GetLastResult() (*ServerInfo, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastResult, b.lastError
}

// IsQuerying returns true if a query is currently in progress
// Thread-safe: can be called from any goroutine
func (b *ServerBrowser) IsQuerying() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.queryInProgress
}

// GetCurrentAddress returns the address being queried (if any)
// Returns empty string if no query is in progress
// Thread-safe: can be called from any goroutine
func (b *ServerBrowser) GetCurrentAddress() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.currentAddress
}

// GetQueryHistory returns a copy of the query history
// Thread-safe: can be called from any goroutine
func (b *ServerBrowser) GetQueryHistory() []*ServerInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return a copy to prevent external modification
	historyCopy := make([]*ServerInfo, len(b.queryHistory))
	copy(historyCopy, b.queryHistory)
	return historyCopy
}

// ClearHistory clears the query history
func (b *ServerBrowser) ClearHistory() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.queryHistory = make([]*ServerInfo, 0)
}
