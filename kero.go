package kero

import (
	"github.com/galaco/kero/client"
	"github.com/galaco/kero/shared"
	"github.com/galaco/kero/shared/game"
	"github.com/galaco/kero/shared/physics"
	"github.com/galaco/kero/shared/scene"
	"time"
)

// Kero provides a game loop
type Kero struct {
	sharedScene   *scene.Scene
	sharedPhysics *physics.Simulation

	client *client.Client
}

// RegisterGameDefinitions sets up provided game-specific configuration
func (kero *Kero) RegisterGameDefinitions(def game.Definition) {
	def.RegisterEntityClasses()
}

// Start runs the game loop
func (kero *Kero) Start() {
	shared.BindSharedConsoleCommands()

	// Shared systems
	kero.sharedScene = scene.NewScene()
	kero.sharedPhysics = physics.NewSimulation()
	kero.sharedScene.Initialize()
	kero.sharedPhysics.Initialize()

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
		// Client stuff
		kero.client.FixedUpdate(dt)
		kero.client.Update()

		// Server stuff
		kero.sharedPhysics.Update(dt)
		kero.sharedScene.Update(dt)

		dt = float64(time.Now().UTC().Sub(startingTime).Nanoseconds()/1000000) / 1000
		startingTime = time.Now().UTC()
	}
}

func (kero *Kero) exit() {
	kero.sharedPhysics.Cleanup()
	kero.client.Cleanup()
}

// NewKero returns a new Kero instance
func NewKero() *Kero {
	return &Kero{}
}
