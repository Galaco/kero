package entity

import (
	"github.com/galaco/kero/framework/entity"
)

// PropRagdoll
type PropRagdoll struct {
	entity.Entity
	PropRenderableBase
}

// Classname
func (entity PropRagdoll) Classname() string {
	return "prop_ragdoll"
}
