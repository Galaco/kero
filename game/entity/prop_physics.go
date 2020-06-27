package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropPhysics
type PropPhysics struct {
	entity.Entity
	PropRenderableBase
}

// Classname
func (entity PropPhysics) Classname() string {
	return "prop_physics"
}
