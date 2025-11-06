package network

import (
	"fmt"
	"net"
	"time"
)

// Connection represents a UDP connection to a Source Engine server
type Connection struct {
	conn       *net.UDPConn
	serverAddr *net.UDPAddr
	timeout    time.Duration
}

// NewConnection creates a new UDP connection to a server
// address should be in format "host:port" (e.g., "192.168.1.100:27015")
func NewConnection(address string, timeout time.Duration) (*Connection, error) {
	// Resolve server address
	serverAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address %s: %w", address, err)
	}

	// Create UDP connection (client doesn't need to bind to specific port)
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}

	return &Connection{
		conn:       conn,
		serverAddr: serverAddr,
		timeout:    timeout,
	}, nil
}

// Send sends raw bytes to the server
func (c *Connection) Send(data []byte) error {
	_, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send packet: %w", err)
	}
	return nil
}

// Receive receives raw bytes from the server
// Returns the received data or an error if timeout occurs
func (c *Connection) Receive() ([]byte, error) {
	// Set read deadline
	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Buffer for receiving (max UDP packet size)
	buffer := make([]byte, 1500)

	// Use Read() for connected UDP sockets (more idiomatic than ReadFromUDP)
	n, err := c.conn.Read(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("receive timeout after %v", c.timeout)
		}
		return nil, fmt.Errorf("failed to receive packet: %w", err)
	}

	return buffer[:n], nil
}

// SendAndReceive sends a packet and waits for a response
// This is a convenience method for simple request-response patterns
func (c *Connection) SendAndReceive(data []byte) ([]byte, error) {
	if err := c.Send(data); err != nil {
		return nil, err
	}
	return c.Receive()
}

// SetTimeout sets the receive timeout
func (c *Connection) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// ServerAddress returns the server's address
func (c *Connection) ServerAddress() *net.UDPAddr {
	return c.serverAddr
}

// Close closes the UDP connection
func (c *Connection) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
