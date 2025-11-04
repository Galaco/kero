package main

import (
	"github.com/galaco/kero/framework/console"
)

type client struct {
}

func (client *client) Update(dt float64) {

}

func NewClient() client {
	// Initialize game convars
	console.AddConvarFloat("sv_gravity", "World gravity", 800.0)
	console.AddConvarFloat("m_sensitivity", "Mouse sensitivity multiplier", 1.0)

	return client{}
}
