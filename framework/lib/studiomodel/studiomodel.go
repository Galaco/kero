package studiomodel

import (
	"errors"
	"github.com/galaco/StudioModel"
	"github.com/galaco/StudioModel/vtx"
	"github.com/galaco/StudioModel/vvd"
)

// VertexDataForModel loads model vertex data
func VertexDataForModel(studioModel *studiomodel.StudioModel, lodIdx int) ([][]float32, [][]float32, [][]float32, error) {
	vertices := make([][]float32, 0)
	normals := make([][]float32, 0)
	textureCoordinates := make([][]float32, 0)
	for _, bodyPart := range studioModel.Vtx.BodyParts {
		for _, model := range bodyPart.Models {
			if len(model.LODS) < lodIdx {
				return nil, nil, nil, errors.New("invalid LOD index requested for model")
			}
			for _, mesh := range model.LODS[lodIdx].Meshes {
				indices := indicesForMesh(&mesh)
				if len(indices) == 0 {
					continue
				}

				v, n, uv, err := vertexDataForMesh(indices, studioModel.Vvd)
				if err != nil {
					return nil, nil, nil, err
				}
				vertices = append(vertices, v)
				normals = append(normals, n)
				textureCoordinates = append(textureCoordinates, uv)
			}
		}
	}

	return vertices, normals, textureCoordinates, nil
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

func vertexDataForMesh(indices []uint16, vvd *vvd.Vvd) ([]float32, []float32, []float32, error) {
	verts := make([]float32, 0)
	normals := make([]float32, 0)
	textureCoordinates := make([]float32, 0)

	for _, index := range indices {
		if int(index) > len(vvd.Vertices) {
			return nil, nil, nil, errors.New("vertex data bounds out of range")
		}
		vvdVert := &vvd.Vertices[index]

		verts = append(verts, vvdVert.Position.X(), vvdVert.Position.Y(), vvdVert.Position.Z())
		normals = append(normals, vvdVert.Normal.X(), vvdVert.Normal.Y(), vvdVert.Normal.Z())
		textureCoordinates = append(textureCoordinates, vvdVert.UVs.X(), vvdVert.UVs.Y())
	}
	return verts, normals, textureCoordinates, nil
}
