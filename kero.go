package kero

import (
	"github.com/galaco/kero/client"
	"github.com/galaco/kero/server"
	"github.com/galaco/kero/shared"
	"github.com/galaco/kero/shared/game"
	"github.com/galaco/kero/shared/physics"
	"github.com/galaco/kero/shared/scene"
	"time"
)

// Kero provides a game loop
type Kero struct {
	sharedPhysics *physics.Simulation

	client *client.Client
	server *server.Server
}

// RegisterGameDefinitions sets up provided game-specific configuration
func (kero *Kero) RegisterGameDefinitions(def game.Definition) {
	def.RegisterEntityClasses()
}

// Start runs the game loop
func (kero *Kero) Start() {
	shared.BindSharedConsoleCommands()

	// Shared systems
	kero.sharedPhysics = physics.NewSimulation()
	kero.sharedPhysics.Initialize()
	scene.CurrentScene().Initialize()

	// Server systems
	kero.server = server.NewServer()

	// Client systems
	kero.client = client.NewClient()
	kero.client.Initialize()

	// Run the actual simulation
	kero.mainLoop()
	kero.exit()
}

func (kero *Kero) mainLoop() {
	var dt float64
	startingTime := time.Now().UTC()
	for !kero.client.ShouldClose() {

		// Server stuff
		kero.sharedPhysics.Update(dt)
		kero.server.FixedUpdate(dt)
		kero.server.Update()

		// Client stuff
		kero.client.FixedUpdate(dt)
		kero.client.Update()

		dt = float64(time.Now().UTC().Sub(startingTime).Nanoseconds()/1000000) / 1000
		startingTime = time.Now().UTC()
	}
}

func (kero *Kero) exit() {
	kero.sharedPhysics.Cleanup()
	kero.client.Cleanup()
	kero.server.Cleanup()
}

// NewKero returns a new Kero instance
func NewKero() *Kero {
	return &Kero{}
}
