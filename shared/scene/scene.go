package scene

import (
	middleware "github.com/galaco/kero/client/input"
	"github.com/galaco/kero/internal/framework/console"
	"github.com/galaco/kero/internal/framework/event"
	"github.com/galaco/kero/internal/framework/filesystem"
	"github.com/galaco/kero/internal/framework/input"
	scene2 "github.com/galaco/kero/internal/framework/scene"
	messages2 "github.com/galaco/kero/shared/messages"
	loader "github.com/galaco/kero/shared/scene/loaders"
	"github.com/go-gl/mathgl/mgl32"
	"runtime"
)

type Scene struct {
	dataScene *scene2.StaticScene

	listenToInput bool
}

func (s *Scene) Initialize() {
	event.Get().AddListener(messages2.TypeChangeLevel, s.onChangeLevel)
	middleware.InputMiddleware().AddListener(messages2.TypeKeyRelease, s.onKeyRelease)
	middleware.InputMiddleware().AddListener(messages2.TypeMouseMove, s.onMouseMove)

	event.Get().AddListener(messages2.TypeEngineDisconnect, func(e interface{}) {
		scene2.CloseCurrentScene()
		runtime.GC()
	})
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
		console.PrintString(console.LevelInfo, "Generating Static World...")
		s.dataScene = scene2.LoadStaticSceneFromBsp(filesystem.Get(), level, ents)
		console.PrintString(console.LevelInfo, "Complete!")
		// Change level: we must clear the current event queue
		event.Get().CancelPending()
		event.Get().DispatchLegacy(messages2.NewLoadingLevelParsed(s.dataScene))
		event.Get().Dispatch(messages2.TypeLoadingLevelProgress, messages2.LoadingProgressStateFinished)
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
