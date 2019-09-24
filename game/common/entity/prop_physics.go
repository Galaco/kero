package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
	common "github.com/galaco/lambda-core/game/entity"
)

// PropPhysics
type PropPhysics struct {
	entity.EntityBase
	common.PropBase
}

// Classname
func (entity PropPhysics) Classname() string {
	return "prop_physics"
}
