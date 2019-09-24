package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
	common "github.com/galaco/lambda-core/game/entity"
)

// PropDoorRotating
type PropDoorRotating struct {
	entity.EntityBase
	common.PropBase
}

// Classname
func (entity PropDoorRotating) Classname() string {
	return "prop_door_rotating"
}
