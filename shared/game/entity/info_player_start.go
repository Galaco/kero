package entity

import (
	"github.com/galaco/kero/internal/framework/entity"
)

// InfoPlayerStart
type InfoPlayerStart struct {
	entity.Entity
}

// Classname
func (entity InfoPlayerStart) Classname() string {
	return "info_player_start"
}
