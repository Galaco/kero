package client

import (
	"github.com/galaco/kero/client/camera"
	"github.com/galaco/kero/internal/framework/event"
	scene2 "github.com/galaco/kero/internal/framework/scene"
	messages2 "github.com/galaco/kero/shared/messages"
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
	event.Get().AddListener(messages2.TypeLoadingLevelParsed, func(message interface{}) {
		dataScene := message.(*messages2.LoadingLevelParsed).Level().(*scene2.StaticScene)
		s.cameras = append(s.cameras, camera.NewCamera(dataScene.Camera))
		s.activeCamera = s.cameras[0]
	})
}
