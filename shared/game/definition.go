package game

// Definition interface represents a game configuration
type Definition interface {
	// RegisterEntityClasses should setup any game entity classes
	// for use with the engine when loading entdata
	RegisterEntityClasses()
	// Client
	Client() Client
}

type Client interface {
	Update(float64)
}
