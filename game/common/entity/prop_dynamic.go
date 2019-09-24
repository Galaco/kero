package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
	common "github.com/galaco/lambda-core/game/entity"
)

// PropDynamic
type PropDynamic struct {
	entity.EntityBase
	common.PropBase
}

// Classname
func (entity PropDynamic) Classname() string {
	return "prop_dynamic"
}
