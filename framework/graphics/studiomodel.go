package graphics

import (
	"errors"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/studiomodel"
	"github.com/galaco/studiomodel/mdl"
	"github.com/galaco/studiomodel/phy"
	"github.com/galaco/studiomodel/vtx"
	"github.com/galaco/studiomodel/vvd"
	"io"
	"strings"
)

// @TODO This is SUPER incomplete
// right now it does the bare minimum, and many models seem to have
// some corruption.

type virtualFileSystem interface {
	GetFile(string) (io.Reader, error)
}

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
	outModel := mesh.NewModel(filename, studioModel)
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

// VertexDataForModel loads model vertex data
func VertexDataForModel(studioModel *studiomodel.StudioModel, lodIdx int) ([]float32, []float32, []float32, [][]uint32, error) {
	indices := make([][]uint32, 0)
	for _, bodyPart := range studioModel.Vtx.BodyParts {
		for _, model := range bodyPart.Models {
			if lodIdx > len(model.LODS) {
				return nil, nil, nil, nil, errors.New("invalid LOD index requested for model")
			}
			for _, mesh := range model.LODS[lodIdx].Meshes {
				i := indicesForMesh(&mesh)
				if len(i) == 0 {
					return nil, nil, nil, nil, errors.New("invalid studiomodel mesh: 0 indices. ignoring")
				}
				indices = append(indices, i)
			}
		}
	}

	vertices, normals, textureCoordinates, err := vertexDataForMesh(studioModel.Vvd)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return vertices, normals, textureCoordinates, indices, nil
}

// indicesForMesh get indices for mesh
func indicesForMesh(mesh *vtx.Mesh) []uint32 {
	if len(mesh.StripGroups) > 1 {
		return make([]uint32, 0)
	}
	meshIndices := make([]uint32, 0)

	// @TODO Use all strip groups
	stripGroup := mesh.StripGroups[0]

	for _, strip := range stripGroup.Strips {
		for j := int32(0); j < strip.NumIndices; j++ {
			index := stripGroup.Indices[strip.IndexOffset+j]
			vert := stripGroup.Vertexes[index]

			meshIndices = append(meshIndices, uint32(strip.VertOffset)+uint32(vert.OriginalMeshVertexID))
		}
	}

	return meshIndices
}

func vertexDataForMesh(vvd *vvd.Vvd) ([]float32, []float32, []float32, error) {
	vertices := make([]float32, 0, len(vvd.Vertices)*3)
	normals := make([]float32, 0, len(vvd.Vertices)*3)
	uvs := make([]float32, 0, len(vvd.Vertices)*2)

	for _, vertex := range vvd.Vertices {
		vertices = append(vertices, vertex.Position.X(), vertex.Position.Y(), vertex.Position.Z())
		normals = append(normals, vertex.Normal.X(), vertex.Normal.Y(), vertex.Normal.Z())
		uvs = append(uvs, vertex.UVs.X(), vertex.UVs.Y())
	}

	return vertices, normals, uvs, nil
}
