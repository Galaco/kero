package scene

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
	loader "github.com/galaco/kero/systems/scene/loaders"
	"github.com/galaco/kero/valve"
)

type Scene struct {
	context *systems.Context
	currentLevel *valve.Bsp
}

func (s *Scene) Register(ctx *systems.Context) {
	s.context = ctx
}

func (s *Scene) Update(dt float64) {

}

func (s *Scene) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeChangeLevel:
		// LoadLevel
		level, err := loader.LoadBspMap(s.context.Filesystem, message.(*messages.ChangeLevel).LevelName())
		if err != nil {
			event.Singleton().Dispatch(messages.NewConsoleMessage(console.LevelError, err.Error()))
			return
		}
		// Change level: we must clear the current event queue
		event.Singleton().ClearQueue()
		s.currentLevel = level
		event.Singleton().Dispatch(messages.NewLoadingLevelParsed(level))
	}
}

func NewScene() *Scene {
	return &Scene{}
}
