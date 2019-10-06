package cstrike

import (
	"github.com/galaco/kero/game"
	"github.com/galaco/kero/game/common/entity"
	loader "github.com/galaco/kero/valve/entity"
)

// Cstrike
type Cstrike struct {
	client Client
}

// ContentDirectory returns the game content directory relative to the game
// root directory
func (def *Cstrike) ContentDirectory() string {
	return "cstrike"
}

// RegisterEntityClasses loads all Game entity classes into the engine.
func (def *Cstrike) RegisterEntityClasses() {
	loader.RegisterClass(&entity.InfoPlayerStart{})
	loader.RegisterClass(&entity.PropDoorRotating{})
	loader.RegisterClass(&entity.PropDynamic{})
	loader.RegisterClass(&entity.PropDynamicOrnament{})
	loader.RegisterClass(&entity.PropDynamicOverride{})
	loader.RegisterClass(&entity.PropPhysics{})
	loader.RegisterClass(&entity.PropPhysicsMultiplayer{})
	loader.RegisterClass(&entity.PropPhysicsOverride{})
	loader.RegisterClass(&entity.PropRagdoll{})
}

// Client returns the game client
func (def *Cstrike) Client() game.Client {
	return &def.client
}

// NewGameDefinition returns the game definition
func NewGameDefinition() *Cstrike {
	return &Cstrike{
		client: NewClient(),
	}
}
