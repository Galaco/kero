package valve

import (
	"fmt"
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/valve/studiomodel"
	"strings"
	"sync"
)

func LoadStaticProps(fs graphics.VirtualFileSystem, file *bsp.Bsp) []graphics.StaticProp {
	gameLump := file.Lump(bsp.LumpGame).(*lumps.Game)
	propLump := gameLump.GetStaticPropLump()

	// Get StaticProp list to load
	propPaths := make([]string, 0)
	for _, propEntry := range propLump.PropLumps {
		propPaths = append(propPaths, propLump.DictLump.Name[propEntry.GetPropType()])
	}
	propPaths = generateUniquePropList(propPaths)
	event.Dispatch(messages.NewConsoleMessage(console.LevelInfo, fmt.Sprintf("%d staticprops referenced", len(propPaths))))

	// Load Prop data
	propList := asyncLoadProps(fs, propPaths)
	event.Dispatch(messages.NewConsoleMessage(console.LevelInfo, fmt.Sprintf("%d staticprops loaded", len(propList))))

	//Transform to props to
	staticPropList := make([]graphics.StaticProp, 0)

	for _, propEntry := range propLump.PropLumps {
		modelName := propLump.DictLump.Name[propEntry.GetPropType()]
		if m, ok := propList[modelName]; ok {
			staticPropList = append(staticPropList, *graphics.NewStaticProp(propEntry, &propLump.LeafLump, m))
			continue
		} else {
			// error Prop
		}
	}

	return staticPropList
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

func asyncLoadProps(fs graphics.VirtualFileSystem, propPaths []string) map[string]*graphics.Model {
	propMap := map[string]*graphics.Model{}
	var propMapMutex sync.Mutex
	waitGroup := sync.WaitGroup{}

	asyncLoadProps := func(path string) {
		if !strings.HasSuffix(path, ".mdl") {
			path += ".mdl"
		}
		prop, err := studiomodel.LoadProp(path, fs)
		if err != nil {
			waitGroup.Done()
			return
		}
		propMapMutex.Lock()
		propMap[path] = prop
		propMapMutex.Unlock()
		waitGroup.Done()
	}

	for _, path := range propPaths {
		waitGroup.Add(1)
		go asyncLoadProps(path)
	}
	waitGroup.Wait()

	return propMap
}
