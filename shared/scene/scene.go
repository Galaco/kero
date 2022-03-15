package scene

import (
	"github.com/galaco/kero/internal/framework/console"
	"github.com/galaco/kero/internal/framework/event"
	"github.com/galaco/kero/internal/framework/filesystem"
	scene2 "github.com/galaco/kero/internal/framework/scene"
	messages2 "github.com/galaco/kero/shared/messages"
	loader "github.com/galaco/kero/shared/scene/loaders"
	"runtime"
)

type Scene struct {
	dataScene *scene2.StaticScene

	listenToInput bool
}

func (s *Scene) Initialize() {
	event.Get().AddListener(messages2.TypeChangeLevel, s.onChangeLevel)

	event.Get().AddListener(messages2.TypeEngineDisconnect, func(e interface{}) {
		scene2.CloseCurrentScene()
		runtime.GC()
	})
}

func (s *Scene) Update(dt float64) {
	if s.dataScene == nil {
		return
	}
	if s.dataScene.Camera != nil {
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

func NewScene() *Scene {
	return &Scene{}
}
