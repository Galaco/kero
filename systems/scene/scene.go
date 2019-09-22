package scene

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/event/message"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/valve"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
	loader "github.com/galaco/kero/systems/scene/loaders"
)

type Scene struct {
	systems.System

	currentLevel *valve.Bsp
}

func (s *Scene) Update(dt float64) {

}

func (s *Scene) ProcessMessage(message message.Dispatchable) {
	switch message.Type() {
	case messages.TypeChangeLevel:
		// LoadLevel
		level,err := loader.LoadBspMap(message.(*messages.ChangeLevel).LevelName())
		if err != nil {
			event.Singleton().Dispatch(messages.NewConsoleMessage(console.LevelError, err.Error()))
			return
		}
		s.currentLevel = level
		event.Singleton().Dispatch(messages.NewLoadingLevelParsed(level))
	}
}

func NewScene() *Scene {
	return &Scene{}
}
