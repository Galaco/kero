package studiomodel

import (
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/studiomodel"
	"github.com/galaco/studiomodel/mdl"
	"github.com/galaco/studiomodel/phy"
	"github.com/galaco/studiomodel/vtx"
	"github.com/galaco/studiomodel/vvd"
	"io"
	"log"
	"strings"
)

type virtualFileSystem interface {
	GetFile(string) (io.Reader, error)
}

// @TODO This is SUPER incomplete
// right now it does the bare minimum, and many models seem to have
// some corruption.

// LoadProp loads a single prop/model of known filepath
func LoadProp(path string, fs virtualFileSystem) (*graphics.Model, error) {
	prop, err := loadProp(strings.Split(path, ".mdl")[0], fs)
	if prop != nil {
		model, err := modelFromStudioModel(path, prop)
		if err != nil {
			return nil, err
		}
		return model, nil
	}
	return nil, err
}

func loadProp(filePath string, fs virtualFileSystem) (*studiomodel.StudioModel, error) {
	prop := studiomodel.NewStudioModel(filePath)

	// MDL
	f, err := fs.GetFile(filePath + ".mdl")
	if err != nil {
		return nil, err
	}
	mdlFile, err := mdl.ReadFromStream(f)
	if err != nil {
		return nil, err
	}
	prop.AddMdl(mdlFile)

	// VVD
	f, err = fs.GetFile(filePath + ".vvd")
	if err != nil {
		return nil, err
	}
	vvdFile, err := vvd.ReadFromStream(f)
	if err != nil {
		return nil, err
	}
	prop.AddVvd(vvdFile)

	// VTX
	f, err = fs.GetFile(filePath + ".dx90.vtx")
	if err != nil {
		return nil, err
	}
	vtxFile, err := vtx.ReadFromStream(f)

	if err != nil {
		return nil, err
	}
	prop.AddVtx(vtxFile)

	// PHY
	f, err = fs.GetFile(filePath + ".phy")
	if err != nil {
		return prop, err
	}

	phyFile, err := phy.ReadFromStream(f)
	if err != nil {
		return prop, err
	}
	prop.AddPhy(phyFile)

	return prop, nil
}

func modelFromStudioModel(filename string, studioModel *studiomodel.StudioModel) (*graphics.Model, error) {
	if filename == "models/props/de_tides/tides_fences_d.mdl" {
		log.Println(filename)
	}
	verts, normals, textureCoordinates, err := VertexDataForModel(studioModel, 0)
	if err != nil {
		return nil, err
	}
	outModel := graphics.NewModel(filename)
	mats := materialsForStudioModel(studioModel.Mdl)
	for i := 0; i < len(verts); i++ { //verts is a slice of slices, (ie vertex data per mesh)
		smMesh := graphics.NewMesh()
		smMesh.AddVertex(verts[i]...)
		smMesh.AddNormal(normals[i]...)
		smMesh.AddUV(textureCoordinates[i]...)

		//@TODO Map ALL materials to mesh data
		outModel.AddMaterial(mats[0])

		smMesh.GenerateTangents()
		outModel.AddMesh(smMesh)
	}

	return outModel, nil
}

func materialsForStudioModel(mdlData *mdl.Mdl) []string {
	materials := make([]string, 0)
	for _, dir := range mdlData.TextureDirs {
		for _, name := range mdlData.TextureNames {
			materials = append(materials, strings.Replace(dir, "\\", "/", -1)+name)
		}
	}
	return materials
}
