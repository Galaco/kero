package main

import (
	"flag"
	"github.com/galaco/kero"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/systems"
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


	fs := filesystem.InitFilesystem(*gameDirectoryPtr + "/" + game.ContentDirectory())
	if err := initFramework(); err != nil {
		panic(err)
	}
	context := systems.Context{
		Client:     game.Client(),
		Filesystem: fs,
	}

	keroImpl := kero.NewKero(context)
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
	return graphics.Init()
}
