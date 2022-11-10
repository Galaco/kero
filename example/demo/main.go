package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/galaco/kero"
	"github.com/galaco/kero/internal/framework/console"
	"github.com/galaco/kero/internal/framework/debug"
	"github.com/galaco/kero/internal/framework/filesystem"
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
	debug.StartProfiler()

	// Game config
	game := NewGameDefinition()
	_, err := filesystem.Init(*gameDirectoryPtr)
	if err != nil {
		panic(err)
	}

	// Start
	keroImpl := kero.NewKero()
	keroImpl.RegisterGameDefinitions(game)
	keroImpl.Start()
}
