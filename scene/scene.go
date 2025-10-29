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
	"runtime"
)

type Scene struct {
	eventBus       *event.Dispatcher
	fileSystem     filesystem.FileSystem
	sceneManager   *scene2.Manager
	inputMiddleware *middleware.Input

	dataScene *scene2.StaticScene

	listenToInput bool
}

func (s *Scene) Initialize() {
	s.eventBus.AddListener(messages.TypeChangeLevel, s.onChangeLevel)
	s.inputMiddleware.EventBus().AddListener(messages.TypeKeyRelease, s.onKeyRelease)
	s.inputMiddleware.EventBus().AddListener(messages.TypeMouseMove, s.onMouseMove)

	s.eventBus.AddListener(messages.TypeEngineDisconnect, func(e interface{}) {
		s.sceneManager.CloseCurrentScene()
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
		level, ents, err := loader.LoadBspMap(s.fileSystem, s.eventBus, mapName)
		if err != nil {
			console.PrintString(console.LevelError, err.Error())
			return
		}
		console.PrintString(console.LevelInfo, "Generating Static World...")
		s.dataScene = scene2.LoadStaticSceneFromBsp(s.fileSystem, level, ents)
		s.sceneManager.SetCurrentScene(s.dataScene)
		console.PrintString(console.LevelInfo, "Complete!")
		// Change level: we must clear the current event queue
		s.eventBus.CancelPending()
		s.eventBus.DispatchLegacy(messages.NewLoadingLevelParsed(s.dataScene))
		s.eventBus.Dispatch(messages.TypeLoadingLevelProgress, messages.LoadingProgressStateFinished)
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

// NewScene creates a new scene with explicit dependencies
func NewScene(eventBus *event.Dispatcher, fileSystem filesystem.FileSystem, sceneManager *scene2.Manager, inputMiddleware *middleware.Input) *Scene {
	return &Scene{
		eventBus:        eventBus,
		fileSystem:      fileSystem,
		sceneManager:    sceneManager,
		inputMiddleware: inputMiddleware,
	}
}
