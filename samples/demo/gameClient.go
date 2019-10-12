package main

import (
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
)

type client struct {
	camera *graphics3d.Camera
}

func (client *client) Camera() *graphics3d.Camera {
	return client.camera
}

func (client *client) Update(dt float64) {

}

func NewClient() client {
	return client{}
}
