package scene

import (
	"fmt"
	"github.com/galaco/bsp/primitives/leaf"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/kero/framework/scene/vis"
	"github.com/go-gl/mathgl/mgl32"
	"io"
)

type fileSystem interface {
	GetFile(string) (io.Reader, error)
}

var sceneSingleton StaticScene

func CurrentScene() *StaticScene {
	if sceneSingleton.BspMesh == nil {
		return nil
	}

	return &sceneSingleton
}

func CloseCurrentScene() {
	sceneSingleton = StaticScene{}
}

type StaticScene struct {
	RawBsp            *graphics.Bsp
	BspMesh           *mesh.BasicMesh
	BspFaces          []graphics.BspFace
	DisplacementFaces []*graphics.BspFace
	Textures          map[string]graphics.Texture

	StaticProps []graphics.StaticProp
	Entities    []entity.IEntity

	VisData      *vis.Vis
	ClusterLeafs []vis.ClusterLeaf
	LeafCache    *vis.Cluster

	VisibleClusterLeafs []*vis.ClusterLeaf
	CurrentLeaf         *leaf.Leaf

	Camera             *graphics.Camera
	CameraPrevPosition mgl32.Vec3

	SkyboxClusterLeafs []*vis.ClusterLeaf
	SkyCamera          *graphics.Camera

	TexCache TextureCache
}

// RecomputeVisibleClusters rebuilds the current facelist to render, by first
// recalculating using vvis data
func (scene *StaticScene) RecomputeVisibleClusters() {
	if scene.Camera.Transform().Translation.ApproxEqual(scene.CameraPrevPosition) {
		return
	}
	scene.CameraPrevPosition = scene.Camera.Transform().Translation
	// View hasn't moved
	currentLeaf := scene.VisData.FindCurrentLeaf(scene.Camera.Transform().Translation)

	if scene.CurrentLeaf == currentLeaf {
		return
	}

	if currentLeaf == nil || currentLeaf.Cluster == -1 {
		scene.CurrentLeaf = currentLeaf

		scene.asyncRebuildVisibleWorld(currentLeaf)
		return
	}

	// Haven't changed cluster
	if scene.LeafCache != nil && scene.LeafCache.ClusterId == currentLeaf.Cluster {
		return
	}

	scene.CurrentLeaf = currentLeaf
	scene.LeafCache = scene.VisData.GetPVSCacheForCluster(currentLeaf.Cluster)

	scene.VisibleClusterLeafs = scene.asyncRebuildVisibleWorld(scene.CurrentLeaf)
}

// Launches rebuilding the visible world in a separate thread
// Note: This *could* cause rendering issues if the rebuild is slower than
// travelling between clusters
func (scene *StaticScene) asyncRebuildVisibleWorld(currentLeaf *leaf.Leaf) []*vis.ClusterLeaf {
	visibleWorld := make([]*vis.ClusterLeaf, 0, 1024)

	var visibleClusterIds []int16

	if currentLeaf != nil && currentLeaf.Cluster != -1 {
		visibleClusterIds = scene.VisData.PVSForCluster(currentLeaf.Cluster)
	}

	// nothing visible so render everything
	if len(visibleClusterIds) == 0 {
		for idx := range scene.ClusterLeafs {
			visibleWorld = append(visibleWorld, &scene.ClusterLeafs[idx])
		}
	} else {
		for _, clusterId := range visibleClusterIds {
			visibleWorld = append(visibleWorld, &scene.ClusterLeafs[clusterId])
		}
	}

	return visibleWorld
}

func LoadStaticSceneFromBsp(fs fileSystem,
	level *graphics.Bsp,
	entities []entity.IEntity) *StaticScene {

	texCache := NewTextureCache()

	texCache.Add(ErrorTexturePath, graphics.NewErrorTexture(ErrorTexturePath))

	if level.LightmapAtlas() != nil {
		texCache.Add(LightmapTexturePath, level.LightmapAtlas())
	} else {
		texCache.Add(LightmapTexturePath, texCache.Find(ErrorTexturePath))
	}

	// load materials
	var tex graphics.Texture
	var err error
	for _, mat := range level.MaterialDictionary() {
		if tex = texCache.Find(mat.BaseTextureName); tex == nil {
			if mat.BaseTextureName == "" {
				console.PrintString(console.LevelWarning, fmt.Sprintf("%s has no $BaseTexture", mat.FilePath()))
				texCache.Add(mat.BaseTextureName, texCache.Find(ErrorTexturePath))
			} else {
				tex, err = graphics.LoadTexture(fs, mat.BaseTextureName)
				if err != nil || tex == nil {
					if err != nil {
						console.PrintString(console.LevelWarning, err.Error())
					}
					texCache.Add(mat.BaseTextureName, texCache.Find(ErrorTexturePath))
				} else {
					texCache.Add(mat.BaseTextureName, tex)
				}
			}
		}
	}

	// generate displacement faces
	dispFaces := make([]*graphics.BspFace, 0, 1024)
	for _, i := range level.DispFaces() {
		dispFaces = append(dispFaces, &level.Faces()[i])
	}

	// finish bsp mesh

	// Add MATERIALS TO FACES
	tex = nil
	for idx, bspFace := range level.Faces() {
		if level.MaterialDictionary()[bspFace.Material()] == nil {
			console.PrintString(console.LevelWarning, fmt.Sprintf("MATERIAL: %s not found", bspFace.Material()))
			tex = texCache.Find(ErrorTexturePath)
		} else {
			if level.MaterialDictionary()[bspFace.Material()].BaseTextureName == "" {
				tex = texCache.Find(ErrorTexturePath)
			} else {
				tex = texCache.Find(level.MaterialDictionary()[bspFace.Material()].BaseTextureName)
			}
		}
		// Generate texture coordinates
		level.Mesh().AddUV(
			graphics.TexCoordsForFaceFromTexInfo(
				level.Mesh().Vertices()[bspFace.Offset()*3:(bspFace.Offset()*3)+(bspFace.Length()*3)],
				bspFace.TexInfo(),
				tex.Width(),
				tex.Height())...)

		// LightmapCoordsForFaceFromTexInfo
		if level.LightmapAtlas() != nil {
			level.Mesh().AddLightmapUV(
				graphics.LightmapCoordsForFaceFromTexInfo(
					level.Mesh().Vertices()[bspFace.Offset()*3:(bspFace.Offset()*3)+(bspFace.Length()*3)],
					bspFace.RawFace(),
					bspFace.TexInfo(),
					float32(level.LightmapAtlas().Width()),
					float32(level.LightmapAtlas().Height()),
					level.LightmapAtlas().AtlasEntry(idx).X,
					level.LightmapAtlas().AtlasEntry(idx).Y)...)
		}

	}

	level.Mesh().GenerateTangents()

	remappedFaces := make([]graphics.BspFace, 0, 1024)
	// Kero isn't interested in tools faces (for now)
	for idx := range level.Faces() {
		remappedFaces = append(remappedFaces, level.Faces()[idx])
	}

	// Generate visibility tree
	visibility := vis.LoadVisData(level.File())
	clusterLeafs := generateClusterLeafs(level, visibility)

	var skyCameraEntity entity.IEntity
	var infoPlayerStart entity.IEntity
	for idx, e := range entities {
		if e.Classname() == "sky_camera" {
			skyCameraEntity = entities[idx]
			continue
		}
		if e.Classname() == "info_player_start" {
			infoPlayerStart = entities[idx]
			continue
		}
	}
	var skyCamera *graphics.Camera

	if skyCameraEntity != nil {
		skyCamera = graphics.NewCamera(level.Camera().Fov(), level.Camera().AspectRatio())
		skyCamera.Transform().Translation = skyCameraEntity.VectorForKey("origin")
		scale := skyCameraEntity.FloatForKey("scale")
		skyCamera.Transform().Scale = mgl32.Vec3{scale, scale, scale}
	}
	if infoPlayerStart != nil {
		level.Camera().Transform().Translation = infoPlayerStart.VectorForKey("origin")
		angles := infoPlayerStart.VectorForKey("angles")
		level.Camera().Transform().Orientation = mgl32.AnglesToQuat(angles[0], angles[1], angles[2], mgl32.XYZ)
	}

	sceneSingleton = StaticScene{
		RawBsp:             level,
		BspMesh:            level.Mesh(),
		BspFaces:           remappedFaces,
		DisplacementFaces:  dispFaces,
		Entities:           entities,
		StaticProps:        level.StaticProps,
		ClusterLeafs:       clusterLeafs,
		VisData:            visibility,
		Camera:             level.Camera(),
		CameraPrevPosition: mgl32.Vec3{99999, 99999, 99999},
		SkyCamera:          skyCamera,
		TexCache:           texCache,
	}

	// Generate Initial visibility data
	sceneSingleton.asyncRebuildVisibleWorld(nil)
	if skyCamera != nil {
		sceneSingleton.SkyboxClusterLeafs = sceneSingleton.asyncRebuildVisibleWorld(sceneSingleton.VisData.FindCurrentLeaf(skyCamera.Transform().Translation))
	}

	return CurrentScene()
}

func generateClusterLeafs(level *graphics.Bsp, visData *vis.Vis) []vis.ClusterLeaf {
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
			bspClusters[bspLeaf.Cluster].DebugMesh = mesh.NewCuboidFromMinMaxs(bspClusters[bspLeaf.Cluster].Mins, bspClusters[bspLeaf.Cluster].Maxs)

			if bspLeaf.Flags()&leaf.LeafFlagsSky > 0 {
				bspClusters[bspLeaf.Cluster].SkyVisible = true
			}
		}
	}

	// Assign staticprops to clusters
	for idx, prop := range level.StaticProps {
		for _, leafId := range prop.LeafList() {
			clusterId := visData.Leafs[leafId].Cluster
			if clusterId == -1 {
				//defaultCluster.StaticProps = append(defaultCluster.StaticProps, &baseWorldStaticProps[idx])
				continue
			}
			bspClusters[clusterId].StaticProps = append(bspClusters[clusterId].StaticProps, &level.StaticProps[idx])
		}
	}

	//for _, idx := range bspClusters[0].DispFaces {
	//	defaultCluster.Faces = append(defaultCluster.Faces, baseWorldBspFaces[idx])
	//}

	return bspClusters
}
