package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropPhysicsMultiplayer
type PropPhysicsMultiplayer struct {
	entity.Entity
}

// Classname
func (entity PropPhysicsMultiplayer) Classname() string {
	return "prop_physics_multiplayer"
}

// PropPath
func (entity PropPhysicsMultiplayer) PropPath() string {
	return entity.ValueForKey("model")
}
