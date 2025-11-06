package scene

import (
	"context"
	"fmt"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/input"
	scene2 "github.com/galaco/kero/framework/scene"
	gameEntity "github.com/galaco/kero/game/entity"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
	loader "github.com/galaco/kero/scene/loaders"
	"github.com/go-gl/mathgl/mgl32"
	"runtime"
)

type loadResult struct {
	level *graphics.Bsp
	ents  []entity.IEntity
	err   error
}

type Scene struct {
	eventBus        *event.Dispatcher
	fileSystem      filesystem.FileSystem
	sceneManager    *scene2.Manager
	inputMiddleware *middleware.Input
	physicsSystem   interface{} // Physics system reference (interface{} to avoid import cycle)

	dataScene *scene2.StaticScene
	player    *gameEntity.Player // The player entity

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
	// TODO Phase 2: Add event listener for physics world ready to initialize player collision
	event.RegisterTypedEvent(s.eventBus, func(e messages.EngineDisconnectEvent) {
		s.sceneManager.CloseCurrentScene()
		s.dataScene = nil // Clear the scene reference so IsLevelLoaded() returns false
		s.player = nil    // Clear the player

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

			// Spawn player at info_player_start
			s.spawnPlayer()

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

	// Update camera to reflect any changes
	s.dataScene.Camera.Update(dt)

	if s.listenToInput && s.player != nil {
		// Build player input from keyboard/mouse
		var forward, right float32
		var buttons uint32

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
		if input.Keyboard().IsKeyPressed(input.KeySpace) {
			buttons |= gameEntity.ButtonJump
		}

		// Create player input command (mouse delta is applied in onMouseMoveTyped)
		playerInput := gameEntity.PlayerInput{
			Forward: forward,
			Right:   right,
			Up:      0,
			Yaw:     0, // Mouse delta handled separately
			Pitch:   0,
			Buttons: buttons,
		}

		// Debug: Log when we have input
		if forward != 0 || right != 0 {
			console.PrintString(console.LevelInfo, fmt.Sprintf("Processing input: forward=%.1f, right=%.1f", forward, right))
		}

		// Process player input (deterministic movement)
		s.player.ProcessInput(playerInput, dt)
	}

	// Update all entities
	for _, e := range s.dataScene.Entities {
		e.Think(dt)
	}

	// Update player
	if s.player != nil {
		s.player.Think(dt)
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
		s.player = nil
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
			console.PrintString(console.LevelInfo, "Input capture ENABLED - WASD should work now")
		} else {
			input.Mouse().UnlockMousePosition()
			console.PrintString(console.LevelInfo, "Input capture DISABLED - Can use GUI")
		}
	}
}

func (s *Scene) onMouseMoveTyped(e messages.MouseMoveEvent) {
	if s.dataScene == nil || s.dataScene.Camera == nil || !s.listenToInput {
		return
	}

	// If we have a player, apply mouse to player (which controls camera)
	// Otherwise fall back to direct camera control
	if s.player != nil {
		// Get sensitivity multiplier from ConVar (default 1.0)
		sens := console.GetConvarFloat("m_sensitivity")
		if sens <= 0 {
			sens = 1.0
		}
		sensitivity := float32(0.03) * sens

		// Create input command with mouse delta
		input := gameEntity.PlayerInput{
			Yaw:   e.Delta[0] * sensitivity,
			Pitch: e.Delta[1] * sensitivity,
		}
		s.player.ProcessInput(input, 0) // dt=0 for instant rotation update
	} else {
		// Fallback: direct camera control
		s.dataScene.Camera.Rotate(e.Delta[0], 0, e.Delta[1])
	}
}

// spawnPlayer finds the first info_player_start entity and spawns the player there
func (s *Scene) spawnPlayer() {
	if s.dataScene == nil {
		return
	}

	// Find first info_player_start entity
	var spawnPos mgl32.Vec3
	var spawnYaw float32
	found := false

	for _, e := range s.dataScene.Entities {
		if e.Classname() == "info_player_start" {
			spawnPos = e.Origin()
			// Get yaw from angles
			angles := e.VectorForKey("angles")
			spawnYaw = mgl32.DegToRad(angles[1]) // Y angle is yaw in Source Engine
			found = true
			console.PrintString(console.LevelInfo, "Found player spawn point")
			break
		}
	}

	if !found {
		// No spawn point found, use default position
		spawnPos = mgl32.Vec3{0, 0, 64}
		spawnYaw = 0
		console.PrintString(console.LevelWarning, "No info_player_start found, using default spawn")
	}

	// Create player at spawn position
	// NOTE: Source Engine spawn positions are at the player's feet, but Bullet capsules
	// are positioned at their center. Offset up by half the player height (36 units) + 2 unit clearance.
	spawnOffset := float32(gameEntity.PlayerHeight/2) + 2.0 // Extra clearance to avoid floor penetration
	playerCenterPos := spawnPos.Add(mgl32.Vec3{0, 0, spawnOffset})
	console.PrintString(console.LevelInfo, fmt.Sprintf("Spawn: feet=(%.1f,%.1f,%.1f) center=(%.1f,%.1f,%.1f) offset=%.1f",
		spawnPos[0], spawnPos[1], spawnPos[2],
		playerCenterPos[0], playerCenterPos[1], playerCenterPos[2],
		spawnOffset))
	s.player = gameEntity.NewPlayer(playerCenterPos, spawnYaw)

	// Give player the scene camera
	s.player.SetCamera(s.dataScene.Camera)

	// Register player with physics system for collision
	if s.physicsSystem != nil {
		// Call RegisterPlayer via interface (to avoid import cycle)
		type playerRegistrar interface {
			RegisterPlayer(player interface{})
		}
		if registrar, ok := s.physicsSystem.(playerRegistrar); ok {
			registrar.RegisterPlayer(s.player)
		}
	}

	console.PrintString(console.LevelSuccess, "Player spawned")
}

// NewScene creates a new scene with explicit dependencies.
// physicsSystem should be the *PhysicsSystem from the physics package (passed as interface{} to avoid import cycle).
func NewScene(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, sceneManager *scene2.Manager, inputMiddleware *middleware.Input, physicsSystem interface{}) *Scene {
	return &Scene{
		eventBus:        eventBus,
		fileSystem:      fileSystem,
		sceneManager:    sceneManager,
		inputMiddleware: inputMiddleware,
		physicsSystem:   physicsSystem,
	}
}
