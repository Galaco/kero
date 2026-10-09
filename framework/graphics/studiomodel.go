package graphics

import (
	"fmt"
	"io"
	"strings"

	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/studiomodel"
	"github.com/galaco/studiomodel/mdl"
	"github.com/galaco/studiomodel/phy"
	"github.com/galaco/studiomodel/vtx"
	"github.com/galaco/studiomodel/vvd"
)

type virtualFileSystem interface {
	GetFile(string) (io.Reader, error)
}

// LoadProp loads a single prop/model of known filepath
func LoadProp(path string, fs virtualFileSystem) (*mesh.Model, error) {
	prop, err := loadProp(strings.Split(path, ".mdl")[0], fs)
	if prop != nil {
		model, err := modelFromStudioModel(path, prop, fs)
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
	// The vtx strip header layout depends on the mdl version
	vtxFile, err := vtx.ReadFromStreamWithMDLVersion(f, mdlFile.Header.Version)
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

func modelFromStudioModel(filename string, studioModel *studiomodel.StudioModel, fs virtualFileSystem) (*mesh.Model, error) {
	// BuildLOD applies the vvd's fixups. Without them, models with LODs that have fixups (e.g. de_dust2's palm
	// leaves) index the wrong vertices.
	lod, err := studioModel.BuildLOD(0, 0)
	if err != nil {
		return nil, err
	}

	outModel := mesh.NewModel(filename, studioModel)
	modelMesh := outModel.Mesh()
	for _, vertex := range lod.Vertices {
		modelMesh.AddVertex(vertex.Position[0], vertex.Position[1], vertex.Position[2])
		modelMesh.AddNormal(vertex.Normal[0], vertex.Normal[1], vertex.Normal[2])
		modelMesh.AddUV(vertex.UVs[0], vertex.UVs[1])
	}
	for _, tangent := range lod.Tangents {
		modelMesh.AddTangent(tangent[0], tangent[1], tangent[2], tangent[3])
	}

	// Each skin draws the same triangles with its own materials. A skin that cannot be built uses skin 0's.
	skinLODs := []*studiomodel.LOD{lod}
	for skin := 1; skin < len(studioModel.Mdl.SkinFamilies); skin++ {
		skinLOD, err := studioModel.BuildLOD(0, skin)
		if err != nil {
			skinLOD = lod
		}
		skinLODs = append(skinLODs, skinLOD)
	}

	for meshIdx, lodMesh := range lod.Meshes {
		// Every model of a body part is a bodygroup choice; the default body draws the first
		if lodMesh.Model != 0 || len(lodMesh.Indices) == 0 {
			continue
		}
		materials := make([]string, len(skinLODs))
		for skin, skinLOD := range skinLODs {
			materials[skin] = materialPathForStudioModel(studioModel.Mdl, skinLOD.Meshes[meshIdx].Material, fs)
		}
		outModel.AddSubMesh(lodMesh.Indices, materials)
	}

	outModel.ComputeBounds()

	return outModel, nil
}

// materialPathForStudioModel finds which of the mdl's material directories contains a material, in the order the
// mdl lists them, as the engine does. If none do, the first directory is used.
func materialPathForStudioModel(mdlData *mdl.Mdl, name string, fs virtualFileSystem) string {
	name = strings.ReplaceAll(name, "\\", "/")
	if len(mdlData.TextureDirs) == 0 {
		return name
	}
	if len(mdlData.TextureDirs) > 1 {
		for _, dir := range mdlData.TextureDirs {
			path := strings.ReplaceAll(dir, "\\", "/") + name
			if _, err := fs.GetFile(fmt.Sprintf("materials/%s.vmt", path)); err == nil {
				return path
			}
		}
	}
	return strings.ReplaceAll(mdlData.TextureDirs[0], "\\", "/") + name
}
