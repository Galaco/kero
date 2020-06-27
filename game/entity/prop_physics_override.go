package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropPhysicsOverride
type PropPhysicsOverride struct {
	entity.Entity
	PropRenderableBase
}

// Classname
func (entity PropPhysicsOverride) Classname() string {
	return "prop_physics_override"
}
