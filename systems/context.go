package systems

import (
	"github.com/galaco/kero/config"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/game"
)

type Context struct {
	Client     game.Client
	Config     *config.Config
	Filesystem filesystem.FileSystem
}
