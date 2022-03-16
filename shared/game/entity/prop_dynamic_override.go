package entity

import (
	"github.com/galaco/kero/internal/framework/entity"
)

// PropDynamicOverride
type PropDynamicOverride struct {
	entity.Entity
}

// Classname
func (entity PropDynamicOverride) Classname() string {
	return "prop_dynamic_override"
}

// PropPath
func (entity PropDynamicOverride) PropPath() string {
	return entity.ValueForKey("model")
}
