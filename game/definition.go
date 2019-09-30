package game

import (
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
)

// Definition interface represents a game configuration
type Definition interface {
	ContentDirectory() string
	// RegisterEntityClasses should setup any game entity classes
	// for use with the engine when loading entdata
	RegisterEntityClasses()
	// Client
	Client() Client
}

type Client interface {
	Camera() *graphics3d.Camera
	Update(float64)
}
