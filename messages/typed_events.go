package messages

import (
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/scene"
	"github.com/go-gl/mathgl/mgl32"
)

// Type-safe event structs using Go generics.
// These replace the old interface{}-based events with compile-time type checking.

// ============================================================================
// Engine Events
// ============================================================================

// EngineQuitEvent signals that the engine should shut down gracefully.
// This event has no payload data.
type EngineQuitEvent struct{}

// EngineDisconnectEvent signals disconnection from the current game/level.
// This event has no payload data.
type EngineDisconnectEvent struct{}

// ============================================================================
// GUI Events
// ============================================================================

// ChangeLevelEvent requests loading a new map/level.
// The MapName field specifies which map to load.
type ChangeLevelEvent struct {
	MapName string
}

// ============================================================================
// Loading Events
// ============================================================================

// LoadingLevelParsedEvent signals that a level has been successfully parsed and loaded.
// The Level field contains the loaded static scene data.
type LoadingLevelParsedEvent struct {
	Level *scene.StaticScene
}

// LoadingLevelProgressEvent reports the current state of level loading.
// The State field corresponds to LoadingProgressState* constants.
type LoadingLevelProgressEvent struct {
	State int // LoadingProgressStateStarted, LoadingProgressStateBSPParsed, etc.
}

// ============================================================================
// Input Events
// ============================================================================

// KeyPressEvent signals that a keyboard key was pressed.
// The Key field identifies which key was pressed.
type KeyPressEvent struct {
	Key input.Key
}

// KeyReleaseEvent signals that a keyboard key was released.
// The Key field identifies which key was released.
type KeyReleaseEvent struct {
	Key input.Key
}

// MouseMoveEvent signals that the mouse has moved.
// Position contains the current mouse position, Delta contains movement since last event.
type MouseMoveEvent struct {
	Position mgl32.Vec2
	Delta    mgl32.Vec2 // New field: easier to add with type-safe events!
}
