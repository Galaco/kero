package studiomodel

import (
	"errors"
	"github.com/galaco/studiomodel"
	"github.com/galaco/studiomodel/vtx"
	"github.com/galaco/studiomodel/vvd"
)

// VertexDataForModel loads model vertex data
func VertexDataForModel(studioModel *studiomodel.StudioModel, lodIdx int) ([][]float32, [][]float32, [][]float32, [][]uint32, error) {
	vertices := make([][]float32, 0)
	normals := make([][]float32, 0)
	textureCoordinates := make([][]float32, 0)
	indices := make([][]uint32, 0)
	for _, bodyPart := range studioModel.Vtx.BodyParts {
		for _, model := range bodyPart.Models {
			if len(model.LODS) < lodIdx {
				return nil, nil, nil, nil, errors.New("invalid LOD index requested for model")
			}
			for _, mesh := range model.LODS[lodIdx].Meshes {
				rawIndices := indicesForMesh(&mesh)
				if len(rawIndices) == 0 {
					continue
				}

				v, n, uv, i, err := vertexDataForMesh(rawIndices, studioModel.Vvd)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				vertices = append(vertices, v)
				normals = append(normals, n)
				textureCoordinates = append(textureCoordinates, uv)
				indices = append(indices, i)
			}
		}
	}

	return vertices, normals, textureCoordinates, indices, nil
}

// indicesForMesh get indices for mesh
func indicesForMesh(mesh *vtx.Mesh) []uint16 {
	if len(mesh.StripGroups) > 1 {
		return make([]uint16, 0)
	}
	//	indexMap := make([]uint16, 0)
	meshIndices := make([]uint16, 0)

	stripGroup := mesh.StripGroups[0]

	//for i := 0; i < len(stripGroup.Vertexes); i++ {
	//	indexMap = append(indexMap, stripGroup.Vertexes[i].OriginalMeshVertexID)
	//}

	for _, strip := range stripGroup.Strips {
		for j := int32(0); j < strip.NumIndices; j++ {
			index := stripGroup.Indices[strip.IndexOffset+j]
			vert := stripGroup.Vertexes[index]

			meshIndices = append(meshIndices, uint16(strip.VertOffset)+vert.OriginalMeshVertexID)
		}
	}

	return meshIndices
}

func vertexDataForMesh(indices []uint16, vvd *vvd.Vvd) ([]float32, []float32, []float32, []uint32, error) {
	verts := make([]float32, 0)
	normals := make([]float32, 0)
	textureCoordinates := make([]float32, 0)
	resultantIndices := make([]uint32, len(indices))

	for _, i := range vvd.Vertices {
		verts = append(verts, i.Position.X(), i.Position.Y(), i.Position.Z())
		normals = append(normals, i.Normal.X(), i.Normal.Y(), i.Normal.Z())
		textureCoordinates = append(textureCoordinates, i.UVs.X(), i.UVs.Y())
	}

	for idx, i := range indices {
		resultantIndices[idx] = uint32(i)
	}
	return verts, normals, textureCoordinates, resultantIndices, nil
}
