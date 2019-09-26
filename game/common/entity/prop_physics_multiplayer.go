package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
)

// PropPhysicsMultiplayer
type PropPhysicsMultiplayer struct {
	entity.EntityBase
	PropBase
}

// Classname
func (entity PropPhysicsMultiplayer) Classname() string {
	return "prop_physics_multiplayer"
}
