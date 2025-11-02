package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/galaco/kero/framework/console"
)

func AddInitialConvars() {
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

	console.AddConvarBool("developer", "Enable developer mode (more verbose logging)", false)
	console.AddConvarBool("hdr_enable", "Use HDR by default", true)

	// Performance metrics ConVars
	console.AddConvarBool("r_showperf", "Enable performance metrics collection and display", true)
	console.AddConvarInt("r_perfhistory", "Number of performance samples to keep in history", 300)
	console.AddConvarInt("r_perfgraphheight", "Performance graph height in pixels", 200)
	console.AddConvarInt("r_perfgraphwidth", "Performance graph width in pixels", 600)

	// Camera movement ConVars
	console.AddConvarFloat("cam_acceleration", "Camera acceleration rate (units/second^2)", 1024.0)
	console.AddConvarFloat("cam_deceleration", "Camera deceleration multiplier", 512.0)
	console.AddConvarFloat("cam_maxspeed", "Maximum camera movement speed (units/second)", 320.0)

	// Mouse sensitivity ConVar
	console.AddConvarFloat("m_sensitivity", "Mouse sensitivity multiplier", 1.0)
}
