package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
	common "github.com/galaco/lambda-core/game/entity"
)

// PropDynamicOverride
type PropDynamicOverride struct {
	entity.EntityBase
	common.PropBase
}

// Classname
func (entity PropDynamicOverride) Classname() string {
	return "prop_dynamic_override"
}
