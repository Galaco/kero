package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropDynamicOrnament
type PropDynamicOrnament struct {
	entity.Entity
}

// Classname
func (entity PropDynamicOrnament) Classname() string {
	return "prop_dynamic_ornament"
}

// PropPath
func (entity PropDynamicOrnament) PropPath() string {
	return entity.ValueForKey("model")
}
