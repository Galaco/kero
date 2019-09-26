package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
)

// PropPhysicsOverride
type PropPhysicsOverride struct {
	entity.EntityBase
	PropBase
}

// Classname
func (entity PropPhysicsOverride) Classname() string {
	return "prop_physics_override"
}
