package mesh

import "github.com/go-gl/mathgl/mgl32"

var cubeVerts = []float32{
	-1.0, -1.0, -1.0,
	1.0, -1.0, -1.0,
	-1.0, -1.0, 1.0,
	1.0, -1.0, -1.0,
	1.0, -1.0, 1.0,
	-1.0, -1.0, 1.0,
	-1.0, 1.0, -1.0,
	-1.0, 1.0, 1.0,
	1.0, 1.0, -1.0,
	1.0, 1.0, -1.0,
	-1.0, 1.0, 1.0,
	1.0, 1.0, 1.0,
	-1.0, -1.0, 1.0,
	1.0, -1.0, 1.0,
	-1.0, 1.0, 1.0,
	1.0, -1.0, 1.0,
	1.0, 1.0, 1.0,
	-1.0, 1.0, 1.0,
	-1.0, -1.0, -1.0,
	-1.0, 1.0, -1.0,
	1.0, -1.0, -1.0,
	1.0, -1.0, -1.0,
	-1.0, 1.0, -1.0,
	1.0, 1.0, -1.0,
	-1.0, -1.0, 1.0,
	-1.0, 1.0, -1.0,
	-1.0, -1.0, -1.0,
	-1.0, -1.0, 1.0,
	-1.0, 1.0, 1.0,
	-1.0, 1.0, -1.0,
	1.0, -1.0, 1.0,
	1.0, -1.0, -1.0,
	1.0, 1.0, -1.0,
	1.0, -1.0, 1.0,
	1.0, 1.0, -1.0,
	1.0, 1.0, 1.0,
}

var cubeNormals = cubeVerts

var cubeUVs = []float32{
	0, 0,
	1, 0,
	0, 1,
	1, 0,
	0, 1,
	1, 1,

	0, 0,
	1, 0,
	0, 1,
	1, 0,
	0, 1,
	1, 1,

	0, 0,
	1, 0,
	0, 1,
	1, 0,
	0, 1,
	1, 1,

	0, 0,
	1, 0,
	0, 1,
	1, 0,
	0, 1,
	1, 1,
	0, 0,
	1, 0,
	0, 1,
	1, 0,
	0, 1,
	1, 1,
	0, 0,
	1, 0,
	0, 1,
	1, 0,
	0, 1,
	1, 1,
}

// Cube
type Cube struct {
	BasicMesh
}

// NewCube
func NewCube() *Cube {
	c := &Cube{}
	c.AddVertex(cubeVerts...)
	c.AddNormal(cubeNormals...)
	c.AddUV(cubeUVs...)
	c.GenerateTangents()

	return c
}

func NewCuboidFromMinMaxs(mins, maxs mgl32.Vec3) *Cube {
	c := &Cube{}
	verts := make([]float32, 0, len(cubeVerts))
	verts = append(verts, cubeVerts...)

	for idx, vert := range verts {
		if vert == -1 {
			verts[idx] = mins[idx%3]
		} else {
			verts[idx] = maxs[idx%3]
		}
	}

	c.AddVertex(verts...)
	c.AddNormal(cubeNormals...)
	c.AddUV(cubeUVs...)
	c.GenerateTangents()

	return c
}
