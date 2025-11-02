package scene

import (
	"context"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/input"
	scene2 "github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
	loader "github.com/galaco/kero/scene/loaders"
	"runtime"
)

type loadResult struct {
	level *graphics.Bsp
	ents  []entity.IEntity
	err   error
}

type Scene struct {
	eventBus       *event.Dispatcher
	fileSystem     filesystem.FileSystem
	sceneManager   *scene2.Manager
	inputMiddleware *middleware.Input

	dataScene *scene2.StaticScene

	listenToInput bool

	// Async loading support
	loadingCtx    context.Context
	loadingCancel context.CancelFunc
	loadComplete  chan loadResult
	isLoading     bool
}

func (s *Scene) Initialize() {
	// Register typed event listeners (Phase 3)
	event.RegisterTypedEvent(s.inputMiddleware.EventBus(), s.onKeyReleaseTyped)
	event.RegisterTypedEvent(s.inputMiddleware.EventBus(), s.onMouseMoveTyped)
	event.RegisterTypedEvent(s.eventBus, s.onChangeLevelTyped)
	event.RegisterTypedEvent(s.eventBus, func(e messages.EngineDisconnectEvent) {
		s.sceneManager.CloseCurrentScene()
		s.dataScene = nil // Clear the scene reference so IsLevelLoaded() returns false

		// Reset input capture state and unlock mouse
		s.listenToInput = false
		input.Mouse().UnlockMousePosition()

		runtime.GC()
	})
}

func (s *Scene) Update(dt float64) {
	// Check for async load completion (non-blocking)
	if s.isLoading {
		select {
		case result := <-s.loadComplete:
			s.isLoading = false

			if result.err != nil {
				// Handle error/cancellation
				console.PrintString(console.LevelError, result.err.Error())
				event.DispatchTyped(s.eventBus, messages.LoadingLevelProgressEvent{
					State: messages.LoadingProgressStateError,
				})
				return
			}

			// Success - create static scene and dispatch events on main thread
			console.PrintString(console.LevelInfo, "Generating Static World...")
			s.dataScene = scene2.LoadStaticSceneFromBsp(s.fileSystem, result.level, result.ents)
			s.sceneManager.SetCurrentScene(s.dataScene)
			console.PrintString(console.LevelInfo, "Complete!")

			// Clear event queue and dispatch completion events
			s.eventBus.CancelPending()
			event.DispatchTyped(s.eventBus, messages.LoadingLevelParsedEvent{Level: s.dataScene})
			event.DispatchTyped(s.eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateFinished})

		default:
			// Loading still in progress - nothing to do
		}
		return
	}

	if s.dataScene == nil {
		return
	}
	if s.listenToInput {
		// Update camera first to ensure direction vectors are current
		// (rotation from mouse events may have happened since last frame)
		s.dataScene.Camera.Update(dt)

		// Build movement input direction from keyboard
		var forward, right float32
		if input.Keyboard().IsKeyPressed(input.KeyW) {
			forward += 1.0
		}
		if input.Keyboard().IsKeyPressed(input.KeyS) {
			forward -= 1.0
		}
		if input.Keyboard().IsKeyPressed(input.KeyD) {
			right += 1.0
		}
		if input.Keyboard().IsKeyPressed(input.KeyA) {
			right -= 1.0
		}

		// Apply movement input to camera (uses updated direction vectors)
		s.dataScene.Camera.SetMovementInput(forward, right, dt)
	}

	for _, e := range s.dataScene.Entities {
		e.Think(dt)
	}
}

func (s *Scene) onChangeLevelTyped(e messages.ChangeLevelEvent) {
	// Prevent multiple simultaneous loads
	if s.isLoading {
		console.PrintString(console.LevelWarning, "A map is already loading. Please wait or cancel the current load.")
		return
	}

	if s.dataScene != nil {
		// Cleanup old scene
		s.dataScene = nil
	}

	// Mark as loading
	s.isLoading = true

	// Create cancellable context for this load
	s.loadingCtx, s.loadingCancel = context.WithCancel(context.Background())

	// Initialize channel if needed
	if s.loadComplete == nil {
		s.loadComplete = make(chan loadResult, 1)
	}

	// Dispatch loading started event (shows loading screen)
	event.DispatchTyped(s.eventBus, messages.LoadingLevelProgressEvent{
		State: messages.LoadingProgressStateStarted,
	})

	// Start async loading in goroutine
	go s.loadMapAsync(e.MapName)
}

func (s *Scene) loadMapAsync(mapName string) {
	// CPU-only work in goroutine - no OpenGL calls
	level, ents, err := loader.LoadBspMapWithContext(
		s.loadingCtx,
		s.fileSystem,
		s.eventBus,
		mapName,
	)

	// Send result to main thread via channel
	s.loadComplete <- loadResult{
		level: level,
		ents:  ents,
		err:   err,
	}
}

// CancelLoading cancels the current map loading operation
func (s *Scene) CancelLoading() {
	if s.isLoading && s.loadingCancel != nil {
		console.PrintString(console.LevelInfo, "Cancelling map load...")
		s.loadingCancel()
	}
}

// IsLevelLoaded returns true if a level is currently loaded
func (s *Scene) IsLevelLoaded() bool {
	return s.dataScene != nil
}

func (s *Scene) onKeyReleaseTyped(e messages.KeyReleaseEvent) {
	if e.Key == input.KeyEscape {
		// Only allow toggling input capture if a level is loaded
		if s.dataScene == nil {
			return
		}

		// Toggle input capture state
		s.listenToInput = !s.listenToInput

		// Synchronize mouse lock state with input listening state
		if s.listenToInput {
			input.Mouse().LockMousePosition()
		} else {
			input.Mouse().UnlockMousePosition()
		}
	}
}

func (s *Scene) onMouseMoveTyped(e messages.MouseMoveEvent) {
	if s.dataScene == nil || s.dataScene.Camera == nil || !s.listenToInput {
		return
	}
	s.dataScene.Camera.Rotate(e.Delta[0], 0, e.Delta[1])
}

// NewScene creates a new scene with explicit dependencies
func NewScene(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, sceneManager *scene2.Manager, inputMiddleware *middleware.Input) *Scene {
	return &Scene{
		eventBus:        eventBus,
		fileSystem:      fileSystem,
		sceneManager:    sceneManager,
		inputMiddleware: inputMiddleware,
	}
}
