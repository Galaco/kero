package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
	common "github.com/galaco/lambda-core/game/entity"
)

// PropPhysicsOverride
type PropPhysicsOverride struct {
	entity.EntityBase
	common.PropBase
}

// Classname
func (entity PropPhysicsOverride) Classname() string {
	return "prop_physics_override"
}
