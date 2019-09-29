package studiomodel

import (
	"github.com/galaco/StudioModel"
	"github.com/galaco/StudioModel/mdl"
	"github.com/galaco/StudioModel/phy"
	"github.com/galaco/StudioModel/vtx"
	"github.com/galaco/StudioModel/vvd"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"strings"
)

// @TODO This is SUPER incomplete
// right now it does the bare minimum, and many models seem to have
// some corruption.

// LoadProp loads a single prop/model of known filepath
func LoadProp(path string, fs filesystem.FileSystem) (*graphics.Model, error) {
	prop, err := loadProp(strings.Split(path, ".mdl")[0], fs)
	if prop != nil {
		_,err := modelFromStudioModel(path, prop, fs)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return nil,err
}

func loadProp(filePath string, fs filesystem.FileSystem) (*studiomodel.StudioModel, error) {
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

func modelFromStudioModel(filename string, studioModel *studiomodel.StudioModel, fs filesystem.FileSystem) (*graphics.Model, error) {
	verts, normals, textureCoordinates, err := VertexDataForModel(studioModel, 0)
	if err != nil {
		return nil, err
	}
	outModel := graphics.NewModel(filename)
	mats := materialsForStudioModel(studioModel.Mdl, fs)
	for i := 0; i < len(verts); i++ { //verts is a slice of slices, (ie vertex data per mesh)
		smMesh := graphics.NewMesh()
		smMesh.AddVertex(verts[i]...)
		smMesh.AddNormal(normals[i]...)
		smMesh.AddUV(textureCoordinates[i]...)

		//@TODO Map ALL materials to mesh data
		outModel.AddMaterial(mats[0])

		outModel.AddMesh(smMesh)
	}

	return outModel, nil
}

func materialsForStudioModel(mdlData *mdl.Mdl, fs filesystem.FileSystem) []string {
	materials := make([]string, 0)
	for _, dir := range mdlData.TextureDirs {
		for _, name := range mdlData.TextureNames {
			materials = append(materials, strings.Replace(dir, "\\", "/", -1) + name)
		}
	}
	return materials
}

