package main

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/systems"
	"time"
)

// Engine
type Engine struct {
	isRunning bool

	systems []systems.ISystem
}

// RunGameLoop
func (engine *Engine) RunGameLoop() {
	engine.isRunning = true

	engine.initialize()

	dt := 0.0
	startingTime := time.Now().UTC()
	for engine.isRunning {
		event.Singleton().ProcessMessages()

		for _, s := range engine.systems {
			s.Update(dt)
		}

		window.CurrentWindow().SwapBuffers()
		graphics.ClearColor(0, 0, 0, 1)
		graphics.ClearAll()

		dt = float64(time.Now().UTC().Sub(startingTime).Nanoseconds()/1000000) / 1000
		startingTime = time.Now().UTC()
	}

	engine.exit()
}

func (engine *Engine) initialize() {
	for i := range engine.systems {
		engine.systems[i].Register()
		event.Singleton().RegisterSystem(engine.systems[i])
	}
}

func (engine *Engine) exit() {

}

// NewEngine
func NewEngine(systems ...systems.ISystem) *Engine {
	return &Engine{
		isRunning: false,
		systems:   systems,
	}
}
