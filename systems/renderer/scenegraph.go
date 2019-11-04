package renderer

import (
	"fmt"
	"github.com/galaco/bsp/primitives/leaf"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/framework/graphics/adapter"
	"github.com/galaco/kero/messages"
	"github.com/galaco/kero/systems/renderer/cache"
	"github.com/galaco/kero/systems/renderer/deferred"
	"github.com/galaco/kero/systems/renderer/scene"
	"github.com/galaco/kero/systems/renderer/vis"
	"github.com/galaco/kero/valve"
	"github.com/go-gl/mathgl/mgl32"
	"io"
	"strings"
)

type fileSystem interface {
	GetFile(string) (io.Reader, error)
}

type SceneGraph struct {
	bspMesh           *graphics.BasicMesh
	bspFaces          []valve.BspFace
	displacementFaces []*valve.BspFace
	skybox            *scene.Skybox

	gpuMesh          adapter.GpuMesh
	staticProps      []graphics.StaticProp
	entities         []entity.Entity
	lightEnvironment *deferred.DirectionalLight
	pointLights      []entity.Entity

	visData      *vis.Vis
	clusterLeafs []vis.ClusterLeaf
	LeafCache    *vis.Cluster

	visibleClusterLeafs []*vis.ClusterLeaf
	currentLeaf         *leaf.Leaf

	cameraPrevPosition mgl32.Vec3
}

// RecomputeVisibleClusters rebuilds the current facelist to render, by first
// recalculating using vvis data
func (scene *SceneGraph) RecomputeVisibleClusters(camera *graphics3d.Camera) {
	if camera.Transform().Position.ApproxEqual(scene.cameraPrevPosition) {
		return
	}
	scene.cameraPrevPosition = camera.Transform().Position
	// View hasn't moved
	currentLeaf := scene.visData.FindCurrentLeaf(camera.Transform().Position)

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

func NewSceneGraphFromBsp(fs fileSystem,
	level *valve.Bsp,
	entities []entity.Entity,
	materialCache *cache.Material,
	texCache *cache.Texture,
	gpuItemCache *cache.GpuItem,
	gpuStaticProps map[string]*cache.GpuProp) *SceneGraph {
	texCache.Add(cache.ErrorTexturePath, graphics.NewErrorTexture(cache.ErrorTexturePath))
	gpuItemCache.Add(cache.ErrorTexturePath, adapter.UploadTexture(texCache.Find(cache.ErrorTexturePath)))

	// load materials
	var albedoTexture, normalTexture *graphics.Texture
	var err error
	for _, mat := range level.MaterialDictionary() {
		if albedoTexture = texCache.Find(mat.BaseTextureName); albedoTexture == nil {
			albedoTexture, err = graphics.LoadTexture(fs, mat.BaseTextureName)
			if err != nil || albedoTexture == nil {
				event.Dispatch(messages.NewConsoleMessage(console.LevelWarning, err.Error()))
				texCache.Add(mat.BaseTextureName, texCache.Find(cache.ErrorTexturePath))
				gpuItemCache.Add(mat.BaseTextureName, gpuItemCache.Find(cache.ErrorTexturePath))
			} else {
				texCache.Add(mat.BaseTextureName, albedoTexture)
				gpuItemCache.Add(mat.BaseTextureName, adapter.UploadTexture(albedoTexture))
			}
		}
		if mat.HasBumpMap {
			if normalTexture = texCache.Find(mat.BumpMapName); normalTexture == nil {
				normalTexture, err = graphics.LoadTexture(fs, mat.BumpMapName)
				if err != nil || normalTexture == nil {
					event.Dispatch(messages.NewConsoleMessage(console.LevelWarning, err.Error()))
					texCache.Add(mat.BumpMapName, texCache.Find(cache.ErrorTexturePath))
					gpuItemCache.Add(mat.BumpMapName, gpuItemCache.Find(cache.ErrorTexturePath))
				} else {
					texCache.Add(mat.BumpMapName, normalTexture)
					gpuItemCache.Add(mat.BumpMapName, adapter.UploadTexture(normalTexture))
				}
			}
		}
		gpuMat := cache.NewGpuMaterial(gpuItemCache.Find(mat.BaseTextureName), mat)
		if normalTexture != nil {
			gpuMat.AddNormalMap(gpuItemCache.Find(mat.BumpMapName))
		}

		materialCache.Add(strings.ToLower(mat.FilePath()), gpuMat)
	}

	// generate displacement faces
	dispFaces := make([]*valve.BspFace, 0, 1024)
	for _, i := range level.DispFaces() {
		dispFaces = append(dispFaces, &level.Faces()[i])
	}

	// finish bsp mesh
	// Add MATERIALS TO FACES
	albedoTexture = nil
	for _, bspFace := range level.Faces() {
		if level.MaterialDictionary()[bspFace.Material()] == nil {
			event.Dispatch(messages.NewConsoleMessage(console.LevelWarning, fmt.Sprintf("MATERIAL: %s not found", bspFace.Material())))
			albedoTexture = texCache.Find(cache.ErrorTexturePath)
		} else {
			if level.MaterialDictionary()[bspFace.Material()].BaseTextureName == "" {
				albedoTexture = texCache.Find(cache.ErrorTexturePath)
			} else {
				albedoTexture = texCache.Find(level.MaterialDictionary()[bspFace.Material()].BaseTextureName)
			}
		}
		// Generate texture coordinates
		level.Mesh().AddUV(
			valve.TexCoordsForFaceFromTexInfo(
				level.Mesh().Vertices()[bspFace.Offset()*3:(bspFace.Offset()*3)+(bspFace.Length()*3)],
				bspFace.TexInfo(),
				albedoTexture.Width(),
				albedoTexture.Height())...)
	}

	level.Mesh().GenerateTangents()

	remappedFaces := make([]valve.BspFace, 0, 1024)
	// Kero isnt interested in tools faces (for now)
	for idx := range level.Faces() {
		//if strings.HasPrefix(strings.ToLower(bspFace.Material()), "tools") {
		//	continue
		//}
		remappedFaces = append(remappedFaces, level.Faces()[idx])
	}

	// Finish staticprops
	for _, prop := range level.StaticPropDictionary {
		gpuStaticProps[prop.Id] = cache.NewGpuProp()
		for _, m := range prop.Meshes() {
			gpuMesh := adapter.UploadMesh(m)
			gpuStaticProps[prop.Id].AddMesh(&gpuMesh)
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
					event.Dispatch(messages.NewConsoleMessage(console.LevelWarning, err.Error()))
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

	var worldspawn entity.Entity
	var pointlights []entity.Entity
	var lightEnvironment *deferred.DirectionalLight
	for idx, e := range entities {
		if e.Classname() == "worldspawn" {
			worldspawn = entities[idx]
			continue
		}
		if e.Classname() == "light" {
			pointlights = append(pointlights, e)
		}
		if e.Classname() == "light_environment" && lightEnvironment == nil {
			// _ambient = vec3 color, float intensity
			// _light = vec3 color, float intensity
			// angles = vec3
			// pitch -60
			// origin vec3
			lightEnvironment = &deferred.DirectionalLight{
				BaseLight: deferred.BaseLight{},
				Direction: e.VectorForKey("angles"),
			}
			lightEnvironment.Direction[0] = e.FloatForKey("pitch")
			lightEnvironment.Direction[2] = lightEnvironment.Direction[0]
			_, _ = fmt.Sscanf(
				e.ValueForKey("_ambient"),
				"%f %f %f %f",
				&lightEnvironment.AmbientColor[0],
				&lightEnvironment.AmbientColor[1],
				&lightEnvironment.AmbientColor[2],
				&lightEnvironment.AmbientIntensity)
			_, _ = fmt.Sscanf(
				e.ValueForKey("_light"),
				"%f %f %f %f",
				&lightEnvironment.Color[0],
				&lightEnvironment.Color[1],
				&lightEnvironment.Color[2],
				&lightEnvironment.DiffuseIntensity)
			lightEnvironment.AmbientColor[0] /= 255
			lightEnvironment.AmbientColor[1] /= 255
			lightEnvironment.AmbientColor[2] /= 255
			lightEnvironment.AmbientIntensity /= 255
			lightEnvironment.Color[0] /= 255
			lightEnvironment.Color[1] /= 255
			lightEnvironment.Color[2] /= 255
			lightEnvironment.DiffuseIntensity /= 255
		}
	}
	skybox := scene.LoadSkybox(fs, worldspawn)

	return &SceneGraph{
		bspMesh:            level.Mesh(),
		gpuMesh:            adapter.UploadMesh(level.Mesh()),
		bspFaces:           remappedFaces,
		displacementFaces:  dispFaces,
		skybox:             skybox,
		pointLights:        pointlights,
		entities:           entities,
		staticProps:        level.StaticProps,
		clusterLeafs:       clusterLeafs,
		visData:            visibility,
		cameraPrevPosition: mgl32.Vec3{65536, 65536, 65536},
		lightEnvironment:   lightEnvironment,
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

	return bspClusters
}
