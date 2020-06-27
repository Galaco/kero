package filesystem

import (
	"fmt"
	"github.com/galaco/KeyValues"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/kero/framework/console"
	filesystemLib "github.com/golang-source-engine/filesystem"
	"io"
	"os"
	"strings"
)

// FileSystem provides a gateway to interacting with the
// filesystem structures of Source Engine.
type FileSystem interface {
	// GetFile searches for a file path
	GetFile(string) (io.Reader, error)
	// RegisterPakFile adds a bsp pakfile to the filesystem search paths
	RegisterPakFile(pakFile *lumps.Pakfile)
}

var masterFilesystem FileSystem

// InitializeFromGameInfoDefinitions Reads game resource data paths
// from gameinfo.txt
// All games should ship with a gameinfo.txt, but it isn't actually mandatory.
func InitializeFromGameInfoDefinitions(basePath string, gameInfo *keyvalues.KeyValue) (FileSystem, error) {
	lfs, err := filesystemLib.CreateFilesystemFromGameInfoDefinitions(basePath, gameInfo, true)
	if lfs != nil {
		return lfs, err
	}
	return nil, err
}

// Init initializses the master filesystem used by Kero. In theory other filesystems can be used too; but the master fs
// is designed to be loaded with the same configuration and behaviour as the original Source Engine.
func Init(gameDir string) (FileSystem, error) {
	stream, err := os.Open(gameDir + "/gameinfo.txt")
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	kvReader := keyvalues.NewReader(stream)

	gameInfo, err := kvReader.Read()
	if err != nil {
		return nil, err
	}
	fs, err := InitializeFromGameInfoDefinitions(gameDir, &gameInfo)
	if err != nil {
		if fsErr, ok := err.(*filesystemLib.InvalidResourcePathCollectionError); ok {
			for _, s := range strings.Split(fsErr.Error(), "|") {
				console.PrintString(console.LevelError, fmt.Sprintf("Invalid resource path: %s", s))
			}
		}
	}

	// The reasonable assumption is there will only be 1 filesystem; the first initialized is considered the master fs,
	// and can be accessed via Get().
	if masterFilesystem == nil {
		masterFilesystem = fs
	}

	return fs, nil
}

// Get returns the master filesystem singleton
func Get() FileSystem {
	return masterFilesystem
}
