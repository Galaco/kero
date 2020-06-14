package scene

import (
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/library/valve"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
	loader "github.com/galaco/kero/scene/loaders"
)

type Scene struct {
	currentLevel *valve.Bsp
	entities     []entity.IEntity

	listenToInput bool
}

func (s *Scene) Initialize() {
	event.Get().AddListener(messages.TypeChangeLevel, s.onChangeLevel)
	middleware.InputMiddleware().AddListener(messages.TypeKeyRelease, s.onKeyRelease)
	middleware.InputMiddleware().AddListener(messages.TypeMouseMove, s.onMouseMove)
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

func (s *Scene) onChangeLevel(message event.Dispatchable) {
	func(msg *messages.ChangeLevel) {
		level, ents, err := loader.LoadBspMap(filesystem.Get(), msg.LevelName())
		if err != nil {
			console.PrintString(console.LevelError, err.Error())
			return
		}
		s.entities = ents
		// Change level: we must clear the current event queue
		event.Get().CancelPending()
		s.currentLevel = level
		event.Get().Dispatch(messages.NewLoadingLevelParsed(level, ents))
		event.Get().Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateFinished))
	}(message.(*messages.ChangeLevel))
}

func (s *Scene) onKeyRelease(message event.Dispatchable) {
	key := message.(*messages.KeyRelease).Key()
	if key == input.KeyEscape {
		s.listenToInput = !s.listenToInput
	}
}

func (s *Scene) onMouseMove(message event.Dispatchable) {
	if s.currentLevel == nil || s.currentLevel.Camera() == nil {
		return
	}
	msg := message.(*messages.MouseMove)
	s.currentLevel.Camera().Rotate(float32(msg.X), 0, float32(msg.Y))
}

func NewScene() *Scene {
	return &Scene{}
}
