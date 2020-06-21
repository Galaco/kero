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
			if lodIdx > len(model.LODS) {
				return nil, nil, nil, nil, errors.New("invalid LOD index requested for model")
			}
			for _, mesh := range model.LODS[lodIdx].Meshes {
				i := indicesForMesh(&mesh)
				if len(i) == 0 {
					return nil, nil, nil, nil, errors.New("invalid studiomodel mesh: 0 indices. ignoring")
				}

				v, n, uv, err := vertexDataForMesh(studioModel.Vvd)
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
func indicesForMesh(mesh *vtx.Mesh) []uint32 {
	if len(mesh.StripGroups) > 1 {
		return make([]uint32, 0)
	}
	meshIndices := make([]uint32, 0)

	stripGroup := mesh.StripGroups[0]

	for _, strip := range stripGroup.Strips {
		for j := int32(0); j < strip.NumIndices; j++ {
			index := stripGroup.Indices[strip.IndexOffset+j]
			vert := stripGroup.Vertexes[index]

			//meshIndices = append(meshIndices, uint32(index))

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
