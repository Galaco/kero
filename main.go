package main

import (
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/lib/gameinfo"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/internal/config"
	"github.com/galaco/kero/systems/console"
	"github.com/galaco/kero/systems/gui"
	input2 "github.com/galaco/kero/systems/input"
	"github.com/galaco/kero/systems/renderer"
	"github.com/galaco/kero/systems/scene"
	"runtime"
)

func main() {
	runtime.LockOSThread()

	cfg, err := config.Load("./config.json")
	if err != nil {
		panic(err)
	}
	if err = initializeSourceEngine(cfg.GameDirectory); err != nil {
		panic(err)
	}
	if err = initializeFramework(); err != nil {
		panic(err)
	}

	engine := NewEngine(console.NewConsole(), input2.NewInput(), scene.NewScene(), renderer.NewRenderer(), gui.NewGui())
	engine.RunGameLoop()
}

func initializeSourceEngine(gameDir string) error {
	gameInfo, err := gameinfo.LoadConfig(gameDir)
	if err != nil {
		return err
	}
	filesystem.InitializeFromGameInfoDefinitions(gameDir, gameInfo)

	return nil
}

func initializeFramework() error {
	win, err := window.CreateWindow(config.Singleton().Video.Width, config.Singleton().Video.Height, "Lambda2")
	if err != nil {
		return err
	}
	win.SetActive()
	input.SetBoundWindow(win)
	return graphics.Init()
}
