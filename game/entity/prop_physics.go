package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropPhysics
type PropPhysics struct {
	entity.Entity
}

// Classname
func (entity PropPhysics) Classname() string {
	return "prop_physics"
}

// PropPath
func (entity PropPhysics) PropPath() string {
	return entity.ValueForKey("model")
}
