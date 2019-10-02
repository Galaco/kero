package vis

import (
	"github.com/galaco/kero/valve"
	"github.com/go-gl/mathgl/mgl32"
)

type ClusterLeaf struct {
	Id         int16
	Faces      []valve.BspFace
	DispFaces  []int
	Mins, Maxs mgl32.Vec3
	Origin     mgl32.Vec3
}
