package mesh

import (
	"math"

	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/go-gl/mathgl/mgl32"
)

type Mesh adapter.Mesh

// BasicMesh
type BasicMesh struct {
	vertices     []float32
	normals      []float32
	uvs          []float32
	lightmapUVs  []float32
	tangents     []float32
	blendWeights []float32 // 1 float per vertex for 2-texture displacement blending
	indices      []uint32
}

// AddVertex
func (mesh *BasicMesh) AddVertex(vertex ...float32) {
	mesh.vertices = append(mesh.vertices, vertex...)
}

// AddNormal
func (mesh *BasicMesh) AddNormal(normal ...float32) {
	mesh.normals = append(mesh.normals, normal...)
}

// AddUV
func (mesh *BasicMesh) AddUV(uv ...float32) {
	mesh.uvs = append(mesh.uvs, uv...)
}

// AddLightmapUV
func (mesh *BasicMesh) AddLightmapUV(uv ...float32) {
	mesh.lightmapUVs = append(mesh.lightmapUVs, uv...)
}

// AddTangent
func (mesh *BasicMesh) AddTangent(tangent ...float32) {
	mesh.tangents = append(mesh.tangents, tangent...)
}

// AddIndice
func (mesh *BasicMesh) AddIndice(indice ...uint32) {
	mesh.indices = append(mesh.indices, indice...)
}

// Vertices
func (mesh *BasicMesh) Vertices() []float32 {
	return mesh.vertices
}

// Normals
func (mesh *BasicMesh) Normals() []float32 {
	return mesh.normals
}

// UVs
func (mesh *BasicMesh) UVs() []float32 {
	return mesh.uvs
}

// LightmapUVs
func (mesh *BasicMesh) LightmapUVs() []float32 {
	return mesh.lightmapUVs
}

// Tangents
func (mesh *BasicMesh) Tangents() []float32 {
	return mesh.tangents
}

// AddBlendWeight adds a single blend weight for a vertex (for 2-texture displacement blending)
// The weight represents the alpha value for blending between basetexture and basetexture2
// 0.0 = full basetexture, 1.0 = full basetexture2
func (mesh *BasicMesh) AddBlendWeight(weight float32) {
	mesh.blendWeights = append(mesh.blendWeights, weight)
}

// BlendWeights returns the blend weight data
func (mesh *BasicMesh) BlendWeights() []float32 {
	return mesh.blendWeights
}

// Indices
func (mesh *BasicMesh) Indices() []uint32 {
	return mesh.indices
}

// GenerateTangents computes a tangent for every vertex from the triangles it is part of and their texture
// coordinates. Each tangent is 4 floats: its direction, then the handedness of the bitangent (1 or -1).
// Triangles are read from the indices, or from consecutive vertices if there are none.
func (mesh *BasicMesh) GenerateTangents() {
	numVertices := len(mesh.vertices) / 3
	tangents := make([]float32, numVertices*4)
	mesh.tangents = tangents
	if len(mesh.normals) < numVertices*3 || len(mesh.uvs) < numVertices*2 {
		return
	}

	position := func(v int) mgl32.Vec3 {
		return mgl32.Vec3{mesh.vertices[v*3], mesh.vertices[v*3+1], mesh.vertices[v*3+2]}
	}
	uv := func(v int) mgl32.Vec2 {
		return mgl32.Vec2{mesh.uvs[v*2], mesh.uvs[v*2+1]}
	}

	// Sum the directions of increasing u (tan1) and v (tan2) of every triangle that uses a vertex
	tan1Accum := make([]mgl32.Vec3, numVertices)
	tan2Accum := make([]mgl32.Vec3, numVertices)
	addTriangle := func(a, b, c int) {
		if a >= numVertices || b >= numVertices || c >= numVertices {
			return
		}
		q1 := position(b).Sub(position(a))
		q2 := position(c).Sub(position(a))
		st1 := uv(b).Sub(uv(a))
		st2 := uv(c).Sub(uv(a))
		det := st1.X()*st2.Y() - st2.X()*st1.Y()
		if det == 0 {
			// No texture area, so no texture direction
			return
		}
		r := 1 / det
		tan1 := q1.Mul(st2.Y()).Sub(q2.Mul(st1.Y())).Mul(r)
		tan2 := q2.Mul(st1.X()).Sub(q1.Mul(st2.X())).Mul(r)
		if !isFinite(tan1) || !isFinite(tan2) {
			return
		}
		for _, v := range [3]int{a, b, c} {
			tan1Accum[v] = tan1Accum[v].Add(tan1)
			tan2Accum[v] = tan2Accum[v].Add(tan2)
		}
	}
	if len(mesh.indices) > 0 {
		for i := 0; i+2 < len(mesh.indices); i += 3 {
			addTriangle(int(mesh.indices[i]), int(mesh.indices[i+1]), int(mesh.indices[i+2]))
		}
	} else {
		for v := 0; v+2 < numVertices; v += 3 {
			addTriangle(v, v+1, v+2)
		}
	}

	for v := 0; v < numVertices; v++ {
		n := mgl32.Vec3{mesh.normals[v*3], mesh.normals[v*3+1], mesh.normals[v*3+2]}
		if n.Len() > 0 {
			n = n.Normalize()
		}

		// Gram-Schmidt orthogonalize against the normal
		tangent := tan1Accum[v].Sub(n.Mul(n.Dot(tan1Accum[v])))
		if tangent.Len() < 1e-6 || !isFinite(tangent) {
			tangent = perpendicular(n)
		}
		tangent = tangent.Normalize()

		w := float32(1)
		if n.Cross(tangent).Dot(tan2Accum[v]) < 0 {
			w = -1
		}
		tangents[v*4] = tangent.X()
		tangents[v*4+1] = tangent.Y()
		tangents[v*4+2] = tangent.Z()
		tangents[v*4+3] = w
	}
}

// perpendicular returns a unit vector perpendicular to n, or the x axis if n is zero
func perpendicular(n mgl32.Vec3) mgl32.Vec3 {
	if n.Len() == 0 {
		return mgl32.Vec3{1, 0, 0}
	}
	if math.Abs(float64(n.X())) < 0.9 {
		return n.Cross(mgl32.Vec3{1, 0, 0}).Normalize()
	}
	return n.Cross(mgl32.Vec3{0, 1, 0}).Normalize()
}

func isFinite(v mgl32.Vec3) bool {
	for _, c := range v {
		if math.IsNaN(float64(c)) || math.IsInf(float64(c), 0) {
			return false
		}
	}
	return true
}

// NewMesh
func NewMesh() *BasicMesh {
	return &BasicMesh{}
}
