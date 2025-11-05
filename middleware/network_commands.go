package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/network"
)

// AddNetworkCommands registers network-related console commands
func AddNetworkCommands() {
	// net_query command - Query a server for information
	console.AddCommand("net_query", "Query a Source Engine server for information", "net_query <address:port>", func(options string) error {
		if options == "" {
			console.PrintString(console.LevelWarning, "Usage: net_query <address:port>")
			console.PrintString(console.LevelInfo, "Example: net_query 192.168.1.100:27015")
			return nil
		}

		address := strings.TrimSpace(options)

		console.PrintString(console.LevelInfo, fmt.Sprintf("Querying server: %s", address))

		// Query server info
		info, err := network.QuickQueryServerInfo(address)
		if err != nil {
			console.PrintString(console.LevelError, fmt.Sprintf("Failed to query server: %v", err))
			return err
		}

		// Display server information
		console.PrintString(console.LevelSuccess, "Server Information:")
		console.PrintString(console.LevelInfo, fmt.Sprintf("  Name: %s", info.Name))
		console.PrintString(console.LevelInfo, fmt.Sprintf("  Map: %s", info.Map))
		console.PrintString(console.LevelInfo, fmt.Sprintf("  Game: %s (%s)", info.Game, info.Folder))
		console.PrintString(console.LevelInfo, fmt.Sprintf("  Players: %d/%d (%d bots)", info.Players, info.MaxPlayers, info.Bots))
		console.PrintString(console.LevelInfo, fmt.Sprintf("  Server Type: %s on %s", info.Environment, info.ServerType))
		console.PrintString(console.LevelInfo, fmt.Sprintf("  Password: %v | VAC: %v", info.Visibility, info.VAC))
		console.PrintString(console.LevelInfo, fmt.Sprintf("  Version: %s | AppID: %d | Protocol: %d", info.Version, info.AppID, info.Protocol))

		if info.Port != nil {
			console.PrintString(console.LevelInfo, fmt.Sprintf("  Port: %d", *info.Port))
		}
		if info.SteamID != nil {
			console.PrintString(console.LevelInfo, fmt.Sprintf("  SteamID: %d", *info.SteamID))
		}
		if info.SourceTV != nil {
			console.PrintString(console.LevelInfo, fmt.Sprintf("  SourceTV: %s:%d", info.SourceTV.Name, info.SourceTV.Port))
		}
		if info.Keywords != nil && *info.Keywords != "" {
			console.PrintString(console.LevelInfo, fmt.Sprintf("  Tags: %s", *info.Keywords))
		}

		return nil
	})

	// net_query_players command - Query player list
	console.AddCommand("net_query_players", "Query server for player list", "net_query_players <address:port>", func(options string) error {
		if options == "" {
			console.PrintString(console.LevelWarning, "Usage: net_query_players <address:port>")
			console.PrintString(console.LevelInfo, "Example: net_query_players 192.168.1.100:27015")
			return nil
		}

		address := strings.TrimSpace(options)

		console.PrintString(console.LevelInfo, fmt.Sprintf("Querying players on: %s", address))

		// Create client and connect
		client := network.NewClient()
		client.SetTimeout(5 * time.Second)
		if err := client.Connect(address); err != nil {
			console.PrintString(console.LevelError, fmt.Sprintf("Failed to connect: %v", err))
			return err
		}
		defer client.Disconnect()

		// Query players
		players, err := client.QueryPlayers()
		if err != nil {
			console.PrintString(console.LevelError, fmt.Sprintf("Failed to query players: %v", err))
			return err
		}

		// Display player list
		console.PrintString(console.LevelSuccess, fmt.Sprintf("Player List (%d players):", players.Count))
		if players.Count == 0 {
			console.PrintString(console.LevelInfo, "  (No players)")
		} else {
			for _, player := range players.Players {
				duration := time.Duration(player.Duration * float32(time.Second))
				console.PrintString(console.LevelInfo, fmt.Sprintf("  [%d] %s - Score: %d - Time: %v",
					player.Index, player.Name, player.Score, duration.Round(time.Second)))
			}
		}

		return nil
	})

	// net_query_rules command - Query server rules/cvars
	console.AddCommand("net_query_rules", "Query server for rules/cvars", "net_query_rules <address:port>", func(options string) error {
		if options == "" {
			console.PrintString(console.LevelWarning, "Usage: net_query_rules <address:port>")
			console.PrintString(console.LevelInfo, "Example: net_query_rules 192.168.1.100:27015")
			return nil
		}

		address := strings.TrimSpace(options)

		console.PrintString(console.LevelInfo, fmt.Sprintf("Querying rules on: %s", address))

		// Create client and connect
		client := network.NewClient()
		client.SetTimeout(5 * time.Second)
		if err := client.Connect(address); err != nil {
			console.PrintString(console.LevelError, fmt.Sprintf("Failed to connect: %v", err))
			return err
		}
		defer client.Disconnect()

		// Query rules
		rules, err := client.QueryRules()
		if err != nil {
			console.PrintString(console.LevelError, fmt.Sprintf("Failed to query rules: %v", err))
			return err
		}

		// Display rules
		console.PrintString(console.LevelSuccess, fmt.Sprintf("Server Rules (%d rules):", rules.Count))
		if rules.Count == 0 {
			console.PrintString(console.LevelInfo, "  (No rules)")
		} else {
			// Sort keys for consistent display
			for name, value := range rules.Rules {
				console.PrintString(console.LevelInfo, fmt.Sprintf("  %s = %s", name, value))
			}
		}

		return nil
	})

	// Add network timeout convar
	console.AddConvarFloat("net_timeout", "Network operation timeout in seconds", 5.0)
}
