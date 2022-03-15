package main

import (
	loader "github.com/galaco/kero/internal/framework/entity"
	"github.com/galaco/kero/shared/game"
	"github.com/galaco/kero/shared/game/entity"
)

// cstrike
type cstrike struct {
	client client
}

// RegisterEntityClasses loads all Game entity classes into the engine.
func (def *cstrike) RegisterEntityClasses() {
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

// client returns the game client
func (def *cstrike) Client() game.Client {
	return &def.client
}

// NewGameDefinition returns the game definition
func NewGameDefinition() *cstrike {
	return &cstrike{
		client: NewClient(),
	}
}
