package network_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/galaco/kero/network"
)

// ExampleQuickQueryServerInfo demonstrates querying a server for basic information
func ExampleQuickQueryServerInfo() {
	// Query a server (replace with a real server address)
	info, err := network.QuickQueryServerInfo("192.168.1.100:27015")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Server: %s\n", info.Name)
	fmt.Printf("Map: %s\n", info.Map)
	fmt.Printf("Players: %d/%d\n", info.Players, info.MaxPlayers)
}

// ExampleClient_QueryServerInfo demonstrates using a persistent client connection
func ExampleClient_QueryServerInfo() {
	// Create a new client
	client := network.NewClient()
	client.SetTimeout(5 * time.Second)

	// Connect to server
	if err := client.Connect("192.168.1.100:27015"); err != nil {
		fmt.Printf("Connection error: %v\n", err)
		return
	}
	defer client.Disconnect()

	// Query server info
	info, err := client.QueryServerInfo()
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		return
	}

	fmt.Printf("Server: %s\n", info.Name)
	fmt.Printf("Map: %s\n", info.Map)
	fmt.Printf("Game: %s\n", info.Game)
}

// ExampleClient_QueryPlayers demonstrates querying the player list
func ExampleClient_QueryPlayers() {
	client := network.NewClient()
	if err := client.Connect("192.168.1.100:27015"); err != nil {
		fmt.Printf("Connection error: %v\n", err)
		return
	}
	defer client.Disconnect()

	// Query players (handles challenge-response automatically)
	players, err := client.QueryPlayers()
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		return
	}

	fmt.Printf("Players online: %d\n", players.Count)
	for _, player := range players.Players {
		fmt.Printf("  %s - Score: %d\n", player.Name, player.Score)
	}
}

// ExampleClient_QueryRules demonstrates querying server rules/cvars
func ExampleClient_QueryRules() {
	client := network.NewClient()
	if err := client.Connect("192.168.1.100:27015"); err != nil {
		fmt.Printf("Connection error: %v\n", err)
		return
	}
	defer client.Disconnect()

	// Query rules (handles challenge-response automatically)
	rules, err := client.QueryRules()
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		return
	}

	fmt.Printf("Server rules: %d\n", rules.Count)
	for name, value := range rules.Rules {
		fmt.Printf("  %s = %s\n", name, value)
	}
}

// TestConnectionlessPacketParsing tests the basic packet parsing
func TestConnectionlessPacketParsing(t *testing.T) {
	// This test doesn't require a live server - it tests packet parsing logic
	// We'll add actual tests once we have some recorded packet data
	t.Skip("Skipping - requires recorded packet data for testing")
}
