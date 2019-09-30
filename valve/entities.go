package valve

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/valve/entity"
	entityLib "github.com/galaco/source-tools-common/entity"
	"github.com/galaco/vmf"
	"strings"
)

func LoadEntdata(fs filesystem.FileSystem, file *bsp.Bsp) ([]entity.Entity, error) {
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

func fromVmfNodeTree(entityNodes vmf.Node) []entity.Entity {
	numEntities := len(*entityNodes.GetAllValues())

	entities := make([]entity.Entity, numEntities)
	entitiesList := entityLib.FromVmfNodeTree(entityNodes)

	for i := 0; i < entitiesList.Length(); i++ {
		entities[i] = entity.NewEntityBaseFromLib(*entitiesList.Get(i))
	}

	return entities
}
