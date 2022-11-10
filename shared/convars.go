package shared

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/galaco/kero/internal/framework/console"
)

func BindSharedConsoleCommands() {
	console.AddCommand("set", "Sets a convar", "set <convar> <value int|string|bool>", func(options string) error {
		if options == "" || strings.Index(options, " ") < 1 {
			return nil
		}

		parts := strings.Split(options, " ")
		if parts[1] == "true" {
			console.SetConvarBoolean(parts[0], true)
			return nil
		}
		if parts[1] == "false" {
			console.SetConvarBoolean(parts[0], false)
			return nil
		}
		i, err := strconv.Atoi(parts[1])
		if err == nil {
			console.SetConvarInt(parts[0], i)
			return nil
		}

		console.SetConvarString(parts[0], parts[1])
		return nil
	})
	console.AddCommand("get", "Returns the current value of a convar", "get <convar>", func(options string) error {
		if options == "" {
			return nil
		}

		parts := strings.Split(options, " ")
		cv := console.GetConvar(parts[0])
		if cv != nil {
			console.PrintInterface(console.LevelInfo, fmt.Sprintf("%s: %s", parts[0], cv.Value))
		}

		return nil
	})
}
