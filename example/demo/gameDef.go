package main

import (
	loader "github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/game"
	common "github.com/galaco/kero/game/entity"
)

// cstrike
type cstrike struct {
	client client
}

// RegisterEntityClasses loads all Game entity classes into the engine.
func (def *cstrike) RegisterEntityClasses() {
	loader.RegisterClass(&common.InfoPlayerStart{})
	loader.RegisterClass(&common.PropDoorRotating{})
	loader.RegisterClass(&common.PropDynamic{})
	loader.RegisterClass(&common.PropDynamicOrnament{})
	loader.RegisterClass(&common.PropDynamicOverride{})
	loader.RegisterClass(&common.PropPhysics{})
	loader.RegisterClass(&common.PropPhysicsMultiplayer{})
	loader.RegisterClass(&common.PropPhysicsOverride{})
	loader.RegisterClass(&common.PropRagdoll{})
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
