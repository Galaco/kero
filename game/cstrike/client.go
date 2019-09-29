package cstrike

import (
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
)

type Client struct {
	camera *graphics3d.Camera
}

func (client *Client) Camera() *graphics3d.Camera {
	return client.camera
}

func (client *Client) Update(dt float64) {

}

func NewClient() Client {
	return Client{}
}
