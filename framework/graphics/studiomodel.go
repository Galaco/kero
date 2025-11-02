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

const (
	stripIsTriangleList = 0x01
)

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
	verts, normals, textureCoordinates, err := vertexDataForMesh(studioModel.Vvd)
	if err != nil {
		return nil, err
	}

	outModel := mesh.NewModel(filename, studioModel)
	mats := materialsForStudioModel(studioModel.Mdl)

	// Iterate through VTX body parts (matches MDL hierarchy 1:1)
	for bodyPartIdx, bodyPart := range studioModel.Vtx.BodyParts {
		for modelIdx, model := range bodyPart.Models {
			for meshIdx, vtxMesh := range model.LODS[0].Meshes {
				// Get corresponding MDL data for correct vertex offset calculation
				var mdlMesh *mdl.Mesh
				var mdlModel *mdl.Model

				if bodyPartIdx < len(studioModel.Mdl.BodyParts) &&
					modelIdx < len(studioModel.Mdl.BodyParts[bodyPartIdx].Models) &&
					meshIdx < len(studioModel.Mdl.BodyParts[bodyPartIdx].Models[modelIdx].Meshes) {

					mdlMesh = &studioModel.Mdl.BodyParts[bodyPartIdx].Models[modelIdx].Meshes[meshIdx]
					mdlModel = &studioModel.Mdl.BodyParts[bodyPartIdx].Models[modelIdx].Header
				} else {
					// Fallback: create default mesh/model with zero offsets
					mdlMesh = &mdl.Mesh{}
					mdlModel = &mdl.Model{}
				}

				// Extract indices for this mesh using correct vertex offsets
				indices := indicesForMesh(&vtxMesh, mdlMesh, mdlModel)
				if len(indices) == 0 {
					continue
				}

				// Get material index from MDL for this mesh
				materialIdx := int32(0) // Default to first material
				if len(studioModel.Mdl.BodyParts) > 0 {
					if matIdx, err := studioModel.Mdl.GetMaterialIndexForMesh(bodyPartIdx, modelIdx, meshIdx); err == nil {
						materialIdx = matIdx
					}
				}

				// Validate material index and fallback if out of bounds
				if int(materialIdx) >= len(mats) || materialIdx < 0 {
					materialIdx = 0 // Fallback to first material
				}

				// Create mesh with correct material
				smMesh := mesh.NewMesh()
				smMesh.AddVertex(verts...)
				smMesh.AddNormal(normals...)
				smMesh.AddUV(textureCoordinates...)
				smMesh.AddIndice(indices...)

				// @TODO Tangents already exist in props. Use those instead
				smMesh.GenerateTangents()

				outModel.AddMesh(smMesh)
				outModel.AddMaterial(mats[materialIdx]) // Use correct material!
			}
		}
	}

	// Compute bounding box now that all meshes are added
	outModel.ComputeBounds()

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

	for bodyPartIdx, bodyPart := range studioModel.Vtx.BodyParts {
		for modelIdx, model := range bodyPart.Models {
			if lodIdx >= len(model.LODS) {
				return nil, nil, nil, nil, errors.New("invalid LOD index requested for model")
			}
			for meshIdx, m := range model.LODS[lodIdx].Meshes {
				// Get corresponding MDL data for correct vertex offset calculation
				var mdlMesh *mdl.Mesh
				var mdlModel *mdl.Model

				if bodyPartIdx < len(studioModel.Mdl.BodyParts) &&
					modelIdx < len(studioModel.Mdl.BodyParts[bodyPartIdx].Models) &&
					meshIdx < len(studioModel.Mdl.BodyParts[bodyPartIdx].Models[modelIdx].Meshes) {

					mdlMesh = &studioModel.Mdl.BodyParts[bodyPartIdx].Models[modelIdx].Meshes[meshIdx]
					mdlModel = &studioModel.Mdl.BodyParts[bodyPartIdx].Models[modelIdx].Header
				} else {
					// Fallback: create default mesh/model with zero offsets
					mdlMesh = &mdl.Mesh{}
					mdlModel = &mdl.Model{}
				}

				i := indicesForMesh(&m, mdlMesh, mdlModel)
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

// convertTriangleStripToList converts triangle strip indices to triangle list indices
// Triangle strips share vertices between adjacent triangles and require alternating winding order
func convertTriangleStripToList(stripIndices []uint32) []uint32 {
	if len(stripIndices) < 3 {
		return stripIndices
	}

	numTriangles := len(stripIndices) - 2
	triangleIndices := make([]uint32, 0, numTriangles*3)

	for i := 0; i < numTriangles; i++ {
		if i%2 == 0 {
			// Even triangle: normal order (i, i+1, i+2)
			triangleIndices = append(triangleIndices,
				stripIndices[i],
				stripIndices[i+1],
				stripIndices[i+2])
		} else {
			// Odd triangle: reversed winding (i+2, i+1, i)
			triangleIndices = append(triangleIndices,
				stripIndices[i+2],
				stripIndices[i+1],
				stripIndices[i])
		}
	}

	return triangleIndices
}

// indicesForMesh get indices for mesh
// Processes ALL stripgroups within the mesh to extract complete geometry
// Uses correct vertex offset calculation from MDL mesh and model data
func indicesForMesh(mesh *vtx.Mesh, mdlMesh *mdl.Mesh, mdlModel *mdl.Model) []uint32 {
	meshIndices := make([]uint32, 0)

	// Vertex struct size in VVD file (mstudio_vertex_t)
	const vertexStructSize = 48

	// Process ALL stripgroups (not just the first one)
	// Complex models have multiple stripgroups for hardware skinning, flexed geometry, etc.
	for _, stripGroup := range mesh.StripGroups {
		for _, strip := range stripGroup.Strips {
			// Process BOTH triangle lists (0x01) AND triangle strips (0x02)
			isTriangleList := strip.Flags&0x01 != 0
			isTriangleStrip := strip.Flags&0x02 != 0

			if !isTriangleList && !isTriangleStrip {
				continue // Skip unknown strip types
			}

			// Extract raw indices from this strip
			stripIndices := make([]uint32, strip.NumIndices)
			for i := int32(0); i < strip.NumIndices; i++ {
				// Multi-level indirection chain (see NOTES.md):
				// 1. index = strip.IndexOffset + i
				index := strip.IndexOffset + i

				// 2. index2 = stripGroup.Indices[index]
				index2 := stripGroup.Indices[index]

				// 3. index3 = stripGroup.Vertexes[index2].OriginalMeshVertexID
				vert := stripGroup.Vertexes[index2]
				index3 := vert.OriginalMeshVertexID

				// 4. index4 = mesh.VertexOffset + index3
				index4 := mdlMesh.VertexOffset + int32(index3)

				// 5. index5 = index4 + (model.VertexIndex / sizeof(vertex))
				finalIndex := uint32(index4 + (mdlModel.VertexIndex / vertexStructSize))

				stripIndices[i] = finalIndex
			}

			// Convert triangle strips to triangle lists if needed
			if isTriangleStrip {
				stripIndices = convertTriangleStripToList(stripIndices)
			}

			meshIndices = append(meshIndices, stripIndices...)
		}
	}

	return meshIndices
}

func vertexDataForMesh(vvd *vvd.Vvd) ([]float32, []float32, []float32, error) {
	vertices := make([]float32, 0, len(vvd.Vertices)*3)
	normals := make([]float32, 0, len(vvd.Vertices)*3)
	uvs := make([]float32, 0, len(vvd.Vertices)*2)

	for _, vertex := range vvd.Vertices {
		vertices = append(vertices, vertex.Position[0], vertex.Position[1], vertex.Position[2])
		normals = append(normals, vertex.Normal[0], vertex.Normal[1], vertex.Normal[2])
		uvs = append(uvs, vertex.UVs[0], vertex.UVs[1])
	}

	return vertices, normals, uvs, nil
}
