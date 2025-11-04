package entity

import (
	"testing"

	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/graphics"
	"github.com/go-gl/mathgl/mgl32"
)

func TestNewPlayer(t *testing.T) {
	// Initialize console (needed for convars)
	console.AddConvarFloat("sv_gravity", "Test gravity", 800.0)

	spawnPos := mgl32.Vec3{100, 200, 50}
	spawnYaw := mgl32.DegToRad(90)

	player := NewPlayer(spawnPos, spawnYaw)

	if player == nil {
		t.Fatal("NewPlayer returned nil")
	}

	// Verify initial position
	if player.Origin() != spawnPos {
		t.Errorf("Expected position %v, got %v", spawnPos, player.Origin())
	}

	// Verify initial rotation
	if player.yaw != spawnYaw {
		t.Errorf("Expected yaw %f, got %f", spawnYaw, player.yaw)
	}

	// Verify physics body was created
	if player.physics.RigidBody == nil {
		t.Error("Physics rigid body not created")
	}

	// Verify movement properties initialized
	if player.movement.GroundSpeed <= 0 {
		t.Error("Ground speed not initialized")
	}
}

func TestPlayerProcessInput_Movement(t *testing.T) {
	console.AddConvarFloat("sv_gravity", "Test gravity", 800.0)

	player := NewPlayer(mgl32.Vec3{0, 0, 0}, 0)
	initialPos := player.Origin()

	// Process forward movement input
	input := PlayerInput{
		Forward: 1.0,
		Right:   0,
		Yaw:     0,
		Pitch:   0,
		Buttons: 0,
	}

	dt := 0.016 // ~60 FPS

	player.ProcessInput(input, dt)

	// Player should have moved forward
	newPos := player.Origin()
	if newPos == initialPos {
		t.Error("Player did not move after processing input")
	}

	// Velocity should be set
	if player.movement.Velocity.Len() == 0 {
		t.Error("Player velocity is zero after movement input")
	}
}

func TestPlayerProcessInput_Rotation(t *testing.T) {
	console.AddConvarFloat("sv_gravity", "Test gravity", 800.0)

	player := NewPlayer(mgl32.Vec3{0, 0, 0}, 0)
	initialYaw := player.yaw
	initialPitch := player.pitch

	// Verify player starts looking forward (pitch = 180°)
	if player.pitch != CameraPitchCenter {
		t.Errorf("Player should start at CameraPitchCenter, got %f", player.pitch)
	}

	// Process rotation input
	input := PlayerInput{
		Forward: 0,
		Right:   0,
		Yaw:     0.1,  // Rotate right
		Pitch:   0.05, // Look up slightly
		Buttons: 0,
	}

	player.ProcessInput(input, 0.016)

	// Yaw and pitch should have changed
	if player.yaw == initialYaw {
		t.Error("Yaw did not change after rotation input")
	}
	if player.pitch == initialPitch {
		t.Error("Pitch did not change after rotation input")
	}

	// Pitch should have increased (looking up)
	if player.pitch <= initialPitch {
		t.Error("Pitch should have increased after positive pitch input")
	}
}

func TestPlayerProcessInput_Jump(t *testing.T) {
	console.AddConvarFloat("sv_gravity", "Test gravity", 800.0)

	player := NewPlayer(mgl32.Vec3{0, 0, 0}, 0)
	player.movement.OnGround = true

	// Process jump input
	input := PlayerInput{
		Forward: 0,
		Right:   0,
		Yaw:     0,
		Pitch:   0,
		Buttons: ButtonJump,
	}

	player.ProcessInput(input, 0.016)

	// Player should have upward velocity
	if player.movement.Velocity[2] <= 0 {
		t.Error("Player did not jump (no upward velocity)")
	}

	// Player should no longer be on ground
	if player.movement.OnGround {
		t.Error("Player still on ground after jump")
	}
}

func TestPlayerCameraFollow(t *testing.T) {
	console.AddConvarFloat("sv_gravity", "Test gravity", 800.0)

	player := NewPlayer(mgl32.Vec3{100, 200, 50}, 0)

	// Create and attach camera
	camera := graphics.NewCamera(mgl32.DegToRad(90), 16.0/9.0)
	player.SetCamera(camera)

	// Process movement
	input := PlayerInput{
		Forward: 1.0,
		Right:   0,
		Yaw:     0,
		Pitch:   0,
		Buttons: 0,
	}

	player.ProcessInput(input, 0.1)

	// Camera should have moved with player
	playerPos := player.Origin()
	cameraPos := camera.Transform().Translation

	// Camera should be at eye height above player
	expectedCameraZ := playerPos.Z() + float32(PlayerEyeHeight)
	if cameraPos.Z() != expectedCameraZ {
		t.Errorf("Camera Z position incorrect. Expected %f, got %f",
			expectedCameraZ, cameraPos.Z())
	}
}

func TestPlayerPitchClamping(t *testing.T) {
	console.AddConvarFloat("sv_gravity", "Test gravity", 800.0)

	player := NewPlayer(mgl32.Vec3{0, 0, 0}, 0)

	// Try to look way up (beyond limit)
	// Player starts at CameraPitchCenter (180°), add way more than 90°
	input := PlayerInput{
		Pitch: mgl32.DegToRad(200), // Way beyond limit
	}

	player.ProcessInput(input, 0.016)

	// Pitch should be clamped to max (270° = looking straight up)
	if player.pitch > CameraPitchMax {
		t.Errorf("Pitch not clamped. Expected max %f, got %f",
			CameraPitchMax, player.pitch)
	}

	// Reset and try to look way down
	player = NewPlayer(mgl32.Vec3{0, 0, 0}, 0)
	input = PlayerInput{
		Pitch: -mgl32.DegToRad(200), // Way below limit
	}

	player.ProcessInput(input, 0.016)

	// Pitch should be clamped to min (90° = looking straight down)
	if player.pitch < CameraPitchMin {
		t.Errorf("Pitch not clamped. Expected min %f, got %f",
			CameraPitchMin, player.pitch)
	}
}
