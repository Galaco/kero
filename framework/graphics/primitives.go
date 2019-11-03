package graphics

import "math"

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

type Sphere struct {
	BasicMesh
}

func NewSphere() *Sphere {
	stacks := 20
	slices := 20
	PI := 3.14

	positions := make([]float32, 0)
	//indices := make([]uint, 0)

	// loop through stacks.
	for i := 0; i <= stacks; i++ {
		V := float64(i) / float64(stacks)
		phi := V * PI

		// loop through the slices.
		for j := 0; j <= slices; j++ {

			U := float64(j) / float64(slices)
			theta := U * (PI * 2)

			// use spherical coordinates to calculate the positions.
			x := math.Cos(theta) * math.Sin(phi)
			y := math.Cos(phi)
			z := math.Sin(theta) * math.Sin(phi)

			positions = append(positions, float32(x), float32(y), float32(z))
		}
	}

	vertices := make([]float32, 0)

	// Calc The Index Positions
	for i := 0; i < slices*stacks+slices; i++ {
		vertices = append(vertices, positions[i*3:i*3+3]...)
		vertices = append(vertices, positions[(i+slices+1)*3:(i+slices+1)*3+3]...)
		vertices = append(vertices, positions[(i+slices)*3:(i+slices)*3+3]...)
		//indices = append(indices, uint(i), uint(i + slices + 1), uint(i + slices))

		vertices = append(vertices, positions[(i+slices+1)*3:(i+slices+1)*3+3]...)
		vertices = append(vertices, positions[i*3:i*3+3]...)
		vertices = append(vertices, positions[(i+1)*3:(i+1)*3+3]...)
		//indices = append(indices, uint(i + slices + 1), uint(i), uint(i + 1))
	}

	sphere := &Sphere{}
	sphere.AddVertex(vertices...)
	sphere.AddUV(make([]float32, len(vertices)/3*2)...)
	sphere.AddNormal(make([]float32, len(vertices))...)
	sphere.GenerateTangents()

	return sphere
}
