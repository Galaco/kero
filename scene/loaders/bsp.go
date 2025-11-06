package loader

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lump"
	"github.com/galaco/bsp/lump/primitive/common"
	"github.com/galaco/bsp/lump/primitive/dispinfo"
	"github.com/galaco/bsp/lump/primitive/dispvert"
	"github.com/galaco/bsp/lump/primitive/face"
	"github.com/galaco/bsp/lump/primitive/plane"
	"github.com/galaco/bsp/lump/primitive/texinfo"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/event"
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/galaco/kero/framework/window"
	"github.com/galaco/kero/messages"
	"github.com/galaco/stringtable"
	"github.com/galaco/vtf/format"
	"github.com/go-gl/mathgl/mgl32"
)

// LoadBspMap is the gateway into loading the core static level. Entities are loaded
// elsewhere
// It loads in the following order:
// BSP Geometry
// BSP Materials
// StaticProps (materials loaded as required)
func LoadBspMap(fs filesystem.FileSystem, eventBus *event.Dispatcher, filename string) (*graphics.Bsp, []entity.IEntity, error) {
	return LoadBspMapWithContext(context.Background(), fs, eventBus, filename)
}

// LoadBspMapWithContext loads a BSP map with cancellation support via context
// This is the async-safe version that checks for cancellation at key points
// Supports both absolute paths (from file dialog) and relative paths (from console/filesystem)
func LoadBspMapWithContext(ctx context.Context, fs filesystem.FileSystem, eventBus *event.Dispatcher, filename string) (*graphics.Bsp, []entity.IEntity, error) {
	// Use typed event dispatch (Phase 3)
	event.DispatchTyped(eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateStarted})

	// Check cancellation before starting
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	var reader io.Reader
	var err error

	// Check if path is absolute (from file dialog) or relative (from console/filesystem)
	if filepath.IsAbs(filename) {
		// Absolute path - use os.Open for direct file access
		handle, err := os.Open(filename)
		if err != nil {
			event.DispatchTyped(eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateError})
			return nil, nil, fmt.Errorf("failed to open map file: %w", err)
		}
		defer handle.Close()
		reader = handle
	} else {
		// Relative path - use virtual filesystem (searches VPKs, registered directories)
		reader, err = fs.GetFile(filename)
		if err != nil {
			event.DispatchTyped(eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateError})
			return nil, nil, fmt.Errorf("map file not found: %s", filename)
		}
	}

	file, err := bsp.NewReader().Read(reader)
	if err != nil {
		event.DispatchTyped(eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateError})
		return nil, nil, fmt.Errorf("failed to parse BSP file: %w", err)
	}
	bspNameParts := strings.Split(filename, "/")
	bspName := bspNameParts[len(bspNameParts)-1]

	console.PrintString(console.LevelInfo, fmt.Sprintf("Map name: %s", bspName))
	console.PrintString(console.LevelInfo, fmt.Sprintf("BSP version: %d", file.Header.Version))
	console.PrintString(console.LevelInfo, fmt.Sprintf("Map revision: %d", file.Header.Revision))

	event.DispatchTyped(eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateBSPParsed})

	// Check cancellation after BSP parse
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	fs.RegisterPakFile(file.Lumps[bsp.LumpPakfile].(*lump.Pakfile))
	// Load the static bsp world
	level, err := loadBSPWorld(fs, file)

	if err != nil {
		event.DispatchTyped(eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateError})
		return nil, nil, err
	}
	level.SetCamera(graphics.NewCamera(
		mgl32.DegToRad(90),
		float32(window.CurrentWindow().Width())/float32(window.CurrentWindow().Height())))
	event.DispatchTyped(eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateGeometryLoaded})

	// Check cancellation after geometry load
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	// Load staticprops
	level.StaticPropDictionary, level.StaticProps = LoadStaticProps(fs, file)
	event.DispatchTyped(eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateStaticPropsLoaded})

	// Check cancellation after static props
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	// Load entities
	ents, err := entity.LoadEntdata(file)
	if err != nil {
		return nil, nil, err
	}

	level.EntityPropDictionary = LoadEntityProps(fs, ents)

	event.DispatchTyped(eventBus, messages.LoadingLevelProgressEvent{State: messages.LoadingProgressStateEntitiesLoaded})

	// Final cancellation check
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	return level, ents, err
}

type bspstructs struct {
	faces       []face.Face
	planes      []plane.Plane
	vertexes    []mgl32.Vec3
	surfEdges   []int32
	edges       [][2]uint16
	texInfos    []texinfo.TexInfo
	dispInfos   []dispinfo.DispInfo
	dispVerts   []dispvert.DispVert
	lightmap    []common.ColorRGBExponent32
	lightmapHDR []common.ColorRGBExponent32
}

// LoadBspMap is the gateway into loading the core static level. Entities are loaded
// elsewhere
// It loads in the following order:
// BSP Geometry
// BSP Materials
// StaticProps (materials loaded as required)
func loadBSPWorld(fs filesystem.FileSystem, file *bsp.Bsp) (*graphics.Bsp, error) {
	bspStructure := bspstructs{
		faces:       file.Lumps[bsp.LumpFaces].(*lump.Face).Data,
		planes:      file.Lumps[bsp.LumpPlanes].(*lump.Planes).Data,
		vertexes:    file.Lumps[bsp.LumpVertexes].(*lump.Vertex).Data,
		surfEdges:   file.Lumps[bsp.LumpSurfEdges].(*lump.Surfedge).Data,
		edges:       file.Lumps[bsp.LumpEdges].(*lump.Edge).Data,
		texInfos:    file.Lumps[bsp.LumpTexInfo].(*lump.TexInfo).Data,
		dispInfos:   file.Lumps[bsp.LumpDispInfo].(*lump.DispInfo).Data,
		dispVerts:   file.Lumps[bsp.LumpDispVerts].(*lump.DispVert).Data,
		lightmap:    file.Lumps[bsp.LumpLighting].(*lump.Lighting).Data,
		lightmapHDR: file.Lumps[bsp.LumpLightingHDR].(*lump.Lighting).Data,
	}

	//MATERIALS
	stringTable := stringtable.NewFromExistingStringTableData(
		file.Lumps[bsp.LumpTexDataStringData].(*lump.TexDataStringData).Data,
		file.Lumps[bsp.LumpTexDataStringTable].(*lump.TexDataStringTable).Data)
	materials := buildUniqueMaterialList(stringTable, &bspStructure.texInfos)

	materialDictionary := buildMaterialDictionary(fs, materials)

	// BSP FACES
	bspMesh := mesh.NewMesh()          // Regular BSP geometry (no blend weights)
	displacementMesh := mesh.NewMesh() // Displacement surfaces (with blend weights)
	bspFaces := make([]graphics.BspFace, len(bspStructure.faces))
	// storeDispFaces until for visibility calculation purposes.
	dispFaces := make([]int, 0)

	var lightmapAtlas *graphics.TextureAtlas
	if console.GetConvarBoolean("hdr_enable") == true {
		if bspStructure.lightmapHDR != nil {
			lightmapAtlas = generateLightmapTexture(bspStructure.faces, bspStructure.lightmapHDR)
		}
	}

	if lightmapAtlas == nil {
		if bspStructure.lightmap != nil {
			lightmapAtlas = generateLightmapTexture(bspStructure.faces, bspStructure.lightmap)
		}
	}

	for idx, f := range bspStructure.faces {
		if f.DispInfo > -1 {
			// This face is a displacement - add to displacement mesh
			bspFaces[idx] = generateDisplacementFace(&bspStructure.faces[idx], &bspStructure, displacementMesh)
			dispFaces = append(dispFaces, idx)
		} else {
			// Regular BSP face - add to regular mesh
			bspFaces[idx] = generateBspFace(&bspStructure.faces[idx], &bspStructure, bspMesh)
		}

		faceVmt, err := stringTable.FindString(int(bspStructure.texInfos[bspStructure.faces[idx].TexInfo].TexData))
		if err != nil {
			console.PrintInterface(console.LevelError, err)
		} else {
			bspFaces[idx].SetMaterial(strings.ToLower(faceVmt))
		}
	}

	if lightmapAtlas != nil {
		console.PrintString(console.LevelInfo, fmt.Sprintf("Lightmap size: %dx%d", lightmapAtlas.Width(), lightmapAtlas.Height()))
	}

	return graphics.NewBsp(file, bspMesh, displacementMesh, bspFaces, dispFaces, materialDictionary, bspStructure.texInfos, lightmapAtlas), nil
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
			console.PrintString(console.LevelError, fmt.Sprintf("Failed to load material: %s, %s", filePath, err.Error()))
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
func generateBspFace(f *face.Face, bspStructure *bspstructs, bspMesh *mesh.BasicMesh) graphics.BspFace {
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
func generateDisplacementFace(f *face.Face, bspStructure *bspstructs, bspMesh *mesh.BasicMesh) graphics.BspFace {
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
			// Calculate vertex indices for this quad
			idxA := int(info.DispVertStart) + x + y*(size+1)
			idxB := int(info.DispVertStart) + x + (y+1)*(size+1)
			idxC := int(info.DispVertStart) + (x + 1) + (y+1)*(size+1)
			idxD := int(info.DispVertStart) + (x + 1) + y*(size+1)

			// Generate vertex positions
			a := generateDispVert(int(info.DispVertStart), x, y, size, corners, firstCorner, &bspStructure.dispVerts)
			b := generateDispVert(int(info.DispVertStart), x, y+1, size, corners, firstCorner, &bspStructure.dispVerts)
			c := generateDispVert(int(info.DispVertStart), x+1, y+1, size, corners, firstCorner, &bspStructure.dispVerts)
			d := generateDispVert(int(info.DispVertStart), x+1, y, size, corners, firstCorner, &bspStructure.dispVerts)

			// Get blend alpha from DispVert for 2-texture blending
			// Alpha ranges from 0-255, normalize to 0.0-1.0
			rawAlphaA := bspStructure.dispVerts[idxA].Alpha
			rawAlphaB := bspStructure.dispVerts[idxB].Alpha
			rawAlphaC := bspStructure.dispVerts[idxC].Alpha
			rawAlphaD := bspStructure.dispVerts[idxD].Alpha

			alphaA := rawAlphaA / 255.0
			alphaB := rawAlphaB / 255.0
			alphaC := rawAlphaC / 255.0
			alphaD := rawAlphaD / 255.0

			// Split into triangles (ABC, ACD)
			bspMesh.AddIndice(uint32(len(bspMesh.Vertices()))/3, (uint32(len(bspMesh.Vertices()))/3)+1, (uint32(len(bspMesh.Vertices()))/3)+2)
			bspMesh.AddVertex(a.X(), a.Y(), a.Z(), b.X(), b.Y(), b.Z(), c.X(), c.Y(), c.Z())
			bspMesh.AddNormal(normal.X(), normal.Y(), normal.Z(), normal.X(), normal.Y(), normal.Z(), normal.X(), normal.Y(), normal.Z())
			bspMesh.AddBlendWeight(alphaA)
			bspMesh.AddBlendWeight(alphaB)
			bspMesh.AddBlendWeight(alphaC)

			bspMesh.AddIndice(uint32(len(bspMesh.Vertices()))/3, (uint32(len(bspMesh.Vertices()))/3)+1, (uint32(len(bspMesh.Vertices()))/3)+2)
			bspMesh.AddVertex(a.X(), a.Y(), a.Z(), c.X(), c.Y(), c.Z(), d.X(), d.Y(), d.Z())
			bspMesh.AddNormal(normal.X(), normal.Y(), normal.Z(), normal.X(), normal.Y(), normal.Z(), normal.X(), normal.Y(), normal.Z())
			bspMesh.AddBlendWeight(alphaA)
			bspMesh.AddBlendWeight(alphaC)
			bspMesh.AddBlendWeight(alphaD)

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

	var tex *graphics.Texture2D
	for _, f := range faces {
		tex = lightmapTextureFromFace(&f, samples)
		lightMapAtlas.AddRaw(tex.Width(), tex.Height(), tex.Image())
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

	// For basic rendering, we only use the first lightstyle's first bumpmap
	// The samples are stored as: [lightstyle0_bump0, lightstyle0_bump1, lightstyle0_bump2, lightstyle0_bump3, lightstyle1_bump0, ...]
	// For non-bumpmapped faces, there's just one set per lightstyle
	// We use the first set (bump0 or the only set)
	sampleOffset := firstSampleIdx

	raw := make([]uint8, (numLuxels)*4)

	// Bounds check
	if sampleOffset < 0 || int(sampleOffset)+int(numLuxels) > len(samples) {
		console.PrintString(console.LevelWarning, fmt.Sprintf("Lightmap out of bounds for face: offset=%d, numLuxels=%d, totalSamples=%d", sampleOffset, numLuxels, len(samples)))
		// Return white texture on error
		for i := 0; i < int(numLuxels); i++ {
			raw[i*4] = 255
			raw[i*4+1] = 255
			raw[i*4+2] = 255
			raw[i*4+3] = 255
		}
		return graphics.NewTexture("__lightmap_subtex__", int(width), int(height), uint32(format.RGBA8888), raw)
	}

	// Read the first lightstyle's first bump sample set (or only sample set for non-bumpmapped)
	for idx, sample := range samples[sampleOffset : sampleOffset+int32(numLuxels)] {
		raw[(idx * 4)] = uint8(math.Min(255, float64(sample.R)*math.Pow(2, float64(sample.Exponent))))
		raw[(idx*4)+1] = uint8(math.Min(255, float64(sample.G)*math.Pow(2, float64(sample.Exponent))))
		raw[(idx*4)+2] = uint8(math.Min(255, float64(sample.B)*math.Pow(2, float64(sample.Exponent))))
		raw[(idx*4)+3] = 255
	}

	return graphics.NewTexture("__lightmap_subtex__", int(width), int(height), uint32(format.RGBA8888), raw)
}
