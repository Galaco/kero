package systems

import (
	"github.com/galaco/kero/framework/ecs"
	"github.com/galaco/kero/framework/ecs/components"
	"github.com/galaco/kero/framework/graphics"
)

// PlayerCameraSystem synchronizes camera position and rotation with player entities.
// This system ensures the camera follows the player and matches their view direction.
type PlayerCameraSystem struct {
	ecsWorld *ecs.World
}

// NewPlayerCameraSystem creates a new player camera system
func NewPlayerCameraSystem(world *ecs.World) *PlayerCameraSystem {
	return &PlayerCameraSystem{
		ecsWorld: world,
	}
}

// Update synchronizes all cameras attached to player entities
func (s *PlayerCameraSystem) Update() {
	// Query for entities with Transform + PlayerController + CameraController
	query := s.ecsWorld.Query().
		With(ecs.ComponentTypeTransform).
		With(ecs.ComponentTypePlayerController).
		With(ecs.ComponentTypeCameraController).
		Build()

	entities := query.Entities()
	if len(entities) == 0 {
		return
	}

	// Update each player's camera
	for _, entity := range entities {
		s.updatePlayerCamera(entity)
	}
}

// updatePlayerCamera updates a single player's camera
func (s *PlayerCameraSystem) updatePlayerCamera(entity ecs.Entity) {
	transform, _ := ecs.GetComponent[components.Transform](s.ecsWorld, entity)
	controller, _ := ecs.GetComponent[components.PlayerController](s.ecsWorld, entity)
	cameraController, _ := ecs.GetComponent[components.CameraController](s.ecsWorld, entity)

	// Skip if no camera attached or not active
	if !cameraController.HasCamera() || !cameraController.Active {
		return
	}

	// Get the camera handle
	camera := cameraController.GetCamera().(*graphics.Camera)

	// Update camera rotation from player controller
	// Camera uses a specific coordinate system:
	// - V[0] = yaw (horizontal rotation)
	// - V[1] = roll (always 0 for FPS)
	// - V[2] = pitch (vertical rotation)
	camera.Transform().Orientation.V[0] = controller.Yaw
	camera.Transform().Orientation.V[1] = 0
	camera.Transform().Orientation.V[2] = controller.Pitch

	// Update camera position to player eye position
	// Player transform is at capsule center, camera should be at eye height
	eyePos := transform.Position.Add(cameraController.EyeOffset)
	camera.Transform().Translation = eyePos
}
