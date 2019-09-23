package filesystem

import (
	"github.com/galaco/KeyValues"
	filesystemLib "github.com/golang-source-engine/filesystem"
)

var fs *filesystemLib.FileSystem

// Singleton
func Singleton() *filesystemLib.FileSystem {
	return fs
}

// InitializeFromGameInfoDefinitions Reads game resource data paths
// from gameinfo.txt
// All games should ship with a gameinfo.txt, but it isn't actually mandatory.
func InitializeFromGameInfoDefinitions(basePath string, gameInfo *keyvalues.KeyValue) (*filesystemLib.FileSystem, error) {
	lfs, err := filesystemLib.CreateFilesystemFromGameInfoDefinitions(basePath, gameInfo, true)
	fs = lfs
	return fs, err
}
