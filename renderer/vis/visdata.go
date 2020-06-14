package vis

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/leaf"
	"github.com/galaco/bsp/primitives/node"
	"github.com/galaco/bsp/primitives/plane"
	"github.com/galaco/bsp/primitives/visibility"
	"github.com/go-gl/mathgl/mgl32"
)

func LoadVisData(file *bsp.Bsp) *Vis {
	return NewVisFromBSP(file)
}

type Cluster struct {
	Leafs      []uint16
	ClusterId  int16
	SkyVisible bool
	Faces      []uint16
}

type Vis struct {
	Cluster        []Cluster
	VisibilityLump *visibility.Vis
	Leafs          []leaf.Leaf
	LeafFaces      []uint16
	Nodes          []node.Node
	Planes         []plane.Plane

	viewPosition    mgl32.Vec3
	viewCurrentLeaf *leaf.Leaf
}

func (vis *Vis) PVSForCluster(clusterId int16) []int16 {
	return vis.VisibilityLump.GetVisibleClusters(clusterId)
}

func (vis *Vis) GetPVSCacheForCluster(clusterId int16) *Cluster {
	if clusterId == -1 {
		clusterId = int16(vis.findCurrentLeafIndex(vis.viewPosition))
	}
	for _, cacheEntry := range vis.Cluster {
		if cacheEntry.ClusterId == clusterId {
			return &cacheEntry
		}
	}
	return vis.cachePVSForCluster(clusterId)
}

// Cluster visible data for current cluster
func (vis *Vis) cachePVSForCluster(clusterId int16) *Cluster {
	clusterList := vis.VisibilityLump.GetPVSForCluster(clusterId)

	skyVisible := false

	faces := make([]uint16, 0)
	leafs := make([]uint16, 0)
	for idx, l := range vis.Leafs {
		//Check if cluster is in pvs
		if !vis.clusterVisible(&clusterList, l.Cluster) {
			continue
		}
		if l.Flags()&leaf.LeafFlagsSky > 0 {
			skyVisible = true
		}
		leafs = append(leafs, uint16(idx))
		faces = append(faces, vis.LeafFaces[l.FirstLeafFace:l.FirstLeafFace+l.NumLeafFaces]...)
	}

	cache := Cluster{
		ClusterId:  clusterId,
		Faces:      faces,
		Leafs:      leafs,
		SkyVisible: skyVisible,
	}

	vis.Cluster = append(vis.Cluster, cache)

	return &cache
}

// Determine if a cluster is visible
func (vis *Vis) clusterVisible(pvs *[]bool, leafCluster int16) bool {
	if leafCluster < 0 {
		return true
	}

	if (*pvs)[leafCluster] {
		return true
	}

	return false
}

// Test if the camera has moved, and find the current leaf if so
func (vis *Vis) FindCurrentLeaf(position mgl32.Vec3) *leaf.Leaf {
	vis.viewPosition = position
	vis.viewCurrentLeaf = &vis.Leafs[vis.findCurrentLeafIndex(position)]
	return vis.viewCurrentLeaf
}

// Find the index into the leaf array for the leaf the player
// is inside of
// Based on: https://bitbucket.org/fallahn/chuf-arc
func (vis *Vis) findCurrentLeafIndex(position mgl32.Vec3) int32 {
	i := int32(0)

	//walk the bsp to find the index of the leaf which contains our position
	for i >= 0 {
		node := &vis.Nodes[i]
		plane := vis.Planes[node.PlaneNum]

		//check which side of the plane the position is on so we know which direction to go
		distance := plane.Normal.X()*position.X() + plane.Normal.Y()*position.Y() + plane.Normal.Z()*position.Z() - plane.Distance
		i = node.Children[0]
		if distance < 0 {
			i = node.Children[1]
		}
	}

	return ^i
}

func NewVisFromBSP(file *bsp.Bsp) *Vis {
	return &Vis{
		VisibilityLump: file.Lump(bsp.LumpVisibility).(*lumps.Visibility).GetData(),
		viewPosition:   mgl32.Vec3{65536, 65536, 65536},
		Leafs:          file.Lump(bsp.LumpLeafs).(*lumps.Leaf).GetData(),
		LeafFaces:      file.Lump(bsp.LumpLeafFaces).(*lumps.LeafFace).GetData(),
		Nodes:          file.Lump(bsp.LumpNodes).(*lumps.Node).GetData(),
		Planes:         file.Lump(bsp.LumpPlanes).(*lumps.Planes).GetData(),
	}
}
