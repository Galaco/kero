package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropDoorRotating
type PropDoorRotating struct {
	entity.Entity
	PropRenderableBase
}

// Classname
func (entity PropDoorRotating) Classname() string {
	return "prop_door_rotating"
}
