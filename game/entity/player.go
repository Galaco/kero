package entity

import (
	"fmt"
	"math"

	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/physics/collision"
	"github.com/go-gl/mathgl/mgl32"
)

// Source Engine player constants
const (
	PlayerHeight     = float64(72) // Total player height in units
	PlayerRadius     = float64(16) // Capsule radius
	PlayerEyeHeight  = float64(64) // Eye position from feet
	PlayerStepHeight = float64(18) // Maximum stair climb height
	PlayerMass       = float32(0)  // Kinematic (zero mass)
)

// Camera coordinate system constants
// The camera uses a coordinate system where pitch is centered at π (180°)
const (
	CameraPitchCenter = float32(math.Pi)         // Forward (180°)
	CameraPitchMin    = float32(math.Pi / 2)     // Down (90°)
	CameraPitchMax    = float32(math.Pi * 3 / 2) // Up (270°)
	CameraMaxPitchDev = float32(math.Pi / 2)     // ±90° from center
)

// PlayerInput represents player input commands.
// This uses the command pattern - storing intent rather than direct actions.
// This is critical for future multiplayer: clients send inputs, server processes them.
type PlayerInput struct {
	Forward float32 // -1.0 to 1.0 (S to W)
	Right   float32 // -1.0 to 1.0 (A to D)
	Up      float32 // -1.0 to 1.0 (Crouch to Jump)
	Yaw     float32 // Radians (horizontal rotation)
	Pitch   float32 // Radians (vertical rotation)
	Buttons uint32  // Bitfield for discrete actions (jump, crouch, use, etc.)
}

// Button bit flags for PlayerInput.Buttons
const (
	ButtonJump   = 1 << 0
	ButtonCrouch = 1 << 1
	ButtonUse    = 1 << 2
	ButtonAttack = 1 << 3
	ButtonReload = 1 << 4
)

// PlayerMovement handles movement physics and state.
// This component is responsible for velocity, ground detection, and movement calculation.
type PlayerMovement struct {
	Velocity       mgl32.Vec3 // Current velocity (units/second)
	GroundSpeed    float32    // Maximum ground movement speed
	AirSpeed       float32    // Maximum air movement speed
	JumpVelocity   float32    // Upward velocity applied on jump
	Friction       float32    // Ground friction coefficient
	Acceleration   float32    // Ground acceleration (units/sec^2)
	OnGround       bool       // Whether player is on ground
	LastWallNormal mgl32.Vec3 // Normal of wall we're currently against (for preventing re-acceleration into wall)
}

// PlayerPhysics wraps the physics representation of the player.
// Uses a capsule collision shape for smooth movement over geometry.
type PlayerPhysics struct {
	Height              float64                          // Total capsule height
	Radius              float64                          // Capsule radius
	StepHeight          float64                          // Maximum stair climb
	RigidBody           collision.RigidBody              // Wrapper for Bullet rigid body
	CharacterController *collision.CharacterController   // Character controller for collision
}

// Player represents the player entity in the game world.
// The player is a special entity that:
// - Responds to player input
// - Controls camera position/rotation
// - Has physics-based movement with collision
// - Serves as the focal point for gameplay
type Player struct {
	entity.Entity
	input    PlayerInput
	movement PlayerMovement
	physics  PlayerPhysics
	camera   *graphics.Camera // Camera following this player

	// Rotation state (in radians)
	yaw   float32 // Horizontal rotation
	pitch float32 // Vertical rotation (clamped)
}

// Classname returns the entity classname for registration
func (p *Player) Classname() string {
	return "player"
}

// Think is called every frame to update player state.
// This is where we gather input and update the player.
func (p *Player) Think(dt float64) {
	// Input gathering happens in scene layer for now
	// Movement is applied via ProcessInput()
}

// ProcessInput applies player input to movement in a deterministic way.
// This method must be deterministic: same input + same state = same output.
// This is critical for client-side prediction in multiplayer.
func (p *Player) ProcessInput(input PlayerInput, dt float64) {
	// Debug: Log when ProcessInput is called
	//console.PrintString(console.LevelInfo, fmt.Sprintf("ProcessInput: dt=%.3f, forward=%.2f, right=%.2f", dt, input.Forward, input.Right))

	// ========================================
	// ROTATION (instantaneous, no dt needed)
	// ========================================
	// Update rotation from input
	p.yaw += input.Yaw
	p.pitch += input.Pitch

	// Clamp pitch to camera's coordinate system range (90° to 270°, centered at 180°)
	if p.pitch > CameraPitchMax {
		p.pitch = CameraPitchMax
	}
	if p.pitch < CameraPitchMin {
		p.pitch = CameraPitchMin
	}

	// Update camera rotation immediately (even if dt=0 for mouse look)
	if p.camera != nil {
		p.camera.Transform().Orientation.V[0] = p.yaw
		p.camera.Transform().Orientation.V[1] = 0
		p.camera.Transform().Orientation.V[2] = p.pitch
	}

	// ========================================
	// MOVEMENT (requires dt for integration)
	// ========================================
	if dt <= 0 {
		return // No movement without time
	}

	// Clear previous frame's collision debug data
	if p.physics.CharacterController != nil {
		p.physics.CharacterController.ClearDebugHits()
	}

	// Check for noclip mode
	noclipEnabled := console.GetConvarBoolean("sv_noclip")

	// Calculate forward and right vectors based on camera orientation
	var forward, right mgl32.Vec3

	if noclipEnabled {
		// Noclip: Move in camera direction (includes vertical component from pitch)
		// Convert pitch from camera space (centered at π) to standard space
		actualPitch := p.pitch - CameraPitchCenter

		// Forward vector includes vertical component
		// Since actualPitch = pitch - π, we need sin(actualPitch) = -sin(pitch)
		// This matches the camera's direction vector coordinate system
		forward = mgl32.Vec3{
			-float32(math.Sin(float64(p.yaw)) * math.Cos(float64(actualPitch))),
			-float32(math.Cos(float64(p.yaw)) * math.Cos(float64(actualPitch))),
			-float32(math.Sin(float64(actualPitch))),
		}

		// Right vector is always horizontal (perpendicular to yaw)
		right = mgl32.Vec3{
			-float32(math.Cos(float64(p.yaw))),
			float32(math.Sin(float64(p.yaw))),
			0,
		}
	} else {
		// Normal movement: Horizontal only
		// Match camera's coordinate system: -Y is forward at yaw=0
		forward = mgl32.Vec3{
			-float32(math.Sin(float64(p.yaw))),
			-float32(math.Cos(float64(p.yaw))),
			0, // Horizontal movement only
		}
		right = mgl32.Vec3{
			float32(math.Sin(float64(p.yaw) - math.Pi/2)),
			float32(math.Cos(float64(p.yaw) - math.Pi/2)),
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

	// Normalize if we have input (don't normalize zero vector)
	wishSpeed := float32(0)
	if wishDir.Len() > 0.01 {
		wishDir = wishDir.Normalize()

		// If we have a wall normal from previous frame, clip wishDir along it
		// This prevents re-accelerating into walls every frame
		if p.movement.LastWallNormal.Len() > 0.01 {
			dotProduct := wishDir.Dot(p.movement.LastWallNormal)
			// Only clip if moving into the wall (dot < 0)
			if dotProduct < 0 {
				// Clip wishDir to be parallel to wall
				wishDir = wishDir.Sub(p.movement.LastWallNormal.Mul(dotProduct))
				if wishDir.Len() > 0.01 {
					wishDir = wishDir.Normalize()
				}
				console.PrintInterface(console.LevelInfo, fmt.Sprintf("Clipped wishDir against wall"))
			}
		}

		if noclipEnabled {
			wishSpeed = console.GetConvarFloat("sv_noclip_speed")
		} else if p.movement.OnGround {
			wishSpeed = p.movement.GroundSpeed
		} else {
			wishSpeed = p.movement.AirSpeed
		}
	}

	// Noclip mode: Direct movement without physics
	if noclipEnabled {
		// Simple velocity-based movement in noclip
		// Direct velocity toward wish direction
		p.movement.Velocity = wishDir.Mul(wishSpeed)

		// Apply movement directly without collision detection
		desiredMove := p.movement.Velocity.Mul(float32(dt))
		currentPos := p.physics.RigidBody.GetTranslation()
		newPos := currentPos.Add(desiredMove)

		// Update physics body transform
		transform := mgl32.Translate3D(newPos.X(), newPos.Y(), newPos.Z())
		p.physics.RigidBody.SetTransform(transform)

		// Update entity transform
		p.Transform().Translation = newPos
		p.Transform().Orientation = mgl32.AnglesToQuat(0, 0, p.yaw, mgl32.ZYX)

		// Update camera position
		if p.camera != nil {
			eyeOffsetFromCenter := float32(PlayerEyeHeight - PlayerHeight/2)
			eyePos := newPos.Add(mgl32.Vec3{0, 0, eyeOffsetFromCenter})
			p.camera.Transform().Translation = eyePos
		}

		return // Skip normal physics processing
	}

	// Apply movement (simple velocity-based for Phase 1)
	if p.movement.OnGround {
		// Ground movement: accelerate toward wish direction
		targetVel := wishDir.Mul(wishSpeed)
		// Simple acceleration
		p.movement.Velocity = p.movement.Velocity.Add(
			targetVel.Sub(p.movement.Velocity).Mul(float32(dt) * p.movement.Acceleration))

		// Apply friction when not moving
		if wishDir.Len() < 0.01 {
			friction := 1.0 - (p.movement.Friction * float32(dt))
			if friction < 0 {
				friction = 0
			}
			p.movement.Velocity = p.movement.Velocity.Mul(friction)
		}

		// Flatten velocity on ground (Source Engine behavior)
		// Keep movement horizontal - no vertical component when on ground
		p.movement.Velocity[2] = 0
	} else {
		// Air movement: limited air control
		airAccel := p.movement.Acceleration * 0.2 // Reduced air control
		targetVel := wishDir.Mul(wishSpeed)
		p.movement.Velocity = p.movement.Velocity.Add(
			targetVel.Sub(p.movement.Velocity).Mul(float32(dt) * airAccel))
	}

	// Handle jump
	if input.Buttons&ButtonJump != 0 && p.movement.OnGround {
		p.movement.Velocity[2] = p.movement.JumpVelocity
		p.movement.OnGround = false
	}

	// Apply gravity (if not on ground)
	if !p.movement.OnGround {
		gravity := console.GetConvarFloat("sv_gravity")
		if gravity == 0 {
			gravity = 800 // Default Source Engine gravity
		}
		p.movement.Velocity[2] -= gravity * float32(dt)
	}

	// Calculate desired movement
	desiredMove := p.movement.Velocity.Mul(float32(dt))
	currentPos := p.physics.RigidBody.GetTranslation()

	// Debug logging
	if wishDir.Len() > 0.01 {
		console.PrintString(console.LevelInfo, fmt.Sprintf("Movement: vel=(%.1f,%.1f,%.1f) pos=(%.1f,%.1f,%.1f) desiredMove=(%.3f,%.3f,%.3f) onGround=%v",
			p.movement.Velocity[0], p.movement.Velocity[1], p.movement.Velocity[2],
			currentPos[0], currentPos[1], currentPos[2],
			desiredMove[0], desiredMove[1], desiredMove[2],
			p.movement.OnGround))
	}

	// Use character controller for collision-aware movement
	var newPos mgl32.Vec3
	if p.physics.CharacterController != nil {
		// Physics-based movement with collision detection
		moveResult := p.physics.CharacterController.Move(currentPos, desiredMove)
		newPos = moveResult.FinalPosition
		p.movement.OnGround = moveResult.OnGround

		// Debug: Check if we actually moved
		if wishDir.Len() > 0.01 {
			distance := newPos.Sub(currentPos).Len()
			console.PrintString(console.LevelInfo, fmt.Sprintf("Move result: distance=%.3f, hitWall=%v, onGround=%v, newPos=(%.1f,%.1f,%.1f)",
				distance, moveResult.HitWall, moveResult.OnGround,
				newPos[0], newPos[1], newPos[2]))
		}

		// If we hit a wall, store the wall normal for next frame
		// This prevents re-accelerating into the wall
		if moveResult.HitWall {
			p.movement.LastWallNormal = moveResult.WallNormal
			console.PrintInterface(console.LevelInfo, fmt.Sprintf("HitWall! Stored wall normal: [%.2f, %.2f, %.2f]",
				moveResult.WallNormal[0], moveResult.WallNormal[1], moveResult.WallNormal[2]))
		} else {
			// Clear wall normal if we're no longer hitting a wall
			p.movement.LastWallNormal = mgl32.Vec3{0, 0, 0}
		}

		// If we just landed, zero out vertical velocity
		if p.movement.OnGround && p.movement.Velocity[2] < 0 {
			p.movement.Velocity[2] = 0
		}
	} else {
		// Fallback: kinematic movement (Phase 1 behavior)
		console.PrintString(console.LevelWarning, "Using fallback movement - CharacterController is nil!")
		newPos = currentPos.Add(desiredMove)

		// Basic ground detection fallback
		if p.movement.Velocity[2] <= 0 && newPos.Z() < 100 {
			p.movement.OnGround = true
		}
	}

	// Update physics body transform
	transform := mgl32.Translate3D(newPos.X(), newPos.Y(), newPos.Z())
	p.physics.RigidBody.SetTransform(transform)

	// Update entity transform
	p.Transform().Translation = newPos
	p.Transform().Orientation = mgl32.AnglesToQuat(0, 0, p.yaw, mgl32.ZYX)

	// Update camera to follow player (position only, rotation handled above)
	if p.camera != nil {
		// Camera position at eye height (64 units from feet)
		// newPos is capsule center (feet + 36), so eye offset from center = 64 - 36 = 28
		eyeOffsetFromCenter := float32(PlayerEyeHeight - PlayerHeight/2) // 64 - 36 = 28
		eyePos := newPos.Add(mgl32.Vec3{0, 0, eyeOffsetFromCenter})
		p.camera.Transform().Translation = eyePos
	}
}

// SetCamera assigns a camera to follow this player
func (p *Player) SetCamera(camera *graphics.Camera) {
	p.camera = camera
}

// GetCamera returns the camera following this player
func (p *Player) GetCamera() *graphics.Camera {
	return p.camera
}

// GetCharacterController returns the character controller for debug visualization
func (p *Player) GetCharacterController() *collision.CharacterController {
	return p.physics.CharacterController
}

// GetPosition returns the current player position (capsule center)
func (p *Player) GetPosition() mgl32.Vec3 {
	if p.physics.RigidBody != nil {
		return p.physics.RigidBody.GetTranslation()
	}
	return p.Transform().Translation
}

// SetPosition sets the player's position in world space
func (p *Player) SetPosition(pos mgl32.Vec3) {
	p.Transform().Translation = pos
	if p.physics.RigidBody != nil {
		transform := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
		p.physics.RigidBody.SetTransform(transform)
	}
}

// SetRotation sets the player's yaw rotation
func (p *Player) SetRotation(yaw float32) {
	p.yaw = yaw
	p.Transform().Orientation = mgl32.AnglesToQuat(0, 0, yaw, mgl32.ZYX)
}

// InitializePhysics initializes the player's physics collision with the physics world.
// This must be called after the Bullet physics world is created to enable collision detection.
// Creates a character controller for collision-aware movement.
//
// The world parameter should be a bullet.BulletDynamicWorldHandle.
// We use interface{} to avoid import cycles with the physics package.
func (p *Player) InitializePhysics(world interface{}, capsuleShape interface{}) {
	console.PrintString(console.LevelInfo, "Player.InitializePhysics() called")

	// Create character controller using the provided world and capsule shape
	// The collision package will handle type assertions from interface{}
	p.physics.CharacterController = collision.NewCharacterControllerFromInterfaces(
		world,
		capsuleShape,
		p.physics.Height,
		p.physics.Radius,
		p.physics.StepHeight,
	)

	if p.physics.CharacterController == nil {
		console.PrintString(console.LevelError, "Failed to create CharacterController - type assertion failed!")
	} else {
		console.PrintString(console.LevelSuccess, "CharacterController created successfully!")
	}
}

// NewPlayer creates a new player entity at the specified position.
// The player is initialized with default movement properties and a capsule collision shape.
// The position parameter represents the player's feet position (ground level).
func NewPlayer(position mgl32.Vec3, yaw float32) *Player {
	// Convert from feet position to capsule center position
	// In Source Engine, player origin is at feet. In Bullet, capsule origin is at center.
	// Offset upward by half the capsule height to get the center position.
	capsuleCenterOffset := float32(PlayerHeight / 2) // 36 units
	capsuleCenterPos := position.Add(mgl32.Vec3{0, 0, capsuleCenterOffset})

	// Create base entity (using capsule center position for physics consistency)
	baseEntity := entity.NewEntityBase("player", "", graphics.Transform{
		Translation: capsuleCenterPos,
		Orientation: mgl32.AnglesToQuat(0, 0, yaw, mgl32.ZYX),
	})

	// Create capsule physics body (kinematic)
	// Capsule height = cylindrical portion (we want total height of 72, minus 2*radius for hemispheres)
	capsuleHeight := PlayerHeight - (2 * PlayerRadius)
	capsuleBody := collision.NewCapsuleHull(PlayerRadius, capsuleHeight, PlayerMass)

	// Set initial transform (using capsule center position)
	initialTransform := mgl32.Translate3D(capsuleCenterPos.X(), capsuleCenterPos.Y(), capsuleCenterPos.Z())
	capsuleBody.SetTransform(initialTransform)

	player := &Player{
		Entity: *baseEntity,
		input:  PlayerInput{},
		movement: PlayerMovement{
			Velocity:     mgl32.Vec3{0, 0, 0},
			GroundSpeed:  320.0, // Source Engine default
			AirSpeed:     30.0,
			JumpVelocity: 270.0, // Source Engine default
			Friction:     4.0,
			Acceleration: 10.0,
			OnGround:     true,
		},
		physics: PlayerPhysics{
			Height:     PlayerHeight,
			Radius:     PlayerRadius,
			StepHeight: PlayerStepHeight,
			RigidBody:  capsuleBody,
		},
		yaw:   yaw,
		pitch: CameraPitchCenter, // Start looking forward (180°)
	}

	return player
}
