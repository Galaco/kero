package entity

import (
	"github.com/galaco/kero/internal/framework/entity"
)

// PropPhysicsOverride
type PropPhysicsOverride struct {
	entity.Entity
}

// Classname
func (entity PropPhysicsOverride) Classname() string {
	return "prop_physics_override"
}

// PropPath
func (entity PropPhysicsOverride) PropPath() string {
	return entity.ValueForKey("model")
}
