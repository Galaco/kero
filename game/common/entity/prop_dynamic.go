package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
)

// PropDynamic
type PropDynamic struct {
	entity.EntityBase
	PropBase
}

// Classname
func (entity PropDynamic) Classname() string {
	return "prop_dynamic"
}
