package main

import (
	"fmt"
	keyvalues "github.com/galaco/KeyValues"
	kero2 "github.com/galaco/kero"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/input"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/systems"
	filesystemLib "github.com/golang-source-engine/filesystem"
	"log"
	"os"
	"runtime"
	"strings"
)

const (
	GameDirectory = "E:/Program Files/Steam/Steamapps/common/Counter-Strike Source"
)

func main() {
	runtime.LockOSThread()
	defer func() {
		if e := recover(); e != nil {
			panic(e)
		}
	}()

	game := NewGameDefinition()

	fs := initFilesystem(GameDirectory + "/" + game.ContentDirectory())
	if err := initFramework(); err != nil {
		panic(err)
	}
	context := systems.Context{
		Client:     game.Client(),
		Filesystem: fs,
	}

	keroImpl := kero2.NewKero(context)
	keroImpl.RegisterGameDefinitions(game)
	keroImpl.Start()
}

func initFilesystem(gameDir string) filesystem.FileSystem {
	stream, err := os.Open(gameDir + "/gameinfo.txt")
	if err != nil {
		panic(err)
	}
	defer stream.Close()
	kvReader := keyvalues.NewReader(stream)

	gameInfo, err := kvReader.Read()
	if err != nil {
		panic(err)
	}
	fs, err := filesystem.InitializeFromGameInfoDefinitions(gameDir, &gameInfo)
	if err != nil {
		if fsErr, ok := err.(*filesystemLib.InvalidResourcePathCollectionError); ok {
			for _, s := range strings.Split(fsErr.Error(), "|") {
				log.Println(fmt.Sprintf("Invalid resource path: %s", s))
			}
		}
	}

	return fs
}

func initFramework() error {
	win, err := window.CreateWindow(800, 600, "Kero")
	if err != nil {
		return err
	}
	win.SetActive()
	input.SetBoundWindow(win)
	return graphics.Init()
}
