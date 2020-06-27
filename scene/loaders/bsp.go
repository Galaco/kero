package loader

import (
	"fmt"
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/common"
	"github.com/galaco/bsp/primitives/dispinfo"
	"github.com/galaco/bsp/primitives/dispvert"
	"github.com/galaco/bsp/primitives/face"
	"github.com/galaco/bsp/primitives/plane"
	"github.com/galaco/bsp/primitives/texinfo"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/messages"
	"github.com/galaco/vtf/format"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang-source-engine/stringtable"
	"math"
	"strings"
	"sync"
)

// LoadBspMap is the gateway into loading the core static level. Entities are loaded
// elsewhere
// It loads in the following order:
// BSP Geometry
// BSP Materials
// StaticProps (materials loaded as required)
func LoadBspMap(fs filesystem.FileSystem, filename string) (*graphics.Bsp, []entity.IEntity, error) {
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateStarted))
	file, err := bsp.ReadFromFile(filename)
	if err != nil {
		event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateError))
		return nil, nil, err
	}
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateBSPParsed))
	fs.RegisterPakFile(file.Lump(bsp.LumpPakfile).(*lumps.Pakfile))
	// Load the static bsp world
	level, err := loadBSPWorld(fs, file)

	if err != nil {
		event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateError))
		return nil, nil, err
	}
	level.SetCamera(graphics3d.NewCamera(
		mgl32.DegToRad(70),
		float32(window.CurrentWindow().Width())/float32(window.CurrentWindow().Height())))
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateGeometryLoaded))

	// Load staticprops
	level.StaticPropDictionary, level.StaticProps = LoadStaticProps(fs, file)
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateStaticPropsLoaded))

	// Load entities
	ents, err := entity.LoadEntdata(fs, file)
	if err != nil {
		return nil, nil, err
	}
	event.Get().DispatchLegacy(messages.NewLoadingLevelProgress(messages.LoadingProgressStateEntitiesLoaded))

	return level, ents, err
}

type bspstructs struct {
	faces     []face.Face
	planes    []plane.Plane
	vertexes  []mgl32.Vec3
	surfEdges []int32
	edges     [][2]uint16
	texInfos  []texinfo.TexInfo
	dispInfos []dispinfo.DispInfo
	dispVerts []dispvert.DispVert
	lightmap  []common.ColorRGBExponent32
}

// LoadBspMap is the gateway into loading the core static level. Entities are loaded
// elsewhere
// It loads in the following order:
// BSP Geometry
// BSP Materials
// StaticProps (materials loaded as required)
func loadBSPWorld(fs filesystem.FileSystem, file *bsp.Bsp) (*graphics.Bsp, error) {
	bspStructure := bspstructs{
		faces:     file.Lump(bsp.LumpFaces).(*lumps.Face).GetData(),
		planes:    file.Lump(bsp.LumpPlanes).(*lumps.Planes).GetData(),
		vertexes:  file.Lump(bsp.LumpVertexes).(*lumps.Vertex).GetData(),
		surfEdges: file.Lump(bsp.LumpSurfEdges).(*lumps.Surfedge).GetData(),
		edges:     file.Lump(bsp.LumpEdges).(*lumps.Edge).GetData(),
		texInfos:  file.Lump(bsp.LumpTexInfo).(*lumps.TexInfo).GetData(),
		dispInfos: file.Lump(bsp.LumpDispInfo).(*lumps.DispInfo).GetData(),
		dispVerts: file.Lump(bsp.LumpDispVerts).(*lumps.DispVert).GetData(),
		lightmap:  file.Lump(bsp.LumpLighting).(*lumps.Lighting).GetData(),
	}

	//MATERIALS
	stringTable := stringtable.NewFromExistingStringTableData(
		file.Lump(bsp.LumpTexDataStringData).(*lumps.TexDataStringData).GetData(),
		file.Lump(bsp.LumpTexDataStringTable).(*lumps.TexDataStringTable).GetData())
	materials := buildUniqueMaterialList(stringTable, &bspStructure.texInfos)

	materialDictionary := buildMaterialDictionary(fs, materials)

	// BSP FACES
	bspMesh := graphics.NewMesh()
	bspFaces := make([]graphics.BspFace, len(bspStructure.faces))
	// storeDispFaces until for visibility calculation purposes.
	dispFaces := make([]int, 0)

	var lightmapAtlas *graphics.TextureAtlas
	if bspStructure.lightmap != nil {
		lightmapAtlas = generateLightmapTexture(bspStructure.faces, bspStructure.lightmap)
	}

	for idx, f := range bspStructure.faces {
		if f.DispInfo > -1 {
			// This face is a displacement
			bspFaces[idx] = generateDisplacementFace(&bspStructure.faces[idx], &bspStructure, bspMesh)
			dispFaces = append(dispFaces, idx)
		} else {
			bspFaces[idx] = generateBspFace(&bspStructure.faces[idx], &bspStructure, bspMesh)
		}

		faceVmt, err := stringTable.FindString(int(bspStructure.texInfos[bspStructure.faces[idx].TexInfo].TexData))
		if err != nil {
			console.PrintInterface(console.LevelError, err)
		} else {
			bspFaces[idx].SetMaterial(strings.ToLower(faceVmt))
		}
	}

	return graphics.NewBsp(file, bspMesh, bspFaces, dispFaces, materialDictionary, bspStructure.texInfos, lightmapAtlas), nil
}

// SortUnique builds a unique list of materials in a StringTable
// referenced by BSP TexInfo lump data.
func buildUniqueMaterialList(stringTable *stringtable.StringTable, texInfos *[]texinfo.TexInfo) []string {
	materialList := make([]string, 0)
	for _, ti := range *texInfos {
		target, _ := stringTable.FindString(int(ti.TexData))
		found := false
		for _, cur := range materialList {
			if cur == target {
				found = true
				break
			}
		}
		if !found {
			materialList = append(materialList, target)
		}
	}

	return materialList
}

func buildMaterialDictionary(fs filesystem.FileSystem, materials []string) (dictionary map[string]*graphics.Material) {
	dictionary = map[string]*graphics.Material{}
	waitGroup := sync.WaitGroup{}
	dictMutex := sync.Mutex{}

	asyncLoadMaterial := func(filePath string) {
		mat, err := graphics.LoadMaterial(fs, filePath)
		if err != nil {
			console.PrintString(console.LevelError, fmt.Sprintf("%s", err))
			mat = graphics.NewMaterial(filePath)
		}
		dictMutex.Lock()
		dictionary[strings.ToLower(filePath)] = mat
		dictMutex.Unlock()
		waitGroup.Done()
	}

	waitGroup.Add(len(materials))
	for _, filePath := range materials {
		go asyncLoadMaterial(filePath)
	}
	waitGroup.Wait()

	return dictionary
}

// generateBspFace Create primitives from face data in the bsp
func generateBspFace(f *face.Face, bspStructure *bspstructs, bspMesh *graphics.BasicMesh) graphics.BspFace {
	offset := int32(len(bspMesh.Vertices())) / 3
	length := int32(0)

	planeNormal := bspStructure.planes[f.Planenum].Normal
	// All surfedges associated with this face
	// surfEdges are basically indices into the edges lump
	faceSurfEdges := bspStructure.surfEdges[f.FirstEdge:(f.FirstEdge + int32(f.NumEdges))]
	rootIndex := uint16(0)
	for idx, surfEdge := range faceSurfEdges {
		edge := bspStructure.edges[int(math.Abs(float64(surfEdge)))]
		e1 := 0
		e2 := 1
		if surfEdge < 0 {
			e1 = 1
			e2 = 0
		}
		//Capture root indice
		if idx == 0 {
			rootIndex = edge[e1]
		} else {
			// Just create a triangle for every edge now
			bspMesh.AddIndice(uint32(len(bspMesh.Vertices())) / 3)
			bspMesh.AddVertex(bspStructure.vertexes[rootIndex].X(), bspStructure.vertexes[rootIndex].Y(), bspStructure.vertexes[rootIndex].Z())
			bspMesh.AddNormal(planeNormal.X(), planeNormal.Y(), planeNormal.Z())

			bspMesh.AddIndice(uint32(len(bspMesh.Vertices())) / 3)
			bspMesh.AddVertex(bspStructure.vertexes[edge[e1]].X(), bspStructure.vertexes[edge[e1]].Y(), bspStructure.vertexes[edge[e1]].Z())
			bspMesh.AddNormal(planeNormal.X(), planeNormal.Y(), planeNormal.Z())

			bspMesh.AddIndice(uint32(len(bspMesh.Vertices())) / 3)
			bspMesh.AddVertex(bspStructure.vertexes[edge[e2]].X(), bspStructure.vertexes[edge[e2]].Y(), bspStructure.vertexes[edge[e2]].Z())
			bspMesh.AddNormal(planeNormal.X(), planeNormal.Y(), planeNormal.Z())

			length += 3 // num verts (3 b/c face triangles)
		}
	}

	return graphics.NewMeshFace(offset, length, &bspStructure.texInfos[f.TexInfo], f)
}

// generateDisplacementFace Create Primitive from Displacement face
// This is based on:
// https://github.com/Metapyziks/VBspViewer/blob/master/Assets/VBspViewer/Scripts/Importing/VBsp/VBspFile.cs
func generateDisplacementFace(f *face.Face, bspStructure *bspstructs, bspMesh *graphics.BasicMesh) graphics.BspFace {
	corners := make([]mgl32.Vec3, 4)
	normal := bspStructure.planes[f.Planenum].Normal

	info := bspStructure.dispInfos[f.DispInfo]
	size := int(1 << uint32(info.Power))
	firstCorner := int32(0)
	firstCornerDist2 := float32(math.MaxFloat32)

	offset := int32(len(bspMesh.Vertices())) / 3
	length := int32(0)

	for surfId := f.FirstEdge; surfId < f.FirstEdge+int32(f.NumEdges); surfId++ {
		surfEdge := bspStructure.surfEdges[surfId]
		edgeIndex := int32(math.Abs(float64(surfEdge)))
		edge := bspStructure.edges[edgeIndex]
		vert := bspStructure.vertexes[edge[0]]
		if surfEdge < 0 {
			vert = bspStructure.vertexes[edge[1]]
		}
		corners[surfId-f.FirstEdge] = vert

		dist2tmp := info.StartPosition.Sub(vert)
		dist2 := (dist2tmp.X() * dist2tmp.X()) + (dist2tmp.Y() * dist2tmp.Y()) + (dist2tmp.Z() * dist2tmp.Z())
		if dist2 < firstCornerDist2 {
			firstCorner = surfId - f.FirstEdge
			firstCornerDist2 = dist2
		}
	}

	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			a := generateDispVert(int(info.DispVertStart), x, y, size, corners, firstCorner, &bspStructure.dispVerts)
			b := generateDispVert(int(info.DispVertStart), x, y+1, size, corners, firstCorner, &bspStructure.dispVerts)
			c := generateDispVert(int(info.DispVertStart), x+1, y+1, size, corners, firstCorner, &bspStructure.dispVerts)
			d := generateDispVert(int(info.DispVertStart), x+1, y, size, corners, firstCorner, &bspStructure.dispVerts)

			// Split into triangles
			bspMesh.AddIndice(uint32(len(bspMesh.Vertices()))/3, (uint32(len(bspMesh.Vertices()))/3)+1, (uint32(len(bspMesh.Vertices()))/3)+2)
			bspMesh.AddVertex(a.X(), a.Y(), a.Z(), b.X(), b.Y(), b.Z(), c.X(), c.Y(), c.Z())
			bspMesh.AddNormal(normal.X(), normal.Y(), normal.Z(), normal.X(), normal.Y(), normal.Z(), normal.X(), normal.Y(), normal.Z())
			bspMesh.AddIndice(uint32(len(bspMesh.Vertices()))/3, (uint32(len(bspMesh.Vertices()))/3)+1, (uint32(len(bspMesh.Vertices()))/3)+2)
			bspMesh.AddVertex(a.X(), a.Y(), a.Z(), c.X(), c.Y(), c.Z(), d.X(), d.Y(), d.Z())
			bspMesh.AddNormal(normal.X(), normal.Y(), normal.Z(), normal.X(), normal.Y(), normal.Z(), normal.X(), normal.Y(), normal.Z())

			length += 6 // 6 b/c quad = 2*triangle
		}
	}

	return graphics.NewMeshFace(offset, length, &bspStructure.texInfos[f.TexInfo], f)
}

// generateDispVert Create a displacement vertex
func generateDispVert(offset int, x int, y int, size int, corners []mgl32.Vec3, firstCorner int32, dispVerts *[]dispvert.DispVert) mgl32.Vec3 {
	vert := (*dispVerts)[offset+x+y*(size+1)]

	tx := float32(x) / float32(size)
	ty := float32(y) / float32(size)
	sx := 1.0 - tx
	sy := 1.0 - ty

	cornerA := corners[(0+firstCorner)&3]
	cornerB := corners[(1+firstCorner)&3]
	cornerC := corners[(2+firstCorner)&3]
	cornerD := corners[(3+firstCorner)&3]

	origin := ((cornerB.Mul(sx).Add(cornerC.Mul(tx))).Mul(ty)).Add((cornerA.Mul(sx).Add(cornerD.Mul(tx))).Mul(sy))

	return origin.Add(vert.Vec.Mul(vert.Dist))
}

func generateLightmapTexture(faces []face.Face, samples []common.ColorRGBExponent32) *graphics.TextureAtlas {
	lightMapAtlas := graphics.NewTextureAtlas(0, 0)
	textures := make([]*graphics.Texture2D, len(faces))

	for idx, f := range faces {
		textures[idx] = lightmapTextureFromFace(&f, samples)
		lightMapAtlas.AddRaw(textures[idx].Width(), textures[idx].Height(), textures[idx].Image())
	}

	lightMapAtlas.Pack()

	return lightMapAtlas
}

func lightmapTextureFromFace(f *face.Face, samples []common.ColorRGBExponent32) *graphics.Texture2D {
	if f.Lightofs == -1 {
		return graphics.NewTexture("__lightmap_subtex__", 0, 0, uint32(format.RGB888), make([]uint8, 0))
	}

	width := f.LightmapTextureSizeInLuxels[0] + 1
	height := f.LightmapTextureSizeInLuxels[1] + 1
	numLuxels := width * height
	firstSampleIdx := f.Lightofs / 4 // 4 = size of ColorRGBExponent32

	raw := make([]uint8, (numLuxels)*4)

	for idx, sample := range samples[firstSampleIdx : firstSampleIdx+numLuxels] {
		raw[(idx * 4)] = uint8(math.Min(255, float64(sample.R)*math.Pow(2, float64(sample.Exponent))))
		raw[(idx*4)+1] = uint8(math.Min(255, float64(sample.G)*math.Pow(2, float64(sample.Exponent))))
		raw[(idx*4)+2] = uint8(math.Min(255, float64(sample.B)*math.Pow(2, float64(sample.Exponent))))
		raw[(idx*4)+3] = 255
	}

	return graphics.NewTexture("__lightmap_subtex__", int(width), int(height), uint32(format.RGBA8888), raw)
}

