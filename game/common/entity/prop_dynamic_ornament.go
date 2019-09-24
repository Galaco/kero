package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
	common "github.com/galaco/lambda-core/game/entity"
)

// PropDynamicOrnament
type PropDynamicOrnament struct {
	entity.EntityBase
	common.PropBase
}

// Classname
func (entity PropDynamicOrnament) Classname() string {
	return "prop_dynamic_ornament"
}
