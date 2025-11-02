package kero

import (
	"github.com/galaco/kero/engine"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/legacy"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/metrics"
	scene2 "github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/game"
	"github.com/galaco/kero/gui"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
	"github.com/galaco/kero/physics"
	"github.com/galaco/kero/renderer"
	"github.com/galaco/kero/scene"
	"runtime"
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

	// Initialize metrics collector (300 samples = ~5 seconds at 60fps)
	metricsCollector := metrics.NewCollector(300)
	kero.engine.SetMetrics(metricsCollector)

	// Initialize systems with explicit dependencies
	input := middleware.NewInput(eventBus)
	kero.engine.SetInput(input)

	renderer := renderer.NewRenderer(eventBus, fs, ecsWorld, legacyBridge)
	kero.engine.SetRenderer(renderer)

	sceneSystem := scene.NewScene(eventBus, fs, sceneManager, input)
	kero.engine.SetScene(sceneSystem)

	ui := gui.NewGui(eventBus, fs, input, metricsCollector, sceneSystem)
	kero.engine.SetGUI(ui)

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

// collectMemoryStats collects RAM usage statistics
func (kero *Kero) collectMemoryStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Record heap allocated memory in MB
	heapAllocMB := float64(memStats.Alloc) / 1024.0 / 1024.0
	kero.engine.Metrics().RecordValue("memory_heap", heapAllocMB)

	// Record total system memory in MB
	sysTotalMB := float64(memStats.Sys) / 1024.0 / 1024.0
	kero.engine.Metrics().RecordValue("memory_sys", sysTotalMB)
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
		// Start frame timing
		frameStartTime := time.Now()

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
		kero.engine.Metrics().StartTiming("input")
		kero.engine.Input().Poll()
		kero.engine.Metrics().EndTiming("input")

		// Phase 1: Pre-Update (process events queued before game logic)
		kero.engine.Metrics().StartTiming("event_preupdate")
		kero.engine.EventBus().ProcessPhase(event.PhasePreUpdate)
		kero.engine.Metrics().EndTiming("event_preupdate")

		// Fixed timestep physics (may run 0, 1, or multiple times per frame)
		physicsSteps := 0
		kero.engine.Metrics().StartTiming("physics_total")
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
		kero.engine.Metrics().EndTiming("physics_total")
		kero.engine.Metrics().RecordValue("physics_steps", float64(physicsSteps))

		// Variable timestep updates (scene logic, rendering)
		kero.engine.Metrics().StartTiming("scene_update")
		kero.engine.Scene().Update(frameDt)
		kero.engine.Metrics().EndTiming("scene_update")

		// Phase 2: Post-Update (process events after game logic, before rendering)
		kero.engine.Metrics().StartTiming("event_postupdate")
		kero.engine.EventBus().ProcessPhase(event.PhasePostUpdate)
		kero.engine.Metrics().EndTiming("event_postupdate")

		// Phase 3: Pre-Render (process events before rendering begins)
		kero.engine.Metrics().StartTiming("event_prerender")
		kero.engine.EventBus().ProcessPhase(event.PhasePreRender)
		kero.engine.Metrics().EndTiming("event_prerender")

		// Prepare debug visualization (populate debug buffer)
		if console.GetConvarBoolean("r_drawcollisionmodels") {
			kero.engine.Physics().PrepareDebug(kero.engine.Renderer().GetDebugBuffer())
		}

		// Render (interpolation factor for future use)
		// interpolation := float32(accumulator / FixedDt)
		kero.engine.Metrics().StartTiming("render")
		kero.engine.Renderer().Render()
		kero.engine.Metrics().EndTiming("render")

		kero.engine.Metrics().StartTiming("gui")
		kero.engine.GUI().Render(float32(frameDt))
		kero.engine.Metrics().EndTiming("gui")

		kero.engine.Metrics().StartTiming("swap_buffers")
		window.CurrentWindow().SwapBuffers()
		kero.engine.Renderer().FinishFrame()
		kero.engine.Metrics().EndTiming("swap_buffers")

		// Phase 4: Post-Render (process events after frame completes)
		kero.engine.Metrics().StartTiming("event_postrender")
		kero.engine.EventBus().ProcessPhase(event.PhasePostRender)
		kero.engine.Metrics().EndTiming("event_postrender")

		// Record total frame time
		frameTotalTime := time.Since(frameStartTime)
		kero.engine.Metrics().RecordValue("frame_total", float64(frameTotalTime.Microseconds())/1000.0)

		// Calculate and record FPS
		if frameTotalTime.Seconds() > 0 {
			fps := 1.0 / frameTotalTime.Seconds()
			kero.engine.Metrics().RecordValue("fps", fps)
		}

		// Collect memory statistics
		kero.collectMemoryStats()
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
