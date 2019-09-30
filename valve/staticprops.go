package valve

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/valve/studiomodel"
	filesystemLib "github.com/golang-source-engine/filesystem"
	"strings"
)

func LoadStaticProps(fs filesystem.FileSystem, file *bsp.Bsp) {
	gameLump := file.Lump(bsp.LumpGame).(*lumps.Game)
	propLump := gameLump.GetStaticPropLump()

	// Get StaticProp list to load
	propPaths := make([]string, 0)
	for _, propEntry := range propLump.PropLumps {
		propPaths = append(propPaths, propLump.DictLump.Name[propEntry.GetPropType()])
	}
	propPaths = generateUniquePropList(propPaths)

	// Load Prop data
	_ = loadPropsFromFilesystem(fs, propPaths)

	// Transform to props to
	//staticPropList := make([]model.StaticProp, 0)

	//for _, propEntry := range propLump.PropLumps {
	//	modelName := propLump.DictLump.Name[propEntry.GetPropType()]
	//	m := ResourceManager.Model(modelName)
	//	if m != nil {
	//		staticPropList = append(staticPropList, *model.NewStaticProp(propEntry, &propLump.LeafLump, m))
	//		continue
	//	}
	//	// Model missing, use error model
	//	m = ResourceManager.Model(ResourceManager.ErrorModelName())
	//	staticPropList = append(staticPropList, *model.NewStaticProp(propEntry, &propLump.LeafLump, m))
	//}

	//return staticPropList
}

func generateUniquePropList(propList []string) (uniqueList []string) {
	list := map[string]bool{}
	for _, entry := range propList {
		list[entry] = true
	}
	for k := range list {
		uniqueList = append(uniqueList, k)
	}

	return uniqueList
}

func loadPropsFromFilesystem(fs filesystem.FileSystem, propPaths []string) map[string]*graphics.Model {
	propMap := map[string]*graphics.Model{}
	for _, path := range propPaths {
		if !strings.HasSuffix(path, ".mdl") {
			path += ".mdl"
		}
		prop, err := studiomodel.LoadProp(path, fs.(*filesystemLib.FileSystem))
		if err != nil {
			continue
		}
		propMap[path] = prop
	}

	return propMap
}
