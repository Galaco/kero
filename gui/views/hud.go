package views

import (
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/gui/views/menu"
)

// HUD represents the always-visible heads-up display layer
// Contains performance metrics and other always-on UI elements
type HUD struct {
	Performance *menu.Performance
}

// NewHUD creates a new HUD view with the given performance metrics
func NewHUD(performance *menu.Performance) *HUD {
	return &HUD{
		Performance: performance,
	}
}

// Render draws the HUD elements
func (h *HUD) Render(dt float32) {
	// Render performance overlay if enabled by ConVar
	if h.Performance != nil && console.GetConvarBoolean("r_showperf") {
		h.Performance.Render()
	}

	// Future: Add other always-visible HUD elements here
	// - Crosshair
	// - Health/Armor
	// - Ammo counter
	// - Mini-map
	// etc.
}
