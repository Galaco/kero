package loader

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/utils/valve"
	"github.com/go-gl/mathgl/mgl32"
)

// LoadBspMap is the gateway into loading the core static level. Entities are loaded
// elsewhere
// It loads in the following order:
// BSP Geometry
// BSP Materials
// StaticProps (materials loaded as required)
func LoadBspMap(fs filesystem.FileSystem, filename string) (*valve.Bsp, []entity.IEntity, error) {
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateStarted))
	file, err := bsp.ReadFromFile(filename)
	if err != nil {
		event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateError))
		return nil, nil, err
	}
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateBSPParsed))
	fs.RegisterPakFile(file.Lump(bsp.LumpPakfile).(*lumps.Pakfile))
	// Load the static bsp world
	level, err := valve.LoadBSPWorld(fs, file)

	if err != nil {
		event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateError))
		return nil, nil, err
	}
	level.SetCamera(graphics3d.NewCamera(
		mgl32.DegToRad(70),
		float32(window.CurrentWindow().Width())/float32(window.CurrentWindow().Height())))
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateGeometryLoaded))

	// Load staticprops
	level.StaticPropDictionary, level.StaticProps = valve.LoadStaticProps(fs, file)
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateStaticPropsLoaded))

	// Load entities
	ents, err := entity.LoadEntdata(fs, file)
	if err != nil {
		return nil, nil, err
	}
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateEntitiesLoaded))

	return level, ents, err
}
