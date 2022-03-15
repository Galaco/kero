package client

import (
	"github.com/galaco/kero/client/camera"
	"github.com/galaco/kero/internal/framework/event"
	scene2 "github.com/galaco/kero/internal/framework/scene"
	"github.com/galaco/kero/shared/messages"
	"runtime"
)

type sceneEntities struct {
	cameras      []*camera.Camera
	activeCamera *camera.Camera
}

func (s *sceneEntities) Update(dt float64) {
	// Client controls which camera is active. A camera is loosely bound to a shared camera entity
	if s.activeCamera != nil {
		s.activeCamera.Update(dt)
	}
}

func (s *sceneEntities) BindSharedResources() {
	event.Get().AddListener(messages.TypeLoadingLevelParsed, func(message interface{}) {
		dataScene := message.(*messages.LoadingLevelParsed).Level().(*scene2.StaticScene)
		s.cameras = append(s.cameras, camera.NewCamera(dataScene.Camera))
		s.activeCamera = s.cameras[0]
	})
	event.Get().AddListener(messages.TypeEngineDisconnect, func(e interface{}) {
		s.cameras = make([]*camera.Camera, 0)
		s.activeCamera = nil
		runtime.GC()
	})
}
