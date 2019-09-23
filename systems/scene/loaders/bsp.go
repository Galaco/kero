package loader

import (
	"github.com/galaco/bsp"
	"github.com/galaco/kero/valve"
)

// LoadBspMap is the gateway into loading the core static level. Entities are loaded
// elsewhere
// It loads in the following order:
// BSP Geometry
// BSP Materials
// StaticProps (materials loaded as required)
func LoadBspMap(filename string) (*valve.Bsp, error) {
	file, err := bsp.ReadFromFile(filename)
	if err != nil {
		return nil, err
	}
	// Load the static bsp world
	level, err := valve.LoadBSPWorld(file)
	if err != nil {
		return nil,err
	}

	// Load staticprops
	valve.LoadStaticProps(file)

	// Load entities

	// Load visibility optimisations
	level.AddVisibility(valve.LoadVisData(file))

	return level, err
}
