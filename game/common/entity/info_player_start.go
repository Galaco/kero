package entity

import (
	"github.com/galaco/kero/framework/valve/entity"
)

// InfoPlayerStart
type InfoPlayerStart struct {
	entity.EntityBase
}

// Classname
func (entity InfoPlayerStart) Classname() string {
	return "info_player_start"
}
