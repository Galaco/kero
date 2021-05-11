package studiomodel

import (
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/studiomodel"
	"github.com/galaco/studiomodel/mdl"
	"github.com/galaco/studiomodel/phy"
	"github.com/galaco/studiomodel/vtx"
	"github.com/galaco/studiomodel/vvd"
	"io"
	"strings"
)

type virtualFileSystem interface {
	GetFile(string) (io.Reader, error)
}

// @TODO This is SUPER incomplete
// right now it does the bare minimum, and many models seem to have
// some corruption.

// LoadProp loads a single prop/model of known filepath
func LoadProp(path string, fs virtualFileSystem) (*mesh.Model, error) {
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

func modelFromStudioModel(filename string, studioModel *studiomodel.StudioModel) (*mesh.Model, error) {
	verts, normals, textureCoordinates, indices, err := VertexDataForModel(studioModel, 0)
	if err != nil {
		return nil, err
	}
	outModel := mesh.NewModel(filename)
	mats := materialsForStudioModel(studioModel.Mdl)
	for i := 0; i < len(indices); i++ { //indices is a slice of slices, (ie len(indices) = num_meshes)
		smMesh := mesh.NewMesh()
		smMesh.AddVertex(verts...)
		smMesh.AddNormal(normals...)
		smMesh.AddUV(textureCoordinates...)
		smMesh.AddIndice(indices[i]...)

		//@TODO Map ALL materials to mesh data
		outModel.AddMaterial(mats[0])

		// @TODO Tangents already exist in props. Use those instead
		smMesh.GenerateTangents()
		outModel.AddMesh(smMesh)
	}

	return outModel, nil
}

func materialsForStudioModel(mdlData *mdl.Mdl) []string {
	materials := make([]string, 0)
	for _, dir := range mdlData.TextureDirs {
		//trueDir := strings.Replace(dir, "\\", "/", -1)
		for _, name := range mdlData.TextureNames {
			// In some cases the texture name seems to include the directory itself. e.g. csgo de_dust2
			//name = strings.TrimSpace(strings.TrimLeft(strings.Replace(name, "\\", "/", -1), trueDir))
			// materials = append(materials, trueDir + name)
			materials = append(materials, strings.Replace(dir, "\\", "/", -1)+name)
		}
	}
	return materials
}
