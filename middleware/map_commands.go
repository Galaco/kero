package middleware

import (
	"fmt"
	"strings"

	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/messages"
)

// AddMapCommands registers map-related console commands
func AddMapCommands(eventBus *event.Dispatcher) {
	console.AddCommand("map", "Load a map", "map <mapname>", func(options string) error {
		if options == "" {
			console.PrintString(console.LevelWarning, "Usage: map <mapname>")
			console.PrintString(console.LevelInfo, "Example: map de_dust2")
			return nil
		}

		mapName := strings.TrimSpace(options)
		// Strip .bsp extension if user provided it
		mapName = strings.TrimSuffix(mapName, ".bsp")

		// Construct relative path for filesystem lookup
		// The filesystem will search through all registered directories (maps/, download/maps/, etc.)
		relativePath := "maps/" + mapName + ".bsp"

		console.PrintString(console.LevelInfo, fmt.Sprintf("Loading map: %s", mapName))

		// Dispatch event with relative path
		// The BSP loader will use fs.GetFile() which searches through:
		// 1. BSP pakfile (if loaded)
		// 2. Local directories (from gameinfo.txt)
		// 3. VPK packages
		// If the map is not found, the loader will return an error that's handled by Scene.Update()
		event.DispatchTypedDeferred(eventBus, event.PhasePostUpdate, messages.ChangeLevelEvent{MapName: relativePath})

		return nil
	})
}
