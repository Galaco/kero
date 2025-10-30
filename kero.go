package kero

import (
	"github.com/galaco/kero/engine"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/legacy"
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

	// Initialize ECS World (Phase 4)
	ecsWorld := ecs.NewWorld()
	kero.engine.SetECSWorld(ecsWorld)

	// Initialize shared Legacy Bridge (Phase 4)
	// Single bridge instance shared by all systems
	legacyBridge := legacy.NewBridge(ecsWorld)
	kero.engine.SetLegacyBridge(legacyBridge)

	// Initialize systems with explicit dependencies
	input := middleware.NewInput(eventBus)
	kero.engine.SetInput(input)

	renderer := renderer.NewRenderer(eventBus, fs, ecsWorld, legacyBridge)
	kero.engine.SetRenderer(renderer)

	ui := gui.NewGui(eventBus, fs, input)
	kero.engine.SetGUI(ui)

	sceneSystem := scene.NewScene(eventBus, fs, sceneManager, input)
	kero.engine.SetScene(sceneSystem)

	physicsSystem := physics.NewPhysicsSystem(eventBus, sceneManager, ecsWorld, legacyBridge)
	kero.engine.SetPhysics(physicsSystem)

	kero.isRunning = true

	// Initialize all systems
	physicsSystem.Initialize()
	sceneSystem.Initialize()
	renderer.Initialize()
	ui.Initialize()

	// Register typed event listeners (Phase 3)
	event.RegisterTypedEvent(eventBus, kero.onQuitTyped)

	kero.mainLoop()

	kero.exit()

	return nil
}

func (kero *Kero) onQuitTyped(e messages.EngineQuitEvent) {
	window.CurrentWindow().Close()
}

func (kero *Kero) mainLoop() {
	// Fixed timestep configuration
	const PhysicsHz = 60.0
	const FixedDt = 1.0 / PhysicsHz      // 16.67ms per physics step
	const MaxAccumulator = 0.25          // Cap at 250ms (prevents spiral of death)
	const MaxPhysicsSteps = 5            // Max physics iterations per frame

	accumulator := 0.0
	currentTime := time.Now()

	for kero.isRunning && (window.CurrentWindow() != nil && !window.CurrentWindow().ShouldClose()) {
		// Calculate frame time
		newTime := time.Now()
		frameDt := newTime.Sub(currentTime).Seconds()
		currentTime = newTime

		// Prevent spiral of death - cap frame time
		if frameDt > MaxAccumulator {
			frameDt = MaxAccumulator
		}

		accumulator += frameDt

		// Input processing (once per frame)
		kero.engine.Input().Poll()

		// Phase 1: Pre-Update (process events queued before game logic)
		kero.engine.EventBus().ProcessPhase(event.PhasePreUpdate)

		// Fixed timestep physics (may run 0, 1, or multiple times per frame)
		physicsSteps := 0
		for accumulator >= FixedDt {
			kero.engine.Physics().FixedUpdate(FixedDt)
			accumulator -= FixedDt
			physicsSteps++

			// Safety: prevent infinite loop if physics is too slow
			if physicsSteps >= MaxPhysicsSteps {
				accumulator = 0
				break
			}
		}

		// Variable timestep updates (scene logic, rendering)
		kero.engine.Scene().Update(frameDt)

		// Phase 2: Post-Update (process events after game logic, before rendering)
		kero.engine.EventBus().ProcessPhase(event.PhasePostUpdate)

		// Phase 3: Pre-Render (process events before rendering begins)
		kero.engine.EventBus().ProcessPhase(event.PhasePreRender)

		// Render (interpolation factor for future use)
		// interpolation := float32(accumulator / FixedDt)
		kero.engine.Renderer().Render()
		kero.engine.GUI().Render()

		window.CurrentWindow().SwapBuffers()
		kero.engine.Renderer().FinishFrame()

		// Phase 4: Post-Render (process events after frame completes)
		kero.engine.EventBus().ProcessPhase(event.PhasePostRender)
	}
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
