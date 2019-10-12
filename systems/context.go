package systems

import (
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/game"
)

type Context struct {
	Client     game.Client
	Filesystem filesystem.FileSystem
}
