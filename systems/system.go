package systems

import (
	"github.com/galaco/kero/event"
)

type ISystem interface {
	Register()
	Update(dt float64)
	ProcessMessage(message event.Dispatchable)
}

type System struct {
}

func (s *System) Register() {
}

func (s *System) Update(dt float64) {
}

func (s *System) ProcessMessage(message event.Dispatchable) {
}
