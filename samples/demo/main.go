package main

import (
	"github.com/galaco/kero"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/systems"
	"runtime"
)

const (
	GameDirectory = "/Users/galaco/Library/Application Support/Steam/steamapps/common/Counter-Strike Source"
)

func main() {
	runtime.LockOSThread()
	defer func() {
		if e := recover(); e != nil {
			panic(e)
		}
	}()

	game := NewGameDefinition()

	fs := filesystem.InitFilesystem(GameDirectory + "/" + game.ContentDirectory())
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

func initFramework () error {
	win, err := window.CreateWindow(800, 600, "kero")
	if err != nil {
		return err
	}

	win.SetActive()
	input.SetBoundWindow(win)
	return graphics.Init()
}