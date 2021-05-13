package scene

import (
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/input"
	scene2 "github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/middleware"
	loader "github.com/galaco/kero/scene/loaders"
	"github.com/go-gl/mathgl/mgl32"
)

type Scene struct {
	dataScene *scene2.StaticScene

	listenToInput bool
}

func (s *Scene) Initialize() {
	event.Get().AddListener(messages.TypeChangeLevel, s.onChangeLevel)
	middleware.InputMiddleware().AddListener(messages.TypeKeyRelease, s.onKeyRelease)
	middleware.InputMiddleware().AddListener(messages.TypeMouseMove, s.onMouseMove)
}

func (s *Scene) Update(dt float64) {
	if s.dataScene == nil {
		return
	}
	if s.listenToInput {
		if input.Keyboard().IsKeyPressed(input.KeyW) {
			s.dataScene.Camera.Forwards(dt)
		}
		if input.Keyboard().IsKeyPressed(input.KeyA) {
			s.dataScene.Camera.Left(dt)
		}
		if input.Keyboard().IsKeyPressed(input.KeyS) {
			s.dataScene.Camera.Backwards(dt)
		}
		if input.Keyboard().IsKeyPressed(input.KeyD) {
			s.dataScene.Camera.Right(dt)
		}

		s.dataScene.Camera.Update(dt)
	}

	for _, e := range s.dataScene.Entities {
		e.Think(dt)
	}
}

func (s *Scene) onChangeLevel(message interface{}) {
	if s.dataScene != nil {
		// Cleanup

	}

	func(mapName string) {
		level, ents, err := loader.LoadBspMap(filesystem.Get(), mapName)
		if err != nil {
			console.PrintString(console.LevelError, err.Error())
			return
		}
		s.dataScene = scene2.LoadStaticSceneFromBsp(filesystem.Get(), level, ents)
		// Change level: we must clear the current event queue
		event.Get().CancelPending()
		event.Get().DispatchLegacy(messages.NewLoadingLevelParsed(s.dataScene))
		event.Get().Dispatch(messages.TypeLoadingLevelProgress, messages.LoadingProgressStateFinished)
	}(message.(string))
}

func (s *Scene) onKeyRelease(message interface{}) {
	key := message.(input.Key)
	if key == input.KeyEscape {
		s.listenToInput = !s.listenToInput
	}
}

func (s *Scene) onMouseMove(message interface{}) {
	if s.dataScene == nil || s.dataScene.Camera == nil || !s.listenToInput {
		return
	}
	msg := message.(mgl32.Vec2)
	s.dataScene.Camera.Rotate(msg[0], 0, msg[1])
}

func NewScene() *Scene {
	return &Scene{}
}
