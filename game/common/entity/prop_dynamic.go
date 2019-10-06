package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropDynamic
type PropDynamic struct {
	entity.EntityBase
	PropRenderableBase
}

// Classname
func (entity PropDynamic) Classname() string {
	return "prop_dynamic"
}
