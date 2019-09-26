package entity

import (
	"github.com/galaco/kero/valve/entity"
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
