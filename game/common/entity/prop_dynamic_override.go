package entity

import (
	"github.com/galaco/kero/valve/entity"
)

// PropDynamicOverride
type PropDynamicOverride struct {
	entity.EntityBase
	PropRenderableBase
}

// Classname
func (entity PropDynamicOverride) Classname() string {
	return "prop_dynamic_override"
}
