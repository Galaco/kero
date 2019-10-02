package renderer

import (
	"fmt"
	"github.com/galaco/bsp/primitives/leaf"
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems/renderer/cache"
	"github.com/galaco/kero/systems/renderer/vis"
	"github.com/galaco/kero/valve"
	"github.com/go-gl/mathgl/mgl32"
	"strings"
)

type SceneGraph struct {
	bspMesh  *graphics.Mesh
	bspFaces []valve.BspFace

	gpuMesh graphics.GpuMesh

	visData      *vis.Vis
	clusterLeafs []vis.ClusterLeaf
	LeafCache    *vis.Cluster

	visibleClusterLeafs []*vis.ClusterLeaf
	currentLeaf         *leaf.Leaf

	camera             *graphics3d.Camera
	cameraPrevPosition mgl32.Vec3
}

// RecomputeVisibleClusters rebuilds the current facelist to render, by first
// recalculating using vvis data
func (scene *SceneGraph) RecomputeVisibleClusters() {
	if scene.camera.Transform().Position.ApproxEqual(scene.cameraPrevPosition) {
		return
	}
	scene.cameraPrevPosition = scene.camera.Transform().Position
	// View hasn't moved
	currentLeaf := scene.visData.FindCurrentLeaf(scene.camera.Transform().Position)

	if currentLeaf == nil || currentLeaf.Cluster == -1 {
		scene.currentLeaf = currentLeaf

		scene.asyncRebuildVisibleWorld(scene.currentLeaf)
		return
	}

	// Haven't changed cluster
	if scene.LeafCache != nil && scene.LeafCache.ClusterId == currentLeaf.Cluster {
		return
	}

	scene.currentLeaf = currentLeaf
	scene.LeafCache = scene.visData.GetPVSCacheForCluster(currentLeaf.Cluster)

	scene.asyncRebuildVisibleWorld(scene.currentLeaf)
}

// Launches rebuilding the visible world in a separate thread
// Note: This *could* cause rendering issues if the rebuild is slower than
// travelling between clusters
func (scene *SceneGraph) asyncRebuildVisibleWorld(currentLeaf *leaf.Leaf) {
	visibleWorld := make([]*vis.ClusterLeaf, 0, 1024)

	var visibleClusterIds []int16

	if currentLeaf != nil && currentLeaf.Cluster != -1 {
		visibleClusterIds = scene.visData.PVSForCluster(currentLeaf.Cluster)
	}

	// nothing visible so render everything
	if len(visibleClusterIds) == 0 {
		for idx := range scene.clusterLeafs {
			visibleWorld = append(visibleWorld, &scene.clusterLeafs[idx])
		}
	} else {
		for _, clusterId := range visibleClusterIds {
			visibleWorld = append(visibleWorld, &scene.clusterLeafs[clusterId])
		}
	}

	scene.visibleClusterLeafs = visibleWorld
}

func NewSceneGraphFromBsp(fs filesystem.FileSystem, level *valve.Bsp, materialCache *cache.Material, texCache *cache.Texture, gpuItemCache *cache.GpuItem) *SceneGraph {
	texCache.Add(cache.ErrorTexturePath, graphics.NewErrorTexture(cache.ErrorTexturePath))
	gpuItemCache.Add(cache.ErrorTexturePath, graphics.UploadTexture(texCache.Find(cache.ErrorTexturePath)))

	// load materials
	for _, mat := range level.MaterialDictionary() {
		if tex := texCache.Find(mat.BaseTextureName); tex == nil {
			tex, err := graphics.LoadTexture(fs, mat.BaseTextureName)
			if err != nil {
				event.Dispatch(messages.NewConsoleMessage(console.LevelWarning, err.Error()))
				texCache.Add(mat.BaseTextureName, texCache.Find(cache.ErrorTexturePath))
				gpuItemCache.Add(mat.BaseTextureName, gpuItemCache.Find(cache.ErrorTexturePath))
			} else {
				texCache.Add(mat.BaseTextureName, tex)
				gpuItemCache.Add(mat.BaseTextureName, graphics.UploadTexture(tex))
			}
		}
		materialCache.Add(strings.ToLower(mat.FilePath()), cache.NewGpuMaterial(gpuItemCache.Find(mat.BaseTextureName)))
	}

	// finish bsp mesh
	// Add MATERIALS TO FACES
	for _, bspFace := range level.Faces() {
		// Generate texture coordinates
		var tex *graphics.Texture
		if level.MaterialDictionary()[bspFace.Material()] == nil {
			event.Dispatch(messages.NewConsoleMessage(console.LevelWarning, fmt.Sprintf("MATERIAL: %s not found", bspFace.Material())))
			tex = texCache.Find(cache.ErrorTexturePath)
		} else {
			if level.MaterialDictionary()[bspFace.Material()].BaseTextureName == "" {
				tex = texCache.Find(cache.ErrorTexturePath)
			} else {
				tex = texCache.Find(level.MaterialDictionary()[bspFace.Material()].BaseTextureName)
			}
		}
		level.Mesh().AddUV(
			valve.TexCoordsForFaceFromTexInfo(
				level.Mesh().Vertices()[bspFace.Offset()*3:(bspFace.Offset()*3)+(bspFace.Length()*3)],
				bspFace.TexInfo(),
				tex.Width(),
				tex.Height())...)
	}

	level.Mesh().GenerateTangents()

	remappedFaces := make([]valve.BspFace, 0, 1024)
	// Kero isnt interested in tools faces (for now)
	for idx, bspFace := range level.Faces() {
		if strings.HasPrefix(strings.ToLower(bspFace.Material()), "tools") {
			continue
		}
		remappedFaces = append(remappedFaces, level.Faces()[idx])
	}

	// Generate visibility tree
	visibility := vis.LoadVisData(level.File())
	clusterLeafs := generateClusterLeafs(level, visibility)

	return &SceneGraph{
		bspMesh:      level.Mesh(),
		gpuMesh:      graphics.UploadMesh(level.Mesh()),
		bspFaces:     remappedFaces,
		clusterLeafs: clusterLeafs,
		visData:      visibility,
		camera:       level.Camera(),
	}
}

func generateClusterLeafs(level *valve.Bsp, visData *vis.Vis) []vis.ClusterLeaf {
	bspClusters := make([]vis.ClusterLeaf, visData.VisibilityLump.NumClusters)
	//defaultCluster := vis.ClusterLeaf{
	//	Id: 32767,
	//}
	for _, bspLeaf := range visData.Leafs {
		for _, leafFace := range visData.LeafFaces[bspLeaf.FirstLeafFace : bspLeaf.FirstLeafFace+bspLeaf.NumLeafFaces] {
			if bspLeaf.Cluster == -1 {
				//defaultCluster.Faces = append(defaultCluster.Faces, bspFaces[leafFace])
				continue
			}
			bspClusters[bspLeaf.Cluster].Id = bspLeaf.Cluster
			bspClusters[bspLeaf.Cluster].Faces = append(bspClusters[bspLeaf.Cluster].Faces, level.Faces()[leafFace])
			bspClusters[bspLeaf.Cluster].Mins = mgl32.Vec3{
				float32(bspLeaf.Mins[0]),
				float32(bspLeaf.Mins[1]),
				float32(bspLeaf.Mins[2]),
			}
			bspClusters[bspLeaf.Cluster].Maxs = mgl32.Vec3{
				float32(bspLeaf.Maxs[0]),
				float32(bspLeaf.Maxs[1]),
				float32(bspLeaf.Maxs[2]),
			}
			bspClusters[bspLeaf.Cluster].Origin = bspClusters[bspLeaf.Cluster].Mins.Add(bspClusters[bspLeaf.Cluster].Maxs.Sub(bspClusters[bspLeaf.Cluster].Mins))
		}
	}

	//// Assign staticprops to clusters
	//for idx, prop := range baseWorld.StaticProps() {
	//	for _, leafId := range prop.LeafList() {
	//		clusterId := visData.Leafs[leafId].Cluster
	//		if clusterId == -1 {
	//			defaultCluster.StaticProps = append(defaultCluster.StaticProps, &baseWorldStaticProps[idx])
	//			continue
	//		}
	//		bspClusters[clusterId].StaticProps = append(bspClusters[clusterId].StaticProps, &baseWorldStaticProps[idx])
	//	}
	//}
	//
	//for _, idx := range baseWorldBsp.ClusterLeafs()[0].DispFaces {
	//	defaultCluster.Faces = append(defaultCluster.Faces, baseWorldBspFaces[idx])
	//}

	return bspClusters
}
