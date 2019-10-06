package scene

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems"
	loader "github.com/galaco/kero/systems/scene/loaders"
	"github.com/galaco/kero/valve"
)

type Scene struct {
	context      *systems.Context
	currentLevel *valve.Bsp
	entities     []entity.Entity

	listenToInput bool
}

func (s *Scene) Register(ctx *systems.Context) {
	s.context = ctx
}

func (s *Scene) Update(dt float64) {
	if s.currentLevel == nil {
		return
	}
	if s.listenToInput {
		if input.Keyboard().IsKeyPressed(input.KeyW) {
			s.currentLevel.Camera().Forwards(dt)
		}
		if input.Keyboard().IsKeyPressed(input.KeyA) {
			s.currentLevel.Camera().Left(dt)
		}
		if input.Keyboard().IsKeyPressed(input.KeyS) {
			s.currentLevel.Camera().Backwards(dt)
		}
		if input.Keyboard().IsKeyPressed(input.KeyD) {
			s.currentLevel.Camera().Right(dt)
		}

		s.currentLevel.Camera().Update(dt)
	}
	for _, e := range s.entities {
		e.Think(dt)
	}
}

func (s *Scene) ProcessMessage(message event.Dispatchable) {
	switch message.Type() {
	case messages.TypeChangeLevel:
		// LoadLevel
		go func(msg *messages.ChangeLevel) {
			level, ents, err := loader.LoadBspMap(s.context.Filesystem, msg.LevelName())
			if err != nil {
				event.Dispatch(messages.NewConsoleMessage(console.LevelError, err.Error()))
				return
			}
			s.entities = ents
			// Change level: we must clear the current event queue
			event.ClearQueue()
			s.currentLevel = level
			event.Dispatch(messages.NewLoadingLevelParsed(level))
			event.Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateFinished))
		}(message.(*messages.ChangeLevel))
	case messages.TypeKeyRelease:
		key := message.(*messages.KeyRelease).Key()
		if key == input.KeyEscape {
			s.listenToInput = !s.listenToInput
		}
	case messages.TypeMouseMove:
		if s.currentLevel == nil || s.currentLevel.Camera() == nil {
			return
		}
		msg := message.(*messages.MouseMove)
		s.currentLevel.Camera().Rotate(float32(msg.X), 0, float32(msg.Y))
	}
}

func NewScene() *Scene {
	return &Scene{}
}
