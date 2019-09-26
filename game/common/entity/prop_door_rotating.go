package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
)

// PropDoorRotating
type PropDoorRotating struct {
	entity.EntityBase
	PropBase
}

// Classname
func (entity PropDoorRotating) Classname() string {
	return "prop_door_rotating"
}
