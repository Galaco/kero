package scene

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
	loader "github.com/galaco/kero/systems/scene/loaders"
	"github.com/galaco/kero/valve"
	"github.com/galaco/kero/valve/entity"
)

type Scene struct {
	context *systems.Context
	currentLevel *valve.Bsp
	entities []entity.Entity
}

func (s *Scene) Register(ctx *systems.Context) {
	s.context = ctx
}

func (s *Scene) Update(dt float64) {
	for _,e := range s.entities {
		e.Think(dt)
	}
}

func (s *Scene) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeChangeLevel:
		// LoadLevel
		go func(msg *messages.ChangeLevel) {
			level, ents, err := loader.LoadBspMap(s.context.Filesystem, msg.LevelName())
			s.entities = ents
			if err != nil {
				event.Dispatch(messages.NewConsoleMessage(console.LevelError, err.Error()))
				return
			}
			// Change level: we must clear the current event queue
			event.ClearQueue()
			s.currentLevel = level
			event.Dispatch(messages.NewLoadingLevelParsed(level))
			event.Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateFinished))
		}(message.(*messages.ChangeLevel))
	}
}

func NewScene() *Scene {
	return &Scene{}
}
