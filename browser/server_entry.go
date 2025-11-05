package browser

import (
	"fmt"
	"time"
)

// ServerInfo represents information about a Source Engine game server
// This is the browser's domain model, decoupled from the network layer
type ServerInfo struct {
	// Connection
	Address string // "192.168.1.100:27015"

	// Server identity
	Name    string // Server name
	Map     string // Current map
	Game    string // Game name (e.g., "Counter-Strike: Source")
	Folder  string // Game directory (e.g., "cstrike")
	Version string // Game version

	// Player information
	Players    int // Current player count
	MaxPlayers int // Maximum player slots
	Bots       int // Number of bots

	// Server properties
	Password bool // Password protected
	VAC      bool // VAC secured

	// Server type
	ServerType  string // "Linux", "Windows", "Mac"
	Environment string // "Dedicated", "Listen", "SourceTV"

	// Performance
	Ping int // Latency in milliseconds (-1 if not measured)

	// Metadata
	AppID      uint16 // Steam Application ID
	Tags       string // Server keywords/tags
	QueriedAt  time.Time // When this info was retrieved
}

// FormatForDisplay returns a human-readable multi-line representation
// suitable for displaying in a text box or console
func (s *ServerInfo) FormatForDisplay() string {
	result := ""

	// Server header
	result += fmt.Sprintf("═══════════════════════════════════════════════\n")
	result += fmt.Sprintf("  %s\n", s.Name)
	result += fmt.Sprintf("═══════════════════════════════════════════════\n\n")

	// Connection info
	result += fmt.Sprintf("Address:  %s\n", s.Address)
	if s.Ping >= 0 {
		result += fmt.Sprintf("Ping:     %d ms\n", s.Ping)
	}
	result += fmt.Sprintf("\n")

	// Game info
	result += fmt.Sprintf("Game:     %s\n", s.Game)
	result += fmt.Sprintf("Map:      %s\n", s.Map)
	result += fmt.Sprintf("Version:  %s\n", s.Version)
	result += fmt.Sprintf("\n")

	// Players
	result += fmt.Sprintf("Players:  %d / %d", s.Players, s.MaxPlayers)
	if s.Bots > 0 {
		result += fmt.Sprintf(" (%d bots)", s.Bots)
	}
	result += fmt.Sprintf("\n\n")

	// Server properties
	result += fmt.Sprintf("Type:     %s %s\n", s.Environment, s.ServerType)
	result += fmt.Sprintf("Password: %s\n", boolToYesNo(s.Password))
	result += fmt.Sprintf("VAC:      %s\n", boolToYesNo(s.VAC))

	if s.Tags != "" {
		result += fmt.Sprintf("\nTags:     %s\n", s.Tags)
	}

	// Footer
	result += fmt.Sprintf("\n")
	result += fmt.Sprintf("Queried:  %s\n", s.QueriedAt.Format("15:04:05"))

	return result
}

// FormatCompact returns a single-line representation suitable for list views
func (s *ServerInfo) FormatCompact() string {
	passwordIcon := ""
	if s.Password {
		passwordIcon = "[P] "
	}
	vacIcon := ""
	if s.VAC {
		vacIcon = "[VAC] "
	}

	return fmt.Sprintf("%s%s%s | %s | %d/%d | %dms",
		passwordIcon, vacIcon, s.Name, s.Map, s.Players, s.MaxPlayers, s.Ping)
}

// IsFull returns true if the server is at maximum capacity
func (s *ServerInfo) IsFull() bool {
	return s.Players >= s.MaxPlayers
}

// IsEmpty returns true if the server has no players
func (s *ServerInfo) IsEmpty() bool {
	return s.Players == 0
}

// HasPlayers returns true if the server has any non-bot players
func (s *ServerInfo) HasPlayers() bool {
	return (s.Players - s.Bots) > 0
}

// boolToYesNo converts a boolean to "Yes" or "No"
func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// NewServerInfo creates a new ServerInfo with default values
func NewServerInfo(address string) *ServerInfo {
	return &ServerInfo{
		Address:    address,
		Name:       "Unknown",
		Map:        "Unknown",
		Game:       "Unknown",
		Folder:     "",
		Version:    "",
		Players:    0,
		MaxPlayers: 0,
		Bots:       0,
		Password:   false,
		VAC:        false,
		ServerType: "Unknown",
		Environment: "Unknown",
		Ping:       -1,
		AppID:      0,
		Tags:       "",
		QueriedAt:  time.Now(),
	}
}
