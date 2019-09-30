package cstrike

import (
	"github.com/galaco/kero/game"
	"github.com/galaco/kero/game/common/entity"
	loader "github.com/galaco/kero/valve/entity"
)

// CounterstrikeSource
type CounterstrikeSource struct {
	client Client
}

func (def *CounterstrikeSource) ContentDirectory() string {
	return "cstrike"
}

// RegisterEntityClasses loads all Game entity classes into the engine.
func (def *CounterstrikeSource) RegisterEntityClasses() {
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

func (def *CounterstrikeSource) Client() game.Client {
	return &def.client
}

func NewGameDefinition() *CounterstrikeSource {
	return &CounterstrikeSource{
		client: NewClient(),
	}
}
