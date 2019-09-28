package filesystem

import (
	"github.com/galaco/KeyValues"
	filesystemLib "github.com/golang-source-engine/filesystem"
	"io"
)

type FileSystem interface {
	GetFile(string) (io.Reader, error)
}

// InitializeFromGameInfoDefinitions Reads game resource data paths
// from gameinfo.txt
// All games should ship with a gameinfo.txt, but it isn't actually mandatory.
func InitializeFromGameInfoDefinitions(basePath string, gameInfo *keyvalues.KeyValue) (FileSystem, error) {
	lfs, err := filesystemLib.CreateFilesystemFromGameInfoDefinitions(basePath, gameInfo, true)
	if lfs != nil {
		return lfs, err
	}
	return nil,err
}
