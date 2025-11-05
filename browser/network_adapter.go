package browser

import (
	"fmt"
	"time"

	"github.com/galaco/kero/network"
	"github.com/galaco/kero/network/protocol/messages"
)

// NetworkClient interface defines the contract for querying servers
// This allows the browser to be tested without real network calls
type NetworkClient interface {
	QueryServerInfo(address string) (*ServerInfo, error)
}

// RealNetworkAdapter implements NetworkClient using the actual network package
// This is the production adapter that makes real UDP queries
type RealNetworkAdapter struct {
	timeout time.Duration
}

// NewRealNetworkAdapter creates a new adapter with default timeout
func NewRealNetworkAdapter() *RealNetworkAdapter {
	return &RealNetworkAdapter{
		timeout: 5 * time.Second,
	}
}

// NewRealNetworkAdapterWithTimeout creates a new adapter with custom timeout
func NewRealNetworkAdapterWithTimeout(timeout time.Duration) *RealNetworkAdapter {
	return &RealNetworkAdapter{
		timeout: timeout,
	}
}

// QueryServerInfo queries a server and converts the response to browser.ServerInfo
func (a *RealNetworkAdapter) QueryServerInfo(address string) (*ServerInfo, error) {
	// Create network client
	client := network.NewClient()
	client.SetTimeout(a.timeout)

	// Connect to server
	if err := client.Connect(address); err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer client.Disconnect()

	// Measure ping
	startTime := time.Now()

	// Query server info
	networkInfo, err := client.QueryServerInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to query %s: %w", address, err)
	}

	// Calculate ping
	ping := int(time.Since(startTime).Milliseconds())

	// Convert network response to browser domain model
	browserInfo := convertToBrowserInfo(address, networkInfo, ping)

	return browserInfo, nil
}

// convertToBrowserInfo converts network.A2S_INFO_Response to browser.ServerInfo
// This is the adapter between the network layer and browser layer
func convertToBrowserInfo(address string, networkInfo *messages.A2S_INFO_Response, ping int) *ServerInfo {
	info := &ServerInfo{
		Address:     address,
		Name:        networkInfo.Name,
		Map:         networkInfo.Map,
		Game:        networkInfo.Game,
		Folder:      networkInfo.Folder,
		Version:     networkInfo.Version,
		Players:     int(networkInfo.Players),
		MaxPlayers:  int(networkInfo.MaxPlayers),
		Bots:        int(networkInfo.Bots),
		Password:    networkInfo.Visibility,
		VAC:         networkInfo.VAC,
		ServerType:  networkInfo.ServerType.String(),
		Environment: networkInfo.Environment.String(),
		Ping:        ping,
		AppID:       networkInfo.AppID,
		QueriedAt:   time.Now(),
	}

	// Extract tags if present
	if networkInfo.Keywords != nil {
		info.Tags = *networkInfo.Keywords
	}

	return info
}

// MockNetworkAdapter is a mock implementation for testing
// It returns predefined responses without making real network calls
type MockNetworkAdapter struct {
	// Response to return (can be set by tests)
	MockResponse *ServerInfo
	MockError    error

	// Track calls for verification
	LastQueriedAddress string
	CallCount          int
}

// NewMockNetworkAdapter creates a new mock adapter
func NewMockNetworkAdapter() *MockNetworkAdapter {
	return &MockNetworkAdapter{
		MockResponse: nil,
		MockError:    nil,
		CallCount:    0,
	}
}

// QueryServerInfo returns the mock response or error
func (m *MockNetworkAdapter) QueryServerInfo(address string) (*ServerInfo, error) {
	m.LastQueriedAddress = address
	m.CallCount++

	if m.MockError != nil {
		return nil, m.MockError
	}

	if m.MockResponse != nil {
		// Clone the response and update address
		response := *m.MockResponse
		response.Address = address
		return &response, nil
	}

	// Default mock response
	return &ServerInfo{
		Address:     address,
		Name:        "Mock Server",
		Map:         "de_dust2",
		Game:        "Counter-Strike: Source",
		Folder:      "cstrike",
		Version:     "6153",
		Players:     5,
		MaxPlayers:  16,
		Bots:        0,
		Password:    false,
		VAC:         true,
		ServerType:  "Linux",
		Environment: "Dedicated",
		Ping:        25,
		AppID:       240,
		QueriedAt:   time.Now(),
	}, nil
}

// SetMockResponse sets the response that will be returned
func (m *MockNetworkAdapter) SetMockResponse(response *ServerInfo) {
	m.MockResponse = response
}

// SetMockError sets an error to be returned instead of a response
func (m *MockNetworkAdapter) SetMockError(err error) {
	m.MockError = err
}

// Reset clears call tracking
func (m *MockNetworkAdapter) Reset() {
	m.LastQueriedAddress = ""
	m.CallCount = 0
}
