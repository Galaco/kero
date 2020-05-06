package filesystem

import (
	"fmt"
	"github.com/galaco/KeyValues"
	"github.com/galaco/bsp/lumps"
	filesystemLib "github.com/golang-source-engine/filesystem"
	"io"
	"log"
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

func InitFilesystem(gameDir string) FileSystem {
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
	fs, err := InitializeFromGameInfoDefinitions(gameDir, &gameInfo)
	if err != nil {
		if fsErr, ok := err.(*filesystemLib.InvalidResourcePathCollectionError); ok {
			for _, s := range strings.Split(fsErr.Error(), "|") {
				log.Println(fmt.Sprintf("Invalid resource path: %s", s))
			}
		}
	}

	return fs
}