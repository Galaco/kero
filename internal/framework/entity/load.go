package entity

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	entityLib "github.com/galaco/source-tools-common/entity"
	"github.com/galaco/vmf"
	"io"
	"strings"
)

type filesystem interface {
	// GetFile searches for a file path
	GetFile(string) (io.Reader, error)
	// RegisterPakFile adds a bsp pakfile to the filesystem search paths
	RegisterPakFile(pakFile *lumps.Pakfile)
}

// LoadEntdata extracts entity data from the bsp
func LoadEntdata(fs filesystem, file *bsp.Bsp) ([]IEntity, error) {
	entdata := file.Lump(bsp.LumpEntities).(*lumps.EntData)
	vmfEntityTree, err := parseEntdata(entdata.GetData())
	if err != nil {
		return nil, err
	}
	entityList := fromVmfNodeTree(vmfEntityTree.Unclassified)
	//for i := 0; i < entityList.Length(); i++ {
	//	targetScene.AddEntity(entityLib.CreateEntity(entityList.Get(i), fs))
	//}
	return entityList, nil
}

func parseEntdata(data string) (vmf.Vmf, error) {
	stringReader := strings.NewReader(data)
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
