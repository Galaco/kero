package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
)

// PropPhysics
type PropPhysics struct {
	entity.EntityBase
	PropBase
}

// Classname
func (entity PropPhysics) Classname() string {
	return "prop_physics"
}
