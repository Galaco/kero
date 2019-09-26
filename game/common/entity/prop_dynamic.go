package entity

import (
	"github.com/galaco/kero/valve/entity"
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
