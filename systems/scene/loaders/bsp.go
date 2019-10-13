package loader

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/valve"
)

// LoadBspMap is the gateway into loading the core static level. Entities are loaded
// elsewhere
// It loads in the following order:
// BSP Geometry
// BSP Materials
// StaticProps (materials loaded as required)
func LoadBspMap(fs filesystem.FileSystem, filename string) (*valve.Bsp, []entity.Entity, error) {
	event.Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateStarted))
	file, err := bsp.ReadFromFile(filename)
	if err != nil {
		event.Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateError))
		return nil, nil, err
	}
	event.Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateBSPParsed))
	fs.RegisterPakFile(file.Lump(bsp.LumpPakfile).(*lumps.Pakfile))
	// Load the static bsp world
	level, err := valve.LoadBSPWorld(fs, file)
	if err != nil {
		event.Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateError))
		return nil, nil, err
	}
	event.Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateGeometryLoaded))

	// Load staticprops
	level.StaticPropDictionary, level.StaticProps = valve.LoadStaticProps(fs, file)
	event.Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateStaticPropsLoaded))

	// Load entities
	ents, err := valve.LoadEntdata(fs, file)
	if err != nil {
		return nil, nil, err
	}
	event.Dispatch(messages.NewLoadingLevelProgress(messages.LoadingProgressStateEntitiesLoaded))

	return level, ents, err
}
