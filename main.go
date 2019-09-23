package main

import (
	"fmt"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/lib/gameinfo"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/internal/config"
	filesystemLib "github.com/golang-source-engine/filesystem"
	"log"
	"runtime"
	"strings"
)

func main() {
	runtime.LockOSThread()

	initialiseFramework()
	engine := NewKero()
	engine.Start()
}

func initialiseFramework() {
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
}

func initializeSourceEngine(gameDir string) error {
	gameInfo, err := gameinfo.LoadConfig(gameDir)
	if err != nil {
		return err
	}
	_, err = filesystem.InitializeFromGameInfoDefinitions(gameDir, gameInfo)
	if err != nil {
		if fsErr, ok := err.(*filesystemLib.InvalidResourcePathCollectionError); ok {
			for _, s := range strings.Split(fsErr.Error(), "|") {
				log.Println(fmt.Sprintf("Invalid resource path: %s", s))
			}
		}
	}

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
