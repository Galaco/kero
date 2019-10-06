package main

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/game"
	"github.com/galaco/kero/systems"
	"github.com/galaco/kero/systems/console"
	"github.com/galaco/kero/systems/gui"
	"github.com/galaco/kero/systems/input"
	"github.com/galaco/kero/systems/renderer"
	"github.com/galaco/kero/systems/scene"
	"time"
)

// Kero
type Kero struct {
	isRunning bool

	context systems.Context
	systems []System
}

func (kero *Kero) RegisterGameDefinitions(def game.Definition) {
	def.RegisterEntityClasses()
}

// RunGameLoop
func (kero *Kero) Start() {
	kero.systems = []System{
		console.NewConsole(),
		input.NewInput(),
		scene.NewScene(),
		renderer.NewRenderer(),
		gui.NewGui(),
	}

	kero.isRunning = true

	kero.initialize()

	dt := 0.0
	startingTime := time.Now().UTC()
	for kero.isRunning {
		event.ProcessMessages()

		kero.context.Client.Update(dt)

		for _, s := range kero.systems {
			s.Update(dt)
		}

		window.CurrentWindow().SwapBuffers()
		graphics.ClearColor(0, 0, 0, 1)
		graphics.ClearAll()

		dt = float64(time.Now().UTC().Sub(startingTime).Nanoseconds()/1000000) / 1000
		startingTime = time.Now().UTC()
	}

	kero.exit()
}

func (kero *Kero) initialize() {
	for i := range kero.systems {
		kero.systems[i].Register(&kero.context)
		event.AddListener(kero.systems[i])
	}
}

func (kero *Kero) exit() {

}

// NewKero
func NewKero(ctx systems.Context) *Kero {
	return &Kero{
		context:   ctx,
		isRunning: false,
	}
}

type System interface {
	Register(ctx *systems.Context)
	Update(dt float64)
	ProcessMessage(message event.Dispatchable)
}
