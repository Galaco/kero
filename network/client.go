package network

import (
	"fmt"
	"time"

	"github.com/galaco/kero/network/protocol"
	"github.com/galaco/kero/network/protocol/messages"
)

// Client represents a Source Engine network client
// Currently supports server queries (A2S_INFO, A2S_PLAYER, A2S_RULES)
// Future: Will support full client connection for multiplayer
type Client struct {
	connection *Connection
	timeout    time.Duration
}

// NewClient creates a new Source Engine network client
func NewClient() *Client {
	return &Client{
		timeout: 5 * time.Second, // Default 5 second timeout
	}
}

// SetTimeout sets the timeout for network operations
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Connect establishes a connection to a server
// address should be in format "host:port" (e.g., "192.168.1.100:27015")
func (c *Client) Connect(address string) error {
	conn, err := NewConnection(address, c.timeout)
	if err != nil {
		return err
	}
	c.connection = conn
	return nil
}

// Disconnect closes the connection to the server
func (c *Client) Disconnect() error {
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

// IsConnected returns true if connected to a server
func (c *Client) IsConnected() bool {
	return c.connection != nil
}

// QueryServerInfo queries server information using A2S_INFO
// This returns the information you see in the server browser
func (c *Client) QueryServerInfo() (*messages.A2S_INFO_Response, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to a server")
	}

	// Create A2S_INFO request
	request := messages.A2S_INFO_Request()

	// Send request and receive response
	responseData, err := c.connection.SendAndReceive(request.Serialize())
	if err != nil {
		return nil, fmt.Errorf("failed to query server info: %w", err)
	}

	// Parse connectionless packet
	packet, err := protocol.ParseConnectionless(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse A2S_INFO response
	info, err := messages.ParseA2S_INFO_Response(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server info: %w", err)
	}

	return info, nil
}

// QueryPlayers queries the player list using A2S_PLAYER
// Note: This requires a challenge-response mechanism
func (c *Client) QueryPlayers() (*messages.A2S_PLAYER_Response, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to a server")
	}

	// Step 1: Request with challenge -1
	request := messages.A2S_PLAYER_Request(-1)
	responseData, err := c.connection.SendAndReceive(request.Serialize())
	if err != nil {
		return nil, fmt.Errorf("failed to send player query: %w", err)
	}

	// Parse response
	packet, err := protocol.ParseConnectionless(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Step 2: If server responds with challenge, resend with challenge value
	if packet.Type == protocol.S2C_CHALLENGE {
		challengeResponse, err := messages.ParseS2C_CHALLENGE_Response(packet)
		if err != nil {
			return nil, fmt.Errorf("failed to parse challenge: %w", err)
		}

		// Resend with proper challenge
		request = messages.A2S_PLAYER_Request(challengeResponse.Challenge)
		responseData, err = c.connection.SendAndReceive(request.Serialize())
		if err != nil {
			return nil, fmt.Errorf("failed to send player query with challenge: %w", err)
		}

		// Parse again
		packet, err = protocol.ParseConnectionless(responseData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse player response: %w", err)
		}
	}

	// Parse player list
	players, err := messages.ParseA2S_PLAYER_Response(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse player list: %w", err)
	}

	return players, nil
}

// QueryRules queries server rules/cvars using A2S_RULES
// Note: This requires a challenge-response mechanism
func (c *Client) QueryRules() (*messages.A2S_RULES_Response, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to a server")
	}

	// Step 1: Request with challenge -1
	request := messages.A2S_RULES_Request(-1)
	responseData, err := c.connection.SendAndReceive(request.Serialize())
	if err != nil {
		return nil, fmt.Errorf("failed to send rules query: %w", err)
	}

	// Parse response
	packet, err := protocol.ParseConnectionless(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Step 2: If server responds with challenge, resend with challenge value
	if packet.Type == protocol.S2C_CHALLENGE {
		challengeResponse, err := messages.ParseS2C_CHALLENGE_Response(packet)
		if err != nil {
			return nil, fmt.Errorf("failed to parse challenge: %w", err)
		}

		// Resend with proper challenge
		request = messages.A2S_RULES_Request(challengeResponse.Challenge)
		responseData, err = c.connection.SendAndReceive(request.Serialize())
		if err != nil {
			return nil, fmt.Errorf("failed to send rules query with challenge: %w", err)
		}

		// Parse again
		packet, err = protocol.ParseConnectionless(responseData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rules response: %w", err)
		}
	}

	// Parse rules
	rules, err := messages.ParseA2S_RULES_Response(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}

	return rules, nil
}

// QuickQuery is a convenience method that connects, queries info, and disconnects
// Useful for server browser functionality
func QuickQueryServerInfo(address string) (*messages.A2S_INFO_Response, error) {
	client := NewClient()
	if err := client.Connect(address); err != nil {
		return nil, err
	}
	defer client.Disconnect()

	return client.QueryServerInfo()
}
