package entity

import (
	"github.com/galaco/kero/valve/entity"
)

// PropRagdoll
type PropRagdoll struct {
	entity.EntityBase
	PropRenderableBase
}

// Classname
func (entity PropRagdoll) Classname() string {
	return "prop_ragdoll"
}
