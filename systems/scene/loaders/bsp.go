package loader

import (
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/dispinfo"
	"github.com/galaco/bsp/primitives/dispvert"
	"github.com/galaco/bsp/primitives/face"
	"github.com/galaco/bsp/primitives/plane"
	"github.com/galaco/bsp/primitives/texinfo"
	"github.com/galaco/kero/framework/valve"
	"github.com/go-gl/mathgl/mgl32"
)

type bspstructs struct {
	faces     []face.Face
	planes    []plane.Plane
	vertexes  []mgl32.Vec3
	surfEdges []int32
	edges     [][2]uint16
	texInfos  []texinfo.TexInfo
	dispInfos []dispinfo.DispInfo
	dispVerts []dispvert.DispVert
	game      *lumps.Game
}

// LoadBspMap is the gateway into loading the core static level. Entities are loaded
// elsewhere
// It loads in the following order:
// BSP Geometry
// BSP Materials
// StaticProps (materials loaded as required)
func LoadBspMap(filename string) (*valve.Bsp, error) {
	level,err := valve.LoadBspMap(filename)

	return level,err

	// Generate Texture list from materials

	// Load textures

	// Generate uv data from texture and bsp
	//bspMesh.AddUV(
	//	texCoordsForFaceFromTexInfo(
	//		bspMesh.Vertices()[bspFace.Offset()*3:(bspFace.Offset()*3)+(bspFace.Length()*3)],
	//		&bspStructure.texInfos[bspStructure.faces[idx].TexInfo], mat.Width(), mat.Height())...)
}
