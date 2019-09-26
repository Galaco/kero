package entity

import (
	"github.com/galaco/kero/valve/entity"
)

// InfoPlayerStart
type InfoPlayerStart struct {
	entity.EntityBase
}

// Classname
func (entity InfoPlayerStart) Classname() string {
	return "info_player_start"
}
