package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropRagdoll
type PropRagdoll struct {
	entity.Entity
}

// Classname
func (entity PropRagdoll) Classname() string {
	return "prop_ragdoll"
}

// PropPath
func (entity PropRagdoll) PropPath() string {
	return entity.ValueForKey("model")
}
