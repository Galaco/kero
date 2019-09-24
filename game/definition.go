package game

// GameDefinition interface represents a game configuration
type GameDefinition interface {
	// RegisterEntityClasses should setup any game entity classes
	// for use with the engine when loading entdata
	RegisterEntityClasses()
}
