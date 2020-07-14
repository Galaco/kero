package mesh

import (
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/go-gl/mathgl/mgl32"
)

type Mesh adapter.Mesh

// BasicMesh
type BasicMesh struct {
	vertices    []float32
	normals     []float32
	uvs         []float32
	lightmapUVs []float32
	tangents    []float32
	indices     []uint32
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

// Indices
func (mesh *BasicMesh) Indices() []uint32 {
	return mesh.indices
}

// GenerateTangents
func (mesh *BasicMesh) GenerateTangents() {
	//const vector<vec3> & points,
	//const vector<vec3> & normals,
	//const vector<int> & faces,
	//const vector<vec2> & texCoords,
	//	vector<vec4> & tangents)
	//{
	//vector<vec3> tan1Accum;
	tan1Accum := make([]float32, len(mesh.vertices))
	//vector<vec3> tan2Accum;
	tan2Accum := make([]float32, len(mesh.vertices))
	tangents := make([]float32, len(mesh.vertices)+(len(mesh.vertices)/3))

	//for( uint i = 0; i < points.size(); i++ ) {
	//tan1Accum.push_back(vec3(0.0f));
	//tan2Accum.push_back(vec3(0.0f));
	//tangents.push_back(vec4(0.0f));
	//}

	// Compute the tangent vector
	for i := uint(0); i < uint(len(mesh.vertices))-9; i += 9 {
		rootIdx := i / 3
		p1 := mgl32.Vec3{mesh.vertices[i], mesh.vertices[i+1], mesh.vertices[i+2]}
		p2 := mgl32.Vec3{mesh.vertices[i+3], mesh.vertices[i+4], mesh.vertices[i+5]}
		p3 := mgl32.Vec3{mesh.vertices[i+6], mesh.vertices[i+7], mesh.vertices[i+8]}

		uvIdx := rootIdx * 2
		tc1 := mgl32.Vec2{mesh.uvs[uvIdx], mesh.uvs[uvIdx+1]}
		tc2 := mgl32.Vec2{mesh.uvs[uvIdx+2], mesh.uvs[uvIdx+3]}
		tc3 := mgl32.Vec2{mesh.uvs[uvIdx+4], mesh.uvs[uvIdx+5]}

		q1 := p2.Sub(p1)
		q2 := p3.Sub(p1)
		s1 := tc2.X() - tc1.X()
		s2 := tc3.X() - tc1.X()
		t1 := tc2.Y() - tc1.Y()
		t2 := tc3.Y() - tc1.Y()
		r := 1.0 / (s1*t2 - s2*t1)
		tan1 := mgl32.Vec3{
			(t2*q1.X() - t1*q2.X()) * r,
			(t2*q1.Y() - t1*q2.Y()) * r,
			(t2*q1.Z() - t1*q2.Z()) * r,
		}

		tan2 := mgl32.Vec3{
			(s1*q2.X() - s2*q1.X()) * r,
			(s1*q2.Y() - s2*q1.Y()) * r,
			(s1*q2.Z() - s2*q1.Z()) * r,
		}
		tan1Accum[i] += tan1.X()
		tan1Accum[i+1] += tan1.Y()
		tan1Accum[i+2] += tan1.Z()
		tan2Accum[i] += tan2.X()
		tan2Accum[i+1] += tan2.Y()
		tan2Accum[i+2] += tan2.Z()
	}

	for i := uint(0); i < uint(len(mesh.vertices))-2; i++ {
		n := mgl32.Vec3{
			mesh.normals[i],
			mesh.normals[i+1],
			mesh.normals[i+2],
		}
		t1 := mgl32.Vec3{
			tan1Accum[i],
			tan1Accum[i+1],
			tan1Accum[i+2],
		}
		t2 := mgl32.Vec3{
			tan2Accum[i],
			tan2Accum[i+1],
			tan2Accum[i+2],
		}
		//const vec3 &n = normals[i];
		//vec3 &t1 = tan1Accum[i];
		//vec3 &t2 = tan2Accum[i];

		// Gram-Schmidt orthogonalize
		//tangents[i] = vec4(glm::normalize( t1 - (glm::dot(n,t1) * n) ), 0.0f);
		res := t1.Sub(n.Mul(n.Dot(t1))).Normalize()
		tangents[i] = res.X()
		tangents[i+1] = res.Y()
		tangents[i+2] = res.Z()
		// Store handedness in w
		w := float32(1.0)
		if n.Cross(t1).Dot(t2) < 0 {
			w = -1.0
		}
		tangents[i+3] = w
		//tangents[i] = (glm::dot( glm::cross(n,t1), t2 ) < 0.0f) ? -1.0f : 1.0f;
	}

	//tan1Accum.clear();
	//tan2Accum.clear();

	mesh.tangents = tangents
}

// NewMesh
func NewMesh() *BasicMesh {
	return &BasicMesh{}
}
