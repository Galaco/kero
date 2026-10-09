package vis

import (
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/go-gl/mathgl/mgl32"
)

type ClusterLeaf struct {
	Id          int16
	Faces       []graphics.BspFace
	StaticProps []*graphics.StaticProp
	DispFaces   []int // indexes of the bsp's displacement faces that are in this cluster
	Mins, Maxs  mgl32.Vec3
	Origin      mgl32.Vec3
	SkyVisible  bool
	DebugMesh   mesh.Mesh
}

// GroupClusterFacesByMaterial groups all faces in a collections of
// clusters by material
func GroupClusterFacesByMaterial(clusters []*ClusterLeaf) map[string][]*graphics.BspFace {
	clusterFaceMap := map[string][]*graphics.BspFace{}
	// A face is listed in every cluster it is in, but is drawn once. Faces are identified by where they are in the mesh.
	added := map[int]bool{}

	for _, cluster := range clusters {
		for idx, face := range cluster.Faces {
			if added[face.Offset()] {
				continue
			}
			added[face.Offset()] = true
			if _, ok := clusterFaceMap[face.Material()]; !ok {
				clusterFaceMap[face.Material()] = []*graphics.BspFace{&cluster.Faces[idx]}
			} else {
				clusterFaceMap[face.Material()] = append(clusterFaceMap[face.Material()], &cluster.Faces[idx])
			}
		}
	}

	return clusterFaceMap
}

// ClusterDisplacements returns the displacements in a collection of clusters, as indexes of the bsp's displacement
// faces. A displacement in more than one cluster is returned once.
func ClusterDisplacements(clusters []*ClusterLeaf) []int {
	displacements := make([]int, 0)
	added := map[int]bool{}
	for _, cluster := range clusters {
		for _, idx := range cluster.DispFaces {
			if !added[idx] {
				added[idx] = true
				displacements = append(displacements, idx)
			}
		}
	}
	return displacements
}
