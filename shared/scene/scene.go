package scene

import (
	"runtime"

	"github.com/galaco/kero/internal/framework/console"
	"github.com/galaco/kero/internal/framework/entity"
	"github.com/galaco/kero/internal/framework/event"
	"github.com/galaco/kero/internal/framework/filesystem"
	"github.com/galaco/kero/internal/framework/graphics"
	scene2 "github.com/galaco/kero/internal/framework/scene"
	"github.com/galaco/kero/shared/messages"
	loader "github.com/galaco/kero/shared/scene/loaders"
)

type Scene struct {
	dataScene *scene2.StaticScene
}

func (s *Scene) Initialize() {
	event.Get().AddListener(messages.TypeChangeLevel, s.onChangeLevel)

	event.Get().AddListener(messages.TypeEngineDisconnect, func(e interface{}) {
		scene2.CloseCurrentScene()
		s.dataScene = nil
		runtime.GC()
	})
}

func (s *Scene) Entities() []entity.IEntity {
	if s.dataScene == nil {
		return nil
	}
	return s.dataScene.Entities
}

func (s *Scene) Camera() *graphics.Camera {
	if s.dataScene == nil {
		return nil
	}
	return s.dataScene.Camera
}

func (s *Scene) Raw() *scene2.StaticScene {
	return s.dataScene
}

func (s *Scene) onChangeLevel(message interface{}) {
	if s.dataScene != nil {
		// Cleanup
		s.dataScene = nil
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
		event.Get().DispatchLegacy(messages.NewLoadingLevelParsed(s.dataScene))
		event.Get().Dispatch(messages.TypeLoadingLevelProgress, messages.LoadingProgressStateFinished)
	}(message.(string))
}

// Only 1 scene can be active at once
var scene Scene

func CurrentScene() *Scene {
	return &scene
}
