package scene

import (
	"context"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
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

	listenToInput bool

	// Async loading support
	loadingCtx    context.Context
	loadingCancel context.CancelFunc
	loadComplete  chan loadResult
	isLoading     bool

	// Phase 4: Pure ECS player
	ecsWorld             interface{} // *ecs.World (interface{} to avoid import cycle)
	playerEntity         interface{} // ecs.Entity (stored as interface{})
	playerMovementSystem interface{} // *systems.PlayerMovementSystem
	playerCameraSystem   interface{} // *systems.PlayerCameraSystem
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

	if s.listenToInput {
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

		// Create player input command
		playerInput := gameEntity.PlayerInput{
			Forward: forward,
			Right:   right,
			Up:      0,
			Yaw:     0, // Mouse delta handled separately in onMouseMoveTyped
			Pitch:   0,
			Buttons: buttons,
		}

		// Phase 4: Use ECS player systems
		if s.playerMovementSystem != nil {
			type movementSystem interface {
				Update(input gameEntity.PlayerInput, dt float64)
			}
			if sys, ok := s.playerMovementSystem.(movementSystem); ok {
				sys.Update(playerInput, dt)
			}
		}

		// Update camera system
		if s.playerCameraSystem != nil {
			type cameraSystem interface {
				Update()
			}
			if sys, ok := s.playerCameraSystem.(cameraSystem); ok {
				sys.Update()
			}
		}
	}

	// Update all entities
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

	// Get sensitivity multiplier from ConVar (default 1.0)
	sens := console.GetConvarFloat("m_sensitivity")
	if sens <= 0 {
		sens = 1.0
	}
	sensitivity := float32(0.03) * sens

	// Create input command with mouse delta
	mouseInput := gameEntity.PlayerInput{
		Yaw:   e.Delta[0] * sensitivity,
		Pitch: e.Delta[1] * sensitivity,
	}

	// Phase 4: Use ECS player systems
	if s.playerMovementSystem != nil {
		// Use ECS player movement system (dt=0 for instant rotation)
		type movementSystem interface {
			Update(input gameEntity.PlayerInput, dt float64)
		}
		if sys, ok := s.playerMovementSystem.(movementSystem); ok {
			sys.Update(mouseInput, 0)
		}

		// Update camera immediately
		if s.playerCameraSystem != nil {
			type cameraSystem interface {
				Update()
			}
			if sys, ok := s.playerCameraSystem.(cameraSystem); ok {
				sys.Update()
			}
		}
	} else {
		// No player: direct camera control
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

	// Calculate player center position (capsule center)
	spawnOffset := float32(gameEntity.PlayerHeight/2) + 2.0
	playerCenterPos := spawnPos.Add(mgl32.Vec3{0, 0, spawnOffset})

	// Phase 4: Always create ECS player
	if s.ecsWorld != nil {
		s.spawnECSPlayer(playerCenterPos, spawnYaw)
	}
}

// spawnECSPlayer creates a player using ECS components
func (s *Scene) spawnECSPlayer(playerCenterPos mgl32.Vec3, spawnYaw float32) {
	world := s.ecsWorld.(*ecs.World)

	// Create player entity
	playerEntity := world.CreateEntity()

	// Transform component
	transform := components.Transform{
		Position:    playerCenterPos,
		Orientation: mgl32.AnglesToQuat(0, 0, spawnYaw, mgl32.ZYX),
		Scale:       mgl32.Vec3{1, 1, 1},
	}
	ecs.AddComponent(world, playerEntity, transform)

	// PlayerController component
	sens := console.GetConvarFloat("m_sensitivity")
	if sens <= 0 {
		sens = 1.0
	}
	controller := components.NewPlayerController(0, spawnYaw, sens)
	ecs.AddComponent(world, playerEntity, controller)

	// CharacterMovement component
	movement := components.NewCharacterMovement()
	ecs.AddComponent(world, playerEntity, movement)

	// CharacterController component (handle populated by physics system)
	charController := components.NewCharacterController(
		float32(gameEntity.PlayerHeight),
		float32(gameEntity.PlayerRadius),
		float32(gameEntity.PlayerStepHeight))
	ecs.AddComponent(world, playerEntity, charController)

	// CameraController component
	eyeOffset := mgl32.Vec3{0, 0, float32(gameEntity.PlayerEyeHeight - gameEntity.PlayerHeight/2)}
	fov := console.GetConvarFloat("fov")
	if fov <= 0 {
		fov = 90
	}
	cameraController := components.NewCameraController(eyeOffset, fov, true)
	cameraController.SetCamera(s.dataScene.Camera)
	ecs.AddComponent(world, playerEntity, cameraController)

	// Store player entity
	s.playerEntity = playerEntity

	// Phase 4: No legacy player creation - CharacterController created by PhysicsSystem
	// Register ECS player entity with physics system
	if s.physicsSystem != nil {
		type playerRegistrar interface {
			RegisterPlayer(ecsPlayer ...interface{})
		}
		if registrar, ok := s.physicsSystem.(playerRegistrar); ok {
			registrar.RegisterPlayer(playerEntity)
			console.PrintString(console.LevelInfo, "Registered ECS player with physics system")
		} else {
			console.PrintString(console.LevelWarning, "Failed to register ECS player with physics system - type assertion failed")
		}
	}

	console.PrintString(console.LevelSuccess, "ECS Player spawned")
}

// Phase 4: Legacy player removed - pure ECS player only

// NewScene creates a new scene with explicit dependencies.
// physicsSystem should be the *PhysicsSystem from the physics package (passed as interface{} to avoid import cycle).
// ecsWorld should be the *ecs.World (passed as interface{} to avoid import cycle).
// playerMovementSystem should be *systems.PlayerMovementSystem (passed as interface{} to avoid import cycle).
// playerCameraSystem should be *systems.PlayerCameraSystem (passed as interface{} to avoid import cycle).
func NewScene(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, sceneManager *scene2.Manager, inputMiddleware *middleware.Input, physicsSystem interface{}, ecsWorld interface{}, playerMovementSystem interface{}, playerCameraSystem interface{}) *Scene {
	return &Scene{
		eventBus:             eventBus,
		fileSystem:           fileSystem,
		sceneManager:         sceneManager,
		inputMiddleware:      inputMiddleware,
		physicsSystem:        physicsSystem,
		ecsWorld:             ecsWorld,
		playerMovementSystem: playerMovementSystem,
		playerCameraSystem:   playerCameraSystem,
	}
}
