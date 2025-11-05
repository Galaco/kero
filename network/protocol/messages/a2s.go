package messages

import (
	"fmt"

	"github.com/galaco/kero/network/protocol"
)

// A2S_INFO_PAYLOAD is the magic payload for A2S_INFO queries
// Changed in 2020 to "Source Engine Query\x00" to prevent reflection attacks
const A2S_INFO_PAYLOAD = "Source Engine Query\x00"

// ServerType indicates the server operating system
type ServerType byte

const (
	ServerTypeLinux   ServerType = 'l'
	ServerTypeWindows ServerType = 'w'
	ServerTypeMac     ServerType = 'm' // or 'o' for Mac
)

func (st ServerType) String() string {
	switch st {
	case ServerTypeLinux:
		return "Linux"
	case ServerTypeWindows:
		return "Windows"
	case ServerTypeMac:
		return "Mac"
	default:
		return fmt.Sprintf("Unknown(%c)", byte(st))
	}
}

// ServerEnvironment indicates the server visibility
type ServerEnvironment byte

const (
	ServerEnvDedicated ServerEnvironment = 'd'
	ServerEnvLocal     ServerEnvironment = 'l'
	ServerEnvSourceTV  ServerEnvironment = 'p'
)

func (se ServerEnvironment) String() string {
	switch se {
	case ServerEnvDedicated:
		return "Dedicated"
	case ServerEnvLocal:
		return "Local"
	case ServerEnvSourceTV:
		return "SourceTV"
	default:
		return fmt.Sprintf("Unknown(%c)", byte(se))
	}
}

// A2S_INFO_Request creates an A2S_INFO query packet
// This is sent from client to server to request server information
func A2S_INFO_Request() *protocol.ConnectionlessPacket {
	writer := protocol.NewPacketWriter()
	writer.WriteBytes([]byte(A2S_INFO_PAYLOAD))

	return protocol.NewConnectionlessPacket(protocol.A2S_INFO, writer.Bytes())
}

// A2S_INFO_Response represents the server's response to an A2S_INFO query
// This contains all the information displayed in the server browser
type A2S_INFO_Response struct {
	Protocol    byte              // Protocol version used by the server
	Name        string            // Server name
	Map         string            // Current map
	Folder      string            // Game directory (e.g., "cstrike", "tf")
	Game        string            // Game name (e.g., "Counter-Strike: Source")
	AppID       uint16            // Steam Application ID
	Players     byte              // Current player count
	MaxPlayers  byte              // Maximum player slots
	Bots        byte              // Number of bots
	ServerType  ServerType        // Server OS (l/w/m)
	Environment ServerEnvironment // Dedicated/Listen/SourceTV
	Visibility  bool              // Password protected (true = password required)
	VAC         bool              // VAC enabled

	// Optional fields (if present in response)
	Version   string  // Game version
	ExtraData byte    // EDF flags indicating which extra data is present
	Port      *uint16 // Server port (if EDF_PORT is set)
	SteamID   *uint64 // Server SteamID (if EDF_STEAMID is set)
	SourceTV  *struct {
		Port uint16
		Name string
	}                // SourceTV info (if EDF_SOURCETV is set)
	Keywords *string // Server tags/keywords (if EDF_KEYWORDS is set)
	GameID   *uint64 // Game ID (if EDF_GAMEID is set)
}

// Extra Data Flags (EDF)
const (
	EDF_PORT     byte = 0x80 // Server port included
	EDF_STEAMID  byte = 0x10 // Server SteamID included
	EDF_SOURCETV byte = 0x40 // SourceTV port and name included
	EDF_KEYWORDS byte = 0x20 // Server tags/keywords included
	EDF_GAMEID   byte = 0x01 // Game ID included
)

// ParseA2S_INFO_Response parses a server's response to an A2S_INFO query
func ParseA2S_INFO_Response(packet *protocol.ConnectionlessPacket) (*A2S_INFO_Response, error) {
	if packet.Type != protocol.S2A_INFO_SRC {
		return nil, fmt.Errorf("invalid packet type for A2S_INFO response: 0x%02X (expected 0x%02X)", packet.Type, protocol.S2A_INFO_SRC)
	}

	reader := protocol.NewPacketReader(packet.Data)

	info := &A2S_INFO_Response{}

	// Read protocol version
	var err error
	info.Protocol, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read protocol: %w", err)
	}

	// Read server name
	info.Name, err = reader.ReadString()
	if err != nil {
		return nil, fmt.Errorf("failed to read server name: %w", err)
	}

	// Read map name
	info.Map, err = reader.ReadString()
	if err != nil {
		return nil, fmt.Errorf("failed to read map name: %w", err)
	}

	// Read game folder
	info.Folder, err = reader.ReadString()
	if err != nil {
		return nil, fmt.Errorf("failed to read folder: %w", err)
	}

	// Read game name
	info.Game, err = reader.ReadString()
	if err != nil {
		return nil, fmt.Errorf("failed to read game name: %w", err)
	}

	// Read App ID
	info.AppID, err = reader.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf("failed to read AppID: %w", err)
	}

	// Read player counts
	info.Players, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read player count: %w", err)
	}

	info.MaxPlayers, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read max players: %w", err)
	}

	info.Bots, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read bot count: %w", err)
	}

	// Read server type
	serverTypeByte, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read server type: %w", err)
	}
	info.ServerType = ServerType(serverTypeByte)

	// Read environment
	envByte, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read environment: %w", err)
	}
	info.Environment = ServerEnvironment(envByte)

	// Read visibility (password)
	visibilityByte, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read visibility: %w", err)
	}
	info.Visibility = visibilityByte != 0

	// Read VAC status
	vacByte, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read VAC status: %w", err)
	}
	info.VAC = vacByte != 0

	// Read version string
	info.Version, err = reader.ReadString()
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	// Check for Extra Data Flag (EDF)
	if !reader.HasMore() {
		// No extra data
		return info, nil
	}

	info.ExtraData, err = reader.ReadByte()
	if err != nil {
		// Not an error, just no extra data
		return info, nil
	}

	// Parse extra data based on flags
	if info.ExtraData&EDF_PORT != 0 {
		port, err := reader.ReadUint16()
		if err != nil {
			return nil, fmt.Errorf("failed to read port: %w", err)
		}
		info.Port = &port
	}

	if info.ExtraData&EDF_STEAMID != 0 {
		steamIDBytes, err := reader.ReadBytes(8)
		if err != nil {
			return nil, fmt.Errorf("failed to read SteamID: %w", err)
		}
		// Parse 8 bytes as uint64 (little endian)
		steamID := uint64(steamIDBytes[0]) |
			uint64(steamIDBytes[1])<<8 |
			uint64(steamIDBytes[2])<<16 |
			uint64(steamIDBytes[3])<<24 |
			uint64(steamIDBytes[4])<<32 |
			uint64(steamIDBytes[5])<<40 |
			uint64(steamIDBytes[6])<<48 |
			uint64(steamIDBytes[7])<<56
		info.SteamID = &steamID
	}

	if info.ExtraData&EDF_SOURCETV != 0 {
		info.SourceTV = &struct {
			Port uint16
			Name string
		}{}
		info.SourceTV.Port, err = reader.ReadUint16()
		if err != nil {
			return nil, fmt.Errorf("failed to read SourceTV port: %w", err)
		}
		info.SourceTV.Name, err = reader.ReadString()
		if err != nil {
			return nil, fmt.Errorf("failed to read SourceTV name: %w", err)
		}
	}

	if info.ExtraData&EDF_KEYWORDS != 0 {
		keywords, err := reader.ReadString()
		if err != nil {
			return nil, fmt.Errorf("failed to read keywords: %w", err)
		}
		info.Keywords = &keywords
	}

	if info.ExtraData&EDF_GAMEID != 0 {
		gameIDBytes, err := reader.ReadBytes(8)
		if err != nil {
			return nil, fmt.Errorf("failed to read GameID: %w", err)
		}
		gameID := uint64(gameIDBytes[0]) |
			uint64(gameIDBytes[1])<<8 |
			uint64(gameIDBytes[2])<<16 |
			uint64(gameIDBytes[3])<<24 |
			uint64(gameIDBytes[4])<<32 |
			uint64(gameIDBytes[5])<<40 |
			uint64(gameIDBytes[6])<<48 |
			uint64(gameIDBytes[7])<<56
		info.GameID = &gameID
	}

	return info, nil
}

// String returns a human-readable representation of server info
func (info *A2S_INFO_Response) String() string {
	result := fmt.Sprintf("Server: %s\n", info.Name)
	result += fmt.Sprintf("  Map: %s\n", info.Map)
	result += fmt.Sprintf("  Game: %s (%s)\n", info.Game, info.Folder)
	result += fmt.Sprintf("  Players: %d/%d (%d bots)\n", info.Players, info.MaxPlayers, info.Bots)
	result += fmt.Sprintf("  Server Type: %s on %s\n", info.Environment, info.ServerType)
	result += fmt.Sprintf("  Password: %v | VAC: %v\n", info.Visibility, info.VAC)
	result += fmt.Sprintf("  Version: %s | AppID: %d\n", info.Version, info.AppID)

	if info.Port != nil {
		result += fmt.Sprintf("  Port: %d\n", *info.Port)
	}
	if info.SteamID != nil {
		result += fmt.Sprintf("  SteamID: %d\n", *info.SteamID)
	}
	if info.SourceTV != nil {
		result += fmt.Sprintf("  SourceTV: %s:%d\n", info.SourceTV.Name, info.SourceTV.Port)
	}
	if info.Keywords != nil {
		result += fmt.Sprintf("  Tags: %s\n", *info.Keywords)
	}

	return result
}

// A2S_PLAYER_Request creates an A2S_PLAYER query packet
// Challenge should be -1 for initial request, or the challenge from S2C_CHALLENGE response
func A2S_PLAYER_Request(challenge int32) *protocol.ConnectionlessPacket {
	writer := protocol.NewPacketWriter()
	writer.WriteInt32(challenge)

	return protocol.NewConnectionlessPacket(protocol.A2S_PLAYER, writer.Bytes())
}

// A2S_PLAYER_Response represents the server's player list
type A2S_PLAYER_Response struct {
	Count   byte
	Players []PlayerInfo
}

// PlayerInfo represents a single player on the server
type PlayerInfo struct {
	Index    byte    // Player index (0-255)
	Name     string  // Player name
	Score    int32   // Player score
	Duration float32 // Time connected (in seconds)
}

// ParseA2S_PLAYER_Response parses a server's response to an A2S_PLAYER query
func ParseA2S_PLAYER_Response(packet *protocol.ConnectionlessPacket) (*A2S_PLAYER_Response, error) {
	if packet.Type != protocol.S2A_PLAYER {
		return nil, fmt.Errorf("invalid packet type for A2S_PLAYER response: 0x%02X", packet.Type)
	}

	reader := protocol.NewPacketReader(packet.Data)

	response := &A2S_PLAYER_Response{}

	// Read player count
	var err error
	response.Count, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read player count: %w", err)
	}

	// Read each player
	response.Players = make([]PlayerInfo, 0, response.Count)
	for i := byte(0); i < response.Count; i++ {
		player := PlayerInfo{}

		player.Index, err = reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read player %d index: %w", i, err)
		}

		player.Name, err = reader.ReadString()
		if err != nil {
			return nil, fmt.Errorf("failed to read player %d name: %w", i, err)
		}

		player.Score, err = reader.ReadInt32()
		if err != nil {
			return nil, fmt.Errorf("failed to read player %d score: %w", i, err)
		}

		player.Duration, err = reader.ReadFloat32()
		if err != nil {
			return nil, fmt.Errorf("failed to read player %d duration: %w", i, err)
		}

		response.Players = append(response.Players, player)
	}

	return response, nil
}

// A2S_RULES_Request creates an A2S_RULES query packet
// Challenge should be -1 for initial request, or the challenge from S2C_CHALLENGE response
func A2S_RULES_Request(challenge int32) *protocol.ConnectionlessPacket {
	writer := protocol.NewPacketWriter()
	writer.WriteInt32(challenge)

	return protocol.NewConnectionlessPacket(protocol.A2S_RULES, writer.Bytes())
}

// A2S_RULES_Response represents the server's rules/cvars
type A2S_RULES_Response struct {
	Count int16
	Rules map[string]string
}

// ParseA2S_RULES_Response parses a server's response to an A2S_RULES query
func ParseA2S_RULES_Response(packet *protocol.ConnectionlessPacket) (*A2S_RULES_Response, error) {
	if packet.Type != protocol.S2A_RULES {
		return nil, fmt.Errorf("invalid packet type for A2S_RULES response: 0x%02X", packet.Type)
	}

	reader := protocol.NewPacketReader(packet.Data)

	response := &A2S_RULES_Response{
		Rules: make(map[string]string),
	}

	// Read rule count
	var err error
	response.Count, err = reader.ReadInt16()
	if err != nil {
		return nil, fmt.Errorf("failed to read rule count: %w", err)
	}

	// Read each rule
	for i := int16(0); i < response.Count; i++ {
		name, err := reader.ReadString()
		if err != nil {
			return nil, fmt.Errorf("failed to read rule %d name: %w", i, err)
		}

		value, err := reader.ReadString()
		if err != nil {
			return nil, fmt.Errorf("failed to read rule %d value: %w", i, err)
		}

		response.Rules[name] = value
	}

	return response, nil
}

// S2C_CHALLENGE_Response represents a challenge response from the server
// Used for A2S_PLAYER and A2S_RULES queries to prevent spoofing
type S2C_CHALLENGE_Response struct {
	Challenge int32
}

// ParseS2C_CHALLENGE_Response parses a challenge response
func ParseS2C_CHALLENGE_Response(packet *protocol.ConnectionlessPacket) (*S2C_CHALLENGE_Response, error) {
	if packet.Type != protocol.S2C_CHALLENGE {
		return nil, fmt.Errorf("invalid packet type for S2C_CHALLENGE response: 0x%02X", packet.Type)
	}

	reader := protocol.NewPacketReader(packet.Data)

	response := &S2C_CHALLENGE_Response{}

	var err error
	response.Challenge, err = reader.ReadInt32()
	if err != nil {
		return nil, fmt.Errorf("failed to read challenge: %w", err)
	}

	return response, nil
}
