package player

import (
	"github.com/galaco/kero/client/entity/camera"
	"github.com/galaco/kero/internal/framework/entity"
)

type CharacterController struct {
	entity.Entity
}

func (controller *CharacterController) BindCamera(c *camera.Camera) {

}

func NewCharacterController() *CharacterController {
	return &CharacterController{}
}
