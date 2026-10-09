package scene

import (
	"fmt"
	"github.com/galaco/bsp/lump/primitive/leaf"
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

// Deprecated: Use Engine.SceneManager().GetCurrentScene() instead.
// This function will be removed in a future version.
// For new code, use SceneManager passed as an explicit dependency.
func CurrentScene() *StaticScene {
	if sceneSingleton.BspMesh == nil {
		return nil
	}

	return &sceneSingleton
}

// Deprecated: Use Engine.SceneManager().CloseCurrentScene() instead.
// This function will be removed in a future version.
func CloseCurrentScene() {
	sceneSingleton = StaticScene{}
}

type StaticScene struct {
	RawBsp               *graphics.Bsp
	BspMesh              *mesh.BasicMesh // Regular BSP geometry (no blend weights)
	DisplacementBspMesh  *mesh.BasicMesh // Displacement surfaces (with blend weights)
	BspFaces             []graphics.BspFace
	DisplacementFaces    []*graphics.BspFace
	Textures             map[string]graphics.Texture

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
		scene.LeafCache = nil
		scene.VisibleClusterLeafs = scene.asyncRebuildVisibleWorld(currentLeaf)
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

	// Create a set of displacement face indices for quick lookup
	dispFaceSet := make(map[int]bool)
	for _, idx := range level.DispFaces() {
		dispFaceSet[idx] = true
	}

	// Add MATERIALS TO REGULAR BSP FACES (non-displacements)
	tex = nil
	for idx, bspFace := range level.Faces() {
		// Skip displacement faces - they'll be processed separately
		if dispFaceSet[idx] {
			continue
		}

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
		// Generate texture coordinates for regular BSP mesh
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

	// Add MATERIALS TO DISPLACEMENT FACES
	for idx, bspFace := range level.Faces() {
		// Only process displacement faces
		if !dispFaceSet[idx] {
			continue
		}

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
		// Texture coordinates are mapped from the face the displacement is built from, not the displaced vertices
		basePositions := make([]float32, 0, len(bspFace.DisplacementVertices())*3)
		for _, v := range bspFace.DisplacementVertices() {
			basePositions = append(basePositions, v.Base.X(), v.Base.Y(), v.Base.Z())
		}
		level.DisplacementMesh().AddUV(
			graphics.TexCoordsForFaceFromTexInfo(
				basePositions,
				bspFace.TexInfo(),
				tex.Width(),
				tex.Height())...)

		if level.LightmapAtlas() != nil {
			level.DisplacementMesh().AddLightmapUV(
				graphics.DisplacementLightmapCoords(
					bspFace.DisplacementVertices(),
					bspFace.RawFace(),
					bspFace.TexInfo(),
					float32(level.LightmapAtlas().Width()),
					float32(level.LightmapAtlas().Height()),
					level.LightmapAtlas().AtlasEntry(idx).X,
					level.LightmapAtlas().AtlasEntry(idx).Y)...)
		}
	}

	level.DisplacementMesh().GenerateTangents()

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
		RawBsp:              level,
		BspMesh:             level.Mesh(),
		DisplacementBspMesh: level.DisplacementMesh(),
		BspFaces:            remappedFaces,
		DisplacementFaces:   dispFaces,
		Entities:            entities,
		StaticProps:         level.StaticProps,
		ClusterLeafs:        clusterLeafs,
		VisData:             visibility,
		Camera:              level.Camera(),
		CameraPrevPosition:  mgl32.Vec3{99999, 99999, 99999},
		SkyCamera:           skyCamera,
		TexCache:            texCache,
	}

	// Generate Initial visibility data based on camera position
	initialLeaf := visibility.FindCurrentLeaf(level.Camera().Transform().Translation)
	sceneSingleton.CurrentLeaf = initialLeaf
	if initialLeaf != nil && initialLeaf.Cluster != -1 {
		sceneSingleton.LeafCache = visibility.GetPVSCacheForCluster(initialLeaf.Cluster)
	}
	sceneSingleton.VisibleClusterLeafs = sceneSingleton.asyncRebuildVisibleWorld(initialLeaf)

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

	// A cluster's bounds hold every one of its leafs, and every face in them; a face is not split where it crosses
	// from one leaf into another
	hasBounds := make([]bool, len(bspClusters))
	grow := func(cluster int16, mins, maxs mgl32.Vec3) {
		if !hasBounds[cluster] {
			bspClusters[cluster].Mins, bspClusters[cluster].Maxs = mins, maxs
			hasBounds[cluster] = true
			return
		}
		for axis := 0; axis < 3; axis++ {
			bspClusters[cluster].Mins[axis] = min(bspClusters[cluster].Mins[axis], mins[axis])
			bspClusters[cluster].Maxs[axis] = max(bspClusters[cluster].Maxs[axis], maxs[axis])
		}
	}
	vertices := level.Mesh().Vertices()

	for _, bspLeaf := range visData.Leafs {
		if bspLeaf.Cluster == -1 {
			//defaultCluster.Faces = append(defaultCluster.Faces, bspFaces[leafFace])
			continue
		}
		bspClusters[bspLeaf.Cluster].Id = bspLeaf.Cluster
		grow(bspLeaf.Cluster,
			mgl32.Vec3{float32(bspLeaf.Mins[0]), float32(bspLeaf.Mins[1]), float32(bspLeaf.Mins[2])},
			mgl32.Vec3{float32(bspLeaf.Maxs[0]), float32(bspLeaf.Maxs[1]), float32(bspLeaf.Maxs[2])})
		if bspLeaf.Flags()&leaf.LeafFlagsSky > 0 {
			bspClusters[bspLeaf.Cluster].SkyVisible = true
		}

		for _, leafFace := range visData.LeafFaces[bspLeaf.FirstLeafFace : bspLeaf.FirstLeafFace+bspLeaf.NumLeafFaces] {
			face := level.Faces()[leafFace]
			bspClusters[bspLeaf.Cluster].Faces = append(bspClusters[bspLeaf.Cluster].Faces, face)
			for v := face.Offset(); v < face.Offset()+face.Length(); v++ {
				vertex := mgl32.Vec3{vertices[v*3], vertices[v*3+1], vertices[v*3+2]}
				grow(bspLeaf.Cluster, vertex, vertex)
			}
		}
	}

	// Displacements are not in leafs' face lists. A displacement belongs to every cluster with a leaf it overlaps.
	for dispIdx, faceIdx := range level.DispFaces() {
		face := level.Faces()[faceIdx]
		if face.Length() == 0 {
			continue
		}
		mins, maxs := face.Bounds()

		for _, bspLeaf := range visData.Leafs {
			if bspLeaf.Cluster == -1 {
				continue
			}
			overlaps := true
			for axis := 0; axis < 3; axis++ {
				if float32(bspLeaf.Maxs[axis]) < mins[axis] || float32(bspLeaf.Mins[axis]) > maxs[axis] {
					overlaps = false
					break
				}
			}
			if !overlaps {
				continue
			}
			cluster := &bspClusters[bspLeaf.Cluster]
			if n := len(cluster.DispFaces); n > 0 && cluster.DispFaces[n-1] == dispIdx {
				continue
			}
			cluster.DispFaces = append(cluster.DispFaces, dispIdx)
			grow(bspLeaf.Cluster, mins, maxs)
		}
	}

	for idx := range bspClusters {
		cluster := &bspClusters[idx]
		cluster.Origin = cluster.Mins.Add(cluster.Maxs).Mul(0.5)
		cluster.DebugMesh = mesh.NewCuboidFromMinMaxs(cluster.Mins, cluster.Maxs)
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
