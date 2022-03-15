package kero

import (
	"github.com/galaco/kero/client"
	"github.com/galaco/kero/internal/framework/event"
	"github.com/galaco/kero/internal/framework/window"
	"github.com/galaco/kero/shared"
	"github.com/galaco/kero/shared/game"
	"github.com/galaco/kero/shared/messages"
	"github.com/galaco/kero/shared/physics"
	"github.com/galaco/kero/shared/scene"
	"time"
)

// Kero provides a game loop
type Kero struct {
	isRunning bool

	sharedScene   *scene.Scene
	sharedPhysics *physics.PhysicsSystem

	client *client.Client
}

// RegisterGameDefinitions sets up provided game-specific configuration
func (kero *Kero) RegisterGameDefinitions(def game.Definition) {
	def.RegisterEntityClasses()
}

// Start runs the game loop
func (kero *Kero) Start() {
	shared.AddInitialConvars()
	kero.sharedScene = scene.NewScene()
	kero.sharedPhysics = physics.NewPhysicsSystem()

	kero.isRunning = true

	kero.sharedPhysics.Initialize()
	kero.sharedScene.Initialize()

	kero.client = client.NewClient()
	kero.client.Initialize()

	event.Get().AddListener(messages.TypeEngineQuit, kero.onQuit)

	kero.mainLoop()

	kero.exit()
}

func (kero *Kero) mainLoop() {
	dt := 0.0
	startingTime := time.Now().UTC()
	for kero.isRunning && (window.CurrentWindow() != nil && !window.CurrentWindow().ShouldClose()) {
		kero.client.FixedUpdate(dt)
		kero.client.Update()

		kero.sharedPhysics.Update(dt)
		kero.sharedScene.Update(dt)

		dt = float64(time.Now().UTC().Sub(startingTime).Nanoseconds()/1000000) / 1000
		startingTime = time.Now().UTC()
	}
}

func (kero *Kero) onQuit(e interface{}) {
	window.CurrentWindow().Close()
}

func (kero *Kero) exit() {
	kero.sharedPhysics.Cleanup()
	kero.client.Cleanup()
}

// NewKero returns a new Kero instance
func NewKero() *Kero {
	return &Kero{
		isRunning: false,
	}
}
