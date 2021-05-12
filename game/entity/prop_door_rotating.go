package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropDoorRotating
type PropDoorRotating struct {
	entity.Entity
}

// Classname
func (entity PropDoorRotating) Classname() string {
	return "prop_door_rotating"
}

// PropPath
func (entity PropDoorRotating) PropPath() string {
	return entity.ValueForKey("model")
}
