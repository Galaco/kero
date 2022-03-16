package entity

import (
	"github.com/galaco/kero/internal/framework/entity"
)

// PropDynamic
type PropDynamic struct {
	entity.Entity
}

// Classname
func (entity PropDynamic) Classname() string {
	return "prop_dynamic"
}

// PropPath
func (entity PropDynamic) PropPath() string {
	return entity.ValueForKey("model")
}
