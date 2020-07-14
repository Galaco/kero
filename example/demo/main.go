package main

import (
	"flag"
	"fmt"
	"github.com/galaco/kero"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/window"
	"log"
	"runtime"
)

func main() {
	gameDirectoryPtr := flag.String("game", "", "Path to the root game directory")
	flag.Parse()
	if *gameDirectoryPtr == "" {
		panic("No game directory specified. Please run with the flag -game=\"<gameDir>\"")
	}

	runtime.LockOSThread()
	defer func() {
		if e := recover(); e != nil {
			panic(e)
		}
	}()

	console.AddOutputPipe(func(level console.LogLevel, data interface{}) {
		log.Println(data)
	})

	// Framework
	if err := initFramework(); err != nil {
		panic(err)
	}

	// Game config
	game := NewGameDefinition()
	_, err := filesystem.Init(*gameDirectoryPtr + "/" + game.ContentDirectory())
	if err != nil {
		panic(err)
	}
	window.CurrentWindow().Handle().Handle().SetTitle(fmt.Sprintf("Kero: %s", game.ContentDirectory()))

	// Start
	keroImpl := kero.NewKero()
	keroImpl.RegisterGameDefinitions(game)
	keroImpl.Start()
}

func initFramework() error {
	win, err := window.CreateWindow(1920, 1080, "Kero: A Source Engine Implementation")
	if err != nil {
		return err
	}

	win.SetActive()
	input.SetBoundWindow(win)
	win.Handle().Handle().Focus()
	return adapter.Init()
}
