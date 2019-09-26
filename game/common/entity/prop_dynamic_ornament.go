package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
)

// PropDynamicOrnament
type PropDynamicOrnament struct {
	entity.EntityBase
	PropBase
}

// Classname
func (entity PropDynamicOrnament) Classname() string {
	return "prop_dynamic_ornament"
}
