package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
)

// PropRagdoll
type PropRagdoll struct {
	entity.EntityBase
	PropBase
}

// Classname
func (entity PropRagdoll) Classname() string {
	return "prop_ragdoll"
}
