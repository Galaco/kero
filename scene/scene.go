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
	// Register typed event listeners (Phase 3)
	event.RegisterTypedEvent(s.inputMiddleware.EventBus(), s.onKeyReleaseTyped)
	event.RegisterTypedEvent(s.inputMiddleware.EventBus(), s.onMouseMoveTyped)
	event.RegisterTypedEvent(s.eventBus, s.onChangeLevelTyped)
	event.RegisterTypedEvent(s.eventBus, func(e messages.EngineDisconnectEvent) {
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

func (s *Scene) onChangeLevelTyped(e messages.ChangeLevelEvent) {
	if s.dataScene != nil {
		// Cleanup
	}

	level, ents, err := loader.LoadBspMap(s.fileSystem, s.eventBus, e.MapName)
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
	// Use typed event dispatch (Phase 3)
	event.DispatchTyped(s.eventBus, messages.LoadingLevelParsedEvent{Level: s.dataScene})
	event.DispatchTyped(s.eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateFinished})
}

func (s *Scene) onKeyReleaseTyped(e messages.KeyReleaseEvent) {
	if e.Key == input.KeyEscape {
		s.listenToInput = !s.listenToInput
	}
}

func (s *Scene) onMouseMoveTyped(e messages.MouseMoveEvent) {
	if s.dataScene == nil || s.dataScene.Camera == nil || !s.listenToInput {
		return
	}
	s.dataScene.Camera.Rotate(e.Position[0], 0, e.Position[1])
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
