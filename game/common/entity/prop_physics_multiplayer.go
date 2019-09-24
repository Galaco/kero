package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
	common "github.com/galaco/lambda-core/game/entity"
)

// PropPhysicsMultiplayer
type PropPhysicsMultiplayer struct {
	entity.EntityBase
	common.PropBase
}

// Classname
func (entity PropPhysicsMultiplayer) Classname() string {
	return "prop_physics_multiplayer"
}
