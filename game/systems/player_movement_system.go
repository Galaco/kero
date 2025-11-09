package systems

import (
	"math"

	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
	"github.com/galaco/kero/framework/physics/collision"
	"github.com/galaco/kero/game/entity"
	"github.com/go-gl/mathgl/mgl32"
)

// PlayerMovementSystem handles player movement logic using ECS components.
// This system processes player input and updates velocity, applies physics, and handles collision.
type PlayerMovementSystem struct {
	ecsWorld *ecs.World
}

// NewPlayerMovementSystem creates a new player movement system
func NewPlayerMovementSystem(world *ecs.World) *PlayerMovementSystem {
	return &PlayerMovementSystem{
		ecsWorld: world,
	}
}

// Update processes player movement for all player entities
func (s *PlayerMovementSystem) Update(input entity.PlayerInput, dt float64) {
	// Query for player entities
	query := s.ecsWorld.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypePlayerController).
		With(ecs.ComponentTypeCharacterMovement).
		With(ecs.ComponentTypeCharacterController).
		Build()

	entities := query.Entities()
	if len(entities) == 0 {
		return
	}

	// Process each player (typically just one in singleplayer)
	for _, playerEntity := range entities {
		s.processPlayerMovement(playerEntity, input, dt)
	}
}

// processPlayerMovement handles movement for a single player entity
func (s *PlayerMovementSystem) processPlayerMovement(playerEntity ecs.Entity, input entity.PlayerInput, dt float64) {
	// Get components
	transform, _ := ecs.GetComponent[components.Transform](s.ecsWorld, playerEntity)
	controller, _ := ecs.GetComponent[components.PlayerController](s.ecsWorld, playerEntity)
	movement, _ := ecs.GetComponent[components.CharacterMovement](s.ecsWorld, playerEntity)
	charController, _ := ecs.GetComponent[components.CharacterController](s.ecsWorld, playerEntity)

	// ========================================
	// ROTATION (instantaneous, no dt needed)
	// ========================================
	controller.Yaw += input.Yaw
	controller.Pitch += input.Pitch

	// Clamp pitch to camera's coordinate system range (90° to 270°, centered at 180°)
	const cameraPitchMin = math.Pi / 2          // 90 degrees (looking down)
	const cameraPitchMax = 3 * math.Pi / 2      // 270 degrees (looking up)

	if controller.Pitch > cameraPitchMax {
		controller.Pitch = cameraPitchMax
	}
	if controller.Pitch < cameraPitchMin {
		controller.Pitch = cameraPitchMin
	}

	// Note: Camera rotation is handled by PlayerCameraSystem

	// ========================================
	// MOVEMENT (requires dt for integration)
	// ========================================
	if dt <= 0 {
		return // No movement without time
	}

	// Clear previous frame's collision debug data
	if charController.HasController() {
		cc := charController.GetController().(*collision.CharacterController)
		cc.ClearDebugHits()
	}

	// Check for noclip mode
	noclipEnabled := console.GetConvarBoolean("sv_noclip")

	// Calculate forward and right vectors based on camera orientation
	var forward, right mgl32.Vec3

	if noclipEnabled {
		// Noclip: Move in camera direction (includes vertical component from pitch)
		const cameraPitchCenter = math.Pi
		actualPitch := controller.Pitch - cameraPitchCenter

		forward = mgl32.Vec3{
			-float32(math.Sin(float64(controller.Yaw)) * math.Cos(float64(actualPitch))),
			-float32(math.Cos(float64(controller.Yaw)) * math.Cos(float64(actualPitch))),
			-float32(math.Sin(float64(actualPitch))),
		}

		right = mgl32.Vec3{
			-float32(math.Cos(float64(controller.Yaw))),
			float32(math.Sin(float64(controller.Yaw))),
			0,
		}
	} else {
		// Normal movement: Horizontal only
		forward = mgl32.Vec3{
			-float32(math.Sin(float64(controller.Yaw))),
			-float32(math.Cos(float64(controller.Yaw))),
			0,
		}
		right = mgl32.Vec3{
			float32(math.Sin(float64(controller.Yaw) - math.Pi/2)),
			float32(math.Cos(float64(controller.Yaw) - math.Pi/2)),
			0,
		}
	}

	// Build desired movement direction from input
	wishDir := mgl32.Vec3{0, 0, 0}
	if input.Forward != 0 {
		wishDir = wishDir.Add(forward.Mul(input.Forward))
	}
	if input.Right != 0 {
		wishDir = wishDir.Add(right.Mul(input.Right))
	}

	// Normalize if we have input
	wishSpeed := float32(0)
	if wishDir.Len() > 0.01 {
		wishDir = wishDir.Normalize()

		// If we have a wall normal from previous frame, clip wishDir along it
		if movement.LastWallNormal.Len() > 0.01 {
			dotProduct := wishDir.Dot(movement.LastWallNormal)
			if dotProduct < 0 {
				// Clip wishDir to be parallel to wall
				wishDir = wishDir.Sub(movement.LastWallNormal.Mul(dotProduct))
				if wishDir.Len() > 0.01 {
					wishDir = wishDir.Normalize()
				}
			}
		}

		if noclipEnabled {
			wishSpeed = console.GetConvarFloat("sv_noclip_speed")
		} else if movement.OnGround {
			wishSpeed = movement.GroundSpeed
		} else {
			wishSpeed = movement.AirSpeed
		}
	}

	// Noclip mode: Direct movement without physics
	if noclipEnabled {
		s.processNoclipMovement(playerEntity, wishDir, wishSpeed, dt)
		return
	}

	// Apply movement (acceleration-based)
	if movement.OnGround {
		// Ground movement: accelerate toward wish direction
		targetVel := wishDir.Mul(wishSpeed)
		movement.Velocity = movement.Velocity.Add(
			targetVel.Sub(movement.Velocity).Mul(float32(dt) * movement.Acceleration))

		// Apply friction when not moving
		if wishDir.Len() < 0.01 {
			friction := 1.0 - (movement.Friction * float32(dt))
			if friction < 0 {
				friction = 0
			}
			movement.Velocity = movement.Velocity.Mul(friction)
		}

		// Flatten velocity on ground
		movement.Velocity[2] = 0
	} else {
		// Air movement: limited air control
		airAccel := movement.Acceleration * 0.2
		targetVel := wishDir.Mul(wishSpeed)
		movement.Velocity = movement.Velocity.Add(
			targetVel.Sub(movement.Velocity).Mul(float32(dt) * airAccel))
	}

	// Handle jump
	if input.Buttons&entity.ButtonJump != 0 && movement.OnGround {
		movement.Velocity[2] = movement.JumpVelocity
		movement.OnGround = false
	}

	// Apply gravity (if not on ground)
	if !movement.OnGround {
		gravity := console.GetConvarFloat("sv_gravity")
		if gravity == 0 {
			gravity = 800
		}
		movement.Velocity[2] -= gravity * float32(dt)
	}

	// Calculate desired movement
	desiredMove := movement.Velocity.Mul(float32(dt))
	currentPos := transform.Position

	// Use character controller for collision-aware movement
	if charController.HasController() {
		cc := charController.GetController().(*collision.CharacterController)
		moveResult := cc.Move(currentPos, desiredMove)

		transform.Position = moveResult.FinalPosition
		movement.OnGround = moveResult.OnGround

		// Store wall normal for next frame
		if moveResult.HitWall {
			movement.LastWallNormal = moveResult.WallNormal
		} else {
			movement.LastWallNormal = mgl32.Vec3{0, 0, 0}
		}

		// If we just landed, zero out vertical velocity
		if movement.OnGround && movement.Velocity[2] < 0 {
			movement.Velocity[2] = 0
		}
	} else {
		// Fallback: kinematic movement
		console.PrintString(console.LevelWarning, "CharacterController is nil!")
		transform.Position = currentPos.Add(desiredMove)
	}

	// Update entity orientation
	transform.Orientation = mgl32.AnglesToQuat(0, 0, controller.Yaw, mgl32.ZYX)
}

// processNoclipMovement handles noclip mode movement (no collision)
func (s *PlayerMovementSystem) processNoclipMovement(playerEntity ecs.Entity, wishDir mgl32.Vec3, wishSpeed float32, dt float64) {
	transform, _ := ecs.GetComponent[components.Transform](s.ecsWorld, playerEntity)
	movement, _ := ecs.GetComponent[components.CharacterMovement](s.ecsWorld, playerEntity)
	controller, _ := ecs.GetComponent[components.PlayerController](s.ecsWorld, playerEntity)

	// Simple velocity-based movement in noclip
	movement.Velocity = wishDir.Mul(wishSpeed)

	// Apply movement directly without collision detection
	desiredMove := movement.Velocity.Mul(float32(dt))
	transform.Position = transform.Position.Add(desiredMove)
	transform.Orientation = mgl32.AnglesToQuat(0, 0, controller.Yaw, mgl32.ZYX)
}
