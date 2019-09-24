package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
	common "github.com/galaco/lambda-core/game/entity"
)

// PropRagdoll
type PropRagdoll struct {
	entity.EntityBase
	common.PropBase
}

// Classname
func (entity PropRagdoll) Classname() string {
	return "prop_ragdoll"
}
