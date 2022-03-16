package entitylogic

import "github.com/galaco/kero/shared/scene"

type LogicController struct {
}

func (s *LogicController) Update(dt float64) {
	sc := scene.CurrentScene()
	if sc == nil || sc.Entities() == nil {
		return
	}

	for _, e := range sc.Entities() {
		e.Think(dt)
	}

	sc.Camera().Update(dt)

}

func NewLogicController() *LogicController {
	return &LogicController{}
}
