package entity

import (
	"io"
	"strings"

	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lump"
	entityLib "github.com/galaco/source-tools-common/entity"
	"github.com/galaco/vmf"
)

type filesystem interface {
	// GetFile searches for a file path
	GetFile(string) (io.Reader, error)
	// RegisterPakFile adds a bsp pakfile to the filesystem search paths
	RegisterPakFile(pakFile *lump.Pakfile)
}

// LoadEntdata extracts entity data from the bsp
func LoadEntdata(file *bsp.Bsp) ([]IEntity, error) {
	entdata, err := file.Lumps[bsp.LumpEntities].(*lump.EntData).ToBytes()
	if err != nil {
		return nil, err
	}

	vmfEntityTree, err := parseEntdata(entdata)
	if err != nil {
		return nil, err
	}
	entityList := fromVmfNodeTree(vmfEntityTree.Unclassified)
	//for i := 0; i < entityList.Length(); i++ {
	//	targetScene.AddEntity(entityLib.CreateEntity(entityList.Get(i), fs))
	//}
	return entityList, nil
}

func parseEntdata(data []byte) (vmf.Vmf, error) {
	stringReader := strings.NewReader(string(data))
	reader := vmf.NewReader(stringReader)

	return reader.Read()
}

func fromVmfNodeTree(entityNodes vmf.Node) []IEntity {
	numEntities := len(*entityNodes.GetAllValues())

	entities := make([]IEntity, numEntities)
	entitiesList := entityLib.FromVmfNodeTree(entityNodes)

	for i := 0; i < entitiesList.Length(); i++ {
		entities[i] = NewEntityBaseFromLib(*entitiesList.Get(i))
	}

	return entities
}
