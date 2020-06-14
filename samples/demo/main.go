package main

import (
	"flag"
	"github.com/galaco/kero"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/window"
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

	game := NewGameDefinition()


	_, err := filesystem.Init(*gameDirectoryPtr + "/" + game.ContentDirectory())
	if err != nil {
		panic(err)
	}
	if err := initFramework(); err != nil {
		panic(err)
	}

	keroImpl := kero.NewKero()
	keroImpl.RegisterGameDefinitions(game)
	keroImpl.Start()
}

func initFramework() error {
	win, err := window.CreateWindow(1920, 1080, "kero")
	if err != nil {
		return err
	}

	win.SetActive()
	input.SetBoundWindow(win)
	win.Handle().Handle().Focus()
	return graphics.Init()
}
