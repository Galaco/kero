package layer

// Layer represents a GUI rendering layer with explicit z-ordering
type Layer int

const (
	// LayerHUD is the base layer for always-visible HUD elements
	// Does not block input, always rendered
	LayerHUD Layer = 0

	// LayerMenu is for toggle-able menu elements
	// Blocks input when visible, rendered on top of HUD
	LayerMenu Layer = 100

	// LayerModal is for modal overlays like loading screens and dialogs
	// Blocks all input, rendered on top of everything
	LayerModal Layer = 200
)

// Config holds configuration for a specific layer
type Config struct {
	Layer       Layer
	Visible     bool
	BlocksInput bool // When true, prevents lower layers from receiving input
}

// NewConfig creates a new layer configuration with sensible defaults
func NewConfig(layer Layer) *Config {
	return &Config{
		Layer:       layer,
		Visible:     false,
		BlocksInput: layer >= LayerMenu, // Menu and Modal layers block input by default
	}
}
