package systems

import (
	"github.com/galaco/kero/config"
	"github.com/galaco/kero/framework/filesystem"
)

type Context struct {
	Config     *config.Config
	Filesystem filesystem.FileSystem
}
