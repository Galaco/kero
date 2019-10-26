package vis

import (
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/valve"
	"github.com/go-gl/mathgl/mgl32"
)

type ClusterLeaf struct {
	Id          int16
	Faces       []valve.BspFace
	StaticProps []*graphics.StaticProp
	DispFaces   []int
	Mins, Maxs  mgl32.Vec3
	Origin      mgl32.Vec3
	SkyVisible  bool
}

// groupClusterFacesByMaterial groups all faces in a collections of
// clusters by material
func GroupClusterFacesByMaterial(clusters []*ClusterLeaf) map[string][]*valve.BspFace {
	clusterFaceMap := map[string][]*valve.BspFace{}

	for _, cluster := range clusters {
		for idx, face := range cluster.Faces {
			if _, ok := clusterFaceMap[face.Material()]; !ok {
				clusterFaceMap[face.Material()] = []*valve.BspFace{&cluster.Faces[idx]}
			} else {
				clusterFaceMap[face.Material()] = append(clusterFaceMap[face.Material()], &cluster.Faces[idx])
			}
		}
	}

	return clusterFaceMap
}
