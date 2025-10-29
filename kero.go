package kero

import (
	"github.com/galaco/kero/engine"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	scene2 "github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/game"
	"github.com/galaco/kero/gui"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
	"github.com/galaco/kero/physics"
	"github.com/galaco/kero/renderer"
	"github.com/galaco/kero/scene"
	"time"
)

// Kero provides a game loop with explicit dependency injection
type Kero struct {
	isRunning bool
	engine    *engine.Engine
}

// RegisterGameDefinitions sets up provided game-specific configuration
func (kero *Kero) RegisterGameDefinitions(def game.Definition) {
	// Register entity classes with the engine's entity registry
	def.RegisterEntityClasses()
}

// Start runs the game loop
func (kero *Kero) Start(gameDir string) error {
	middleware.AddInitialConvars()

	// Initialize engine context
	kero.engine = engine.NewEngine()

	// Initialize filesystem first (needed by other systems)
	fs, err := filesystem.Init(gameDir)
	if err != nil {
		return err
	}
	kero.engine.SetFileSystem(fs)

	// Initialize event bus
	eventBus := event.NewDispatcher()
	eventBus.Initialize()
	kero.engine.SetEventBus(eventBus)

	// Initialize entity registry
	entityRegistry := entity.NewRegistry()
	kero.engine.SetEntityRegistry(entityRegistry)

	// Initialize scene manager
	sceneManager := scene2.NewManager()
	kero.engine.SetSceneManager(sceneManager)

	// Initialize systems with explicit dependencies
	input := middleware.NewInput(eventBus)
	kero.engine.SetInput(input)

	renderer := renderer.NewRenderer(eventBus, fs)
	kero.engine.SetRenderer(renderer)

	ui := gui.NewGui(eventBus, fs, input)
	kero.engine.SetGUI(ui)

	sceneSystem := scene.NewScene(eventBus, fs, sceneManager, input)
	kero.engine.SetScene(sceneSystem)

	physicsSystem := physics.NewPhysicsSystem(eventBus, sceneManager)
	kero.engine.SetPhysics(physicsSystem)

	kero.isRunning = true

	// Initialize all systems
	physicsSystem.Initialize()
	sceneSystem.Initialize()
	renderer.Initialize()
	ui.Initialize()

	eventBus.AddListener(messages.TypeEngineQuit, kero.onQuit)

	kero.mainLoop()

	kero.exit()

	return nil
}

func (kero *Kero) mainLoop() {
	dt := 0.0
	startingTime := time.Now().UTC()
	for kero.isRunning && (window.CurrentWindow() != nil && !window.CurrentWindow().ShouldClose()) {
		kero.engine.Input().Poll()

		kero.engine.Physics().Update(dt)
		kero.engine.Scene().Update(dt)

		kero.engine.Renderer().Render()
		kero.engine.GUI().Render()

		window.CurrentWindow().SwapBuffers()
		kero.engine.Renderer().FinishFrame()

		dt = float64(time.Now().UTC().Sub(startingTime).Nanoseconds()/1000000) / 1000
		startingTime = time.Now().UTC()
	}
}

func (kero *Kero) onQuit(e interface{}) {
	window.CurrentWindow().Close()
}

func (kero *Kero) exit() {
	kero.engine.Physics().Cleanup()
	kero.engine.Renderer().Cleanup()
}

// NewKero returns a new Kero instance
func NewKero() *Kero {
	return &Kero{
		isRunning: false,
	}
}
