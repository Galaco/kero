package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropDynamicOrnament
type PropDynamicOrnament struct {
	entity.EntityBase
	PropRenderableBase
}

// Classname
func (entity PropDynamicOrnament) Classname() string {
	return "prop_dynamic_ornament"
}
