package renderer

import (
	"fmt"
	"github.com/galaco/bsp/primitives/leaf"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/kero/renderer/cache"
	"github.com/galaco/kero/renderer/scene"
	"github.com/galaco/kero/renderer/vis"
	"github.com/go-gl/mathgl/mgl32"
	"io"
	"strings"
)

type fileSystem interface {
	GetFile(string) (io.Reader, error)
}

type StaticScene struct {
	bspMesh           *mesh.BasicMesh
	bspFaces          []graphics.BspFace
	displacementFaces []*graphics.BspFace
	skybox            *scene.Skybox

	gpuMesh     adapter.GpuMesh
	staticProps []graphics.StaticProp
	entities    []entity.IEntity

	visData      *vis.Vis
	clusterLeafs []vis.ClusterLeaf
	LeafCache    *vis.Cluster

	visibleClusterLeafs []*vis.ClusterLeaf
	currentLeaf         *leaf.Leaf

	camera             *graphics.Camera
	cameraPrevPosition mgl32.Vec3

	skyboxClusterLeafs []*vis.ClusterLeaf
	skyCamera          *graphics.Camera
}

// RecomputeVisibleClusters rebuilds the current facelist to render, by first
// recalculating using vvis data
func (scene *StaticScene) RecomputeVisibleClusters() {
	if scene.camera.Transform().Position.ApproxEqual(scene.cameraPrevPosition) {
		return
	}
	scene.cameraPrevPosition = scene.camera.Transform().Position
	// View hasn't moved
	currentLeaf := scene.visData.FindCurrentLeaf(scene.camera.Transform().Position)

	if scene.currentLeaf == currentLeaf {
		return
	}

	if currentLeaf == nil || currentLeaf.Cluster == -1 {
		scene.currentLeaf = currentLeaf

		scene.asyncRebuildVisibleWorld(currentLeaf)
		return
	}

	// Haven't changed cluster
	if scene.LeafCache != nil && scene.LeafCache.ClusterId == currentLeaf.Cluster {
		return
	}

	scene.currentLeaf = currentLeaf
	scene.LeafCache = scene.visData.GetPVSCacheForCluster(currentLeaf.Cluster)

	scene.visibleClusterLeafs = scene.asyncRebuildVisibleWorld(scene.currentLeaf)
}

// Launches rebuilding the visible world in a separate thread
// Note: This *could* cause rendering issues if the rebuild is slower than
// travelling between clusters
func (scene *StaticScene) asyncRebuildVisibleWorld(currentLeaf *leaf.Leaf) []*vis.ClusterLeaf {
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

	return visibleWorld
}

func NewStaticSceneFromBsp(fs fileSystem,
	level *graphics.Bsp,
	entities []entity.IEntity,
	materialCache *cache.Material,
	texCache *cache.Texture,
	gpuItemCache *cache.GpuItem,
	gpuStaticProps map[string]*cache.GpuProp) *StaticScene {

	if level.LightmapAtlas() != nil {
		texCache.Add(cache.LightmapTexturePath, level.LightmapAtlas())
		gpuItemCache.Add(cache.LightmapTexturePath, adapter.UploadLightmap(texCache.Find(cache.LightmapTexturePath)))
	} else {
		texCache.Add(cache.LightmapTexturePath, texCache.Find(cache.ErrorTexturePath))
		gpuItemCache.Add(cache.LightmapTexturePath, gpuItemCache.Find(cache.ErrorTexturePath))
	}

	// load materials
	var tex graphics.Texture
	var err error
	for _, mat := range level.MaterialDictionary() {
		if tex = texCache.Find(mat.BaseTextureName); tex == nil {
			if mat.BaseTextureName == "" {
				console.PrintString(console.LevelWarning, fmt.Sprintf("%s has no $BaseTexture", mat.FilePath()))
				texCache.Add(mat.BaseTextureName, texCache.Find(cache.ErrorTexturePath))
				gpuItemCache.Add(mat.BaseTextureName, gpuItemCache.Find(cache.ErrorTexturePath))
			} else {
				tex, err = graphics.LoadTexture(fs, mat.BaseTextureName)
				if err != nil || tex == nil {
					if err != nil {
						console.PrintString(console.LevelWarning, err.Error())
					}
					texCache.Add(mat.BaseTextureName, texCache.Find(cache.ErrorTexturePath))
					gpuItemCache.Add(mat.BaseTextureName, gpuItemCache.Find(cache.ErrorTexturePath))
				} else {
					texCache.Add(mat.BaseTextureName, tex)
					gpuItemCache.Add(mat.BaseTextureName, adapter.UploadTexture(tex))
				}
			}
		}
		materialCache.Add(strings.ToLower(mat.FilePath()), cache.NewGpuMaterial(gpuItemCache.Find(mat.BaseTextureName), mat))
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
			tex = texCache.Find(cache.ErrorTexturePath)
		} else {
			if level.MaterialDictionary()[bspFace.Material()].BaseTextureName == "" {
				tex = texCache.Find(cache.ErrorTexturePath)
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

	// Finish staticprops
	for _, prop := range level.StaticPropDictionary {
		gpuStaticProps[prop.Id] = cache.NewGpuProp()
		for _, m := range prop.Meshes() {
			gpuStaticProps[prop.Id].AddMesh(adapter.UploadMesh(m))
		}
		for _, materialPath := range prop.Materials() {
			if _, ok := level.MaterialDictionary()[materialPath]; ok {
				gpuStaticProps[prop.Id].AddMaterial(*materialCache.Find(strings.ToLower(materialPath)))
				continue
			}
			mat, err := graphics.LoadMaterial(fs, materialPath)
			if err != nil {
				mat = graphics.NewMaterial(materialPath)
				mat.BaseTextureName = cache.ErrorTexturePath
			}
			if tex := texCache.Find(mat.BaseTextureName); tex == nil {
				tex, err := graphics.LoadTexture(fs, mat.BaseTextureName)
				if err != nil {
					console.PrintString(console.LevelWarning, err.Error())
					texCache.Add(mat.BaseTextureName, texCache.Find(cache.ErrorTexturePath))
					gpuItemCache.Add(mat.BaseTextureName, gpuItemCache.Find(cache.ErrorTexturePath))
				} else {
					texCache.Add(mat.BaseTextureName, tex)
					gpuItemCache.Add(mat.BaseTextureName, adapter.UploadTexture(tex))
				}
			}
			materialCache.Add(strings.ToLower(mat.FilePath()), cache.NewGpuMaterial(gpuItemCache.Find(mat.BaseTextureName), mat))
			gpuStaticProps[prop.Id].AddMaterial(*materialCache.Find(strings.ToLower(materialPath)))
		}
	}

	// Generate visibility tree
	visibility := vis.LoadVisData(level.File())
	clusterLeafs := generateClusterLeafs(level, visibility)

	var worldspawn entity.IEntity
	var skyCameraEntity entity.IEntity
	var infoPlayerStart entity.IEntity
	for idx, e := range entities {
		if e.Classname() == "worldspawn" {
			worldspawn = entities[idx]
			continue
		}
		if e.Classname() == "sky_camera" {
			skyCameraEntity = entities[idx]
			continue
		}
		if e.Classname() == "info_player_start" {
			infoPlayerStart = entities[idx]
			continue
		}
	}
	skyboxOrigin := mgl32.Vec3{}
	skyName := ""
	if worldspawn != nil {
		skyboxOrigin = worldspawn.VectorForKey("origin")
		skyName = worldspawn.ValueForKey("skyname")
	}
	skybox := scene.LoadSkybox(fs, skyName, skyboxOrigin)
	var skyCamera *graphics.Camera

	if skyCameraEntity != nil {
		skyCamera = graphics.NewCamera(level.Camera().Fov(), level.Camera().AspectRatio())
		skyCamera.Transform().Position = skyCameraEntity.VectorForKey("origin")
		scale := skyCameraEntity.FloatForKey("scale")
		skyCamera.Transform().Scale = mgl32.Vec3{scale, scale, scale}
	}
	if infoPlayerStart != nil {
		level.Camera().Transform().Position = infoPlayerStart.VectorForKey("origin")
		level.Camera().Transform().Rotation = infoPlayerStart.VectorForKey("angles")
	}

	scene := &StaticScene{
		bspMesh:            level.Mesh(),
		gpuMesh:            adapter.UploadMesh(level.Mesh()),
		bspFaces:           remappedFaces,
		displacementFaces:  dispFaces,
		skybox:             skybox,
		entities:           entities,
		staticProps:        level.StaticProps,
		clusterLeafs:       clusterLeafs,
		visData:            visibility,
		camera:             level.Camera(),
		cameraPrevPosition: mgl32.Vec3{99999, 99999, 99999},
		skyCamera:          skyCamera,
	}

	// Generate Initial visibility data
	scene.RecomputeVisibleClusters()
	if skyCamera != nil {
		scene.skyboxClusterLeafs = scene.asyncRebuildVisibleWorld(scene.visData.FindCurrentLeaf(skyCamera.Transform().Position))
	}

	return scene
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
