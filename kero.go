package kero

import (
	"github.com/galaco/kero/client/gui"
	"github.com/galaco/kero/client/input"
	"github.com/galaco/kero/client/renderer"
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

	scene   *scene.Scene
	physics *physics.PhysicsSystem

	input    *input.Input
	renderer *renderer.Renderer
	ui       *gui.Gui
}

// RegisterGameDefinitions sets up provided game-specific configuration
func (kero *Kero) RegisterGameDefinitions(def game.Definition) {
	def.RegisterEntityClasses()
}

// Start runs the game loop
func (kero *Kero) Start() {
	shared.AddInitialConvars()
	kero.input = input.InitializeInput()
	kero.renderer = renderer.NewRenderer()
	kero.ui = gui.NewGui()
	kero.scene = scene.NewScene()
	kero.physics = physics.NewPhysicsSystem()

	kero.isRunning = true

	kero.physics.Initialize()
	kero.scene.Initialize()

	kero.renderer.Initialize()
	kero.ui.Initialize()

	event.Get().AddListener(messages.TypeEngineQuit, kero.onQuit)

	kero.mainLoop()

	kero.exit()
}

func (kero *Kero) mainLoop() {
	dt := 0.0
	startingTime := time.Now().UTC()
	for kero.isRunning && (window.CurrentWindow() != nil && !window.CurrentWindow().ShouldClose()) {
		kero.input.Poll()

		kero.physics.Update(dt)
		kero.scene.Update(dt)

		kero.renderer.Render()
		kero.ui.Render()

		window.CurrentWindow().SwapBuffers()
		kero.renderer.FinishFrame()

		dt = float64(time.Now().UTC().Sub(startingTime).Nanoseconds()/1000000) / 1000
		startingTime = time.Now().UTC()
	}
}

func (kero *Kero) onQuit(e interface{}) {
	window.CurrentWindow().Close()
}

func (kero *Kero) exit() {
	kero.physics.Cleanup()
	kero.renderer.Cleanup()
}

// NewKero returns a new Kero instance
func NewKero() *Kero {
	return &Kero{
		isRunning: false,
	}
}
