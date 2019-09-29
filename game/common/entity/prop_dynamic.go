package entity

import (
	"github.com/galaco/kero/valve/entity"
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
