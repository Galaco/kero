package server

import (
	"github.com/galaco/kero/server/entitylogic"
)

type Server struct {
	logicController *entitylogic.LogicController
}

func (c *Server) FixedUpdate(dt float64) {
	c.logicController.Update(dt)
}

func (c *Server) Update() {

}

func (c *Server) Initialize() error {

	return nil
}

func NewServer() *Server {
	return &Server{
		logicController: entitylogic.NewLogicController(),
	}
}
