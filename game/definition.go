package game

// Definition interface represents a game configuration
type Definition interface {
	// ContentDirectory is the game content directory (e.g. hl2, cstrike, csgo)
	ContentDirectory() string
	// RegisterEntityClasses should setup any game entity classes
	// for use with the engine when loading entdata
	RegisterEntityClasses()
	// Client
	Client() Client
}

type Client interface {
	Update(float64)
}
