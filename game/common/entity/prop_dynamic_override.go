package entity

import (
	"github.com/galaco/kero/valve/entity"
)

// PropDynamicOverride
type PropDynamicOverride struct {
	entity.EntityBase
	PropBase
}

// Classname
func (entity PropDynamicOverride) Classname() string {
	return "prop_dynamic_override"
}
