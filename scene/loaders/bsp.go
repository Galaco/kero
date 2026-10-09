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
	"github.com/galaco/bsp/lump/primitive/texdata"
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
	texDatas    []texdata.TexData
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
		texDatas:    file.Lumps[bsp.LumpTexData].(*lump.TexData).Data,
		dispInfos:   file.Lumps[bsp.LumpDispInfo].(*lump.DispInfo).Data,
		dispVerts:   file.Lumps[bsp.LumpDispVerts].(*lump.DispVert).Data,
		lightmap:    file.Lumps[bsp.LumpLighting].(*lump.Lighting).Data,
		lightmapHDR: file.Lumps[bsp.LumpLightingHDR].(*lump.Lighting).Data,
	}

	//MATERIALS
	stringTable := stringtable.NewFromExistingStringTableData(
		file.Lumps[bsp.LumpTexDataStringData].(*lump.TexDataStringData).Data,
		file.Lumps[bsp.LumpTexDataStringTable].(*lump.TexDataStringTable).Data)
	materials := buildUniqueMaterialList(stringTable, &bspStructure.texInfos, bspStructure.texDatas)

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
			lightmapAtlas = generateLightmapTexture(bspStructure.faces, bspStructure.texInfos, bspStructure.lightmapHDR)
		}
	}

	if lightmapAtlas == nil {
		if bspStructure.lightmap != nil {
			lightmapAtlas = generateLightmapTexture(bspStructure.faces, bspStructure.texInfos, bspStructure.lightmap)
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

		faceVmt, err := materialName(stringTable, bspStructure.texDatas, &bspStructure.texInfos[bspStructure.faces[idx].TexInfo])
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
func buildUniqueMaterialList(stringTable *stringtable.StringTable, texInfos *[]texinfo.TexInfo, texDatas []texdata.TexData) []string {
	materialList := make([]string, 0)
	for _, ti := range *texInfos {
		target, _ := materialName(stringTable, texDatas, &ti)
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

// materialName resolves the material name of a TexInfo. TexInfo references
// a TexData entry, which in turn references the material name in the StringTable.
func materialName(stringTable *stringtable.StringTable, texDatas []texdata.TexData, ti *texinfo.TexInfo) (string, error) {
	if ti.TexData < 0 || int(ti.TexData) >= len(texDatas) {
		return "", fmt.Errorf("texdata index %d out of range", ti.TexData)
	}
	return stringTable.FindString(int(texDatas[ti.TexData].NameStringTableID))
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

	return graphics.NewMeshFace(offset, length, &bspStructure.texInfos[f.TexInfo], f, bspMesh.Vertices())
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

	vertexOffset := uint32(len(bspMesh.Vertices()) / 3)
	indexOffset := int32(len(bspMesh.Indices()))
	onFace := make([]graphics.DisplacementVertex, 0, (size+1)*(size+1))

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

	// A vertex for each point of the grid, in the order of the displacement's vertices: a row of x at a time
	for y := 0; y <= size; y++ {
		for x := 0; x <= size; x++ {
			position, vertex := generateDispVert(int(info.DispVertStart), x, y, size, corners, firstCorner, &bspStructure.dispVerts)
			bspMesh.AddVertex(position.X(), position.Y(), position.Z())
			bspMesh.AddNormal(normal.X(), normal.Y(), normal.Z())
			// Blend alpha ranges from 0-255, normalize to 0.0-1.0 for 2-texture blending
			bspMesh.AddBlendWeight(bspStructure.dispVerts[int(info.DispVertStart)+x+y*(size+1)].Alpha / 255.0)
			onFace = append(onFace, vertex)
		}
	}

	// Split each square of the grid into triangles (ABC, ACD)
	vertexAt := func(x, y int) uint32 {
		return vertexOffset + uint32(x+y*(size+1))
	}
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			a, b, c, d := vertexAt(x, y), vertexAt(x, y+1), vertexAt(x+1, y+1), vertexAt(x+1, y)
			bspMesh.AddIndice(a, b, c, a, c, d)
		}
	}

	dispFace := graphics.NewIndexedMeshFace(indexOffset, int32(size*size*6), int32(vertexOffset), int32((size+1)*(size+1)), &bspStructure.texInfos[f.TexInfo], f, bspMesh.Vertices())
	dispFace.SetDisplacementVertices(onFace)
	return dispFace
}

// generateDispVert returns the position of the displacement vertex at x,y in its grid, and where it is on the face
func generateDispVert(offset int, x int, y int, size int, corners []mgl32.Vec3, firstCorner int32, dispVerts *[]dispvert.DispVert) (mgl32.Vec3, graphics.DisplacementVertex) {
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

	return origin.Add(vert.Vec.Mul(vert.Dist)), graphics.DisplacementVertex{Base: origin, Grid: mgl32.Vec2{tx, ty}}
}

func generateLightmapTexture(faces []face.Face, texInfos []texinfo.TexInfo, samples []common.ColorRGBExponent32) *graphics.TextureAtlas {
	lightMapAtlas := graphics.NewPagedTextureAtlas(graphics.LightmapPages)

	for idx := range faces {
		width, height, colour := lightmapFromFace(&faces[idx], &texInfos[faces[idx].TexInfo], samples)
		lightMapAtlas.AddRaw(width, height, colour)
	}

	lightMapAtlas.Pack()

	return lightMapAtlas
}

const (
	// surfNoLight is the texinfo flag of a face that the map's lighting doesn't light, such as one whose material
	// doesn't use a lightmap
	surfNoLight = 0x400
	// surfBumpLight is the texinfo flag of a face lit for bump mapping
	surfBumpLight = 0x800
)

// lightmapFromFace returns a face's lightmap for each page of the lightmap atlas, as RGBA colour one page after another
func lightmapFromFace(f *face.Face, tx *texinfo.TexInfo, samples []common.ColorRGBExponent32) (width int, height int, colour []uint8) {
	// A face without samples has a single luxel. The engine lights a face that isn't meant to be lit fully, as its
	// material would draw it without a lightmap. VRAD leaves out the samples of any other face that no light reaches,
	// which is unlit.
	if f.Lightofs == -1 {
		luxel := uint8(0)
		if tx.Flags&surfNoLight != 0 {
			luxel = linearToLightmap[1024]
		}
		colour = make([]uint8, 4*graphics.LightmapPages)
		for page := 0; page < graphics.LightmapPages; page++ {
			colour[page*4], colour[page*4+1], colour[page*4+2], colour[page*4+3] = luxel, luxel, luxel, 255
		}
		return 1, 1, colour
	}

	width = int(f.LightmapTextureSizeInLuxels[0] + 1)
	height = int(f.LightmapTextureSizeInLuxels[1] + 1)
	numLuxels := width * height
	pageSize := numLuxels * 4
	colour = make([]uint8, pageSize*graphics.LightmapPages)

	// A face's samples are a lightmap for each of its light styles; only the first is used. A face lit for bump mapping
	// has four lightmaps for each light style: its lightmap, then the lightmap of each bump basis direction.
	bumped := tx.Flags&surfBumpLight != 0
	numSamples := numLuxels
	if bumped {
		numSamples *= 4
	}
	firstSampleIdx := int(f.Lightofs / 4) // 4 = size of ColorRGBExponent32

	// Bounds check
	if firstSampleIdx < 0 || firstSampleIdx+numSamples > len(samples) {
		console.PrintString(console.LevelWarning, fmt.Sprintf("Lightmap out of bounds for face: offset=%d, numLuxels=%d, totalSamples=%d", firstSampleIdx, numSamples, len(samples)))
		// Return white texture on error
		for i := range colour {
			colour[i] = 255
		}
		return width, height, colour
	}

	for idx := 0; idx < numLuxels; idx++ {
		sample := samples[firstSampleIdx+idx]
		var pages [graphics.LightmapPages][3]uint8
		pages[0] = [3]uint8{
			luxelToLightmap(sample.R, sample.Exponent),
			luxelToLightmap(sample.G, sample.Exponent),
			luxelToLightmap(sample.B, sample.Exponent),
		}
		if bumped {
			bumps := bumpedLuxelsToLightmap(sample, [3]common.ColorRGBExponent32{
				samples[firstSampleIdx+numLuxels+idx],
				samples[firstSampleIdx+2*numLuxels+idx],
				samples[firstSampleIdx+3*numLuxels+idx],
			})
			copy(pages[1:], bumps[:])
		} else {
			// A face that isn't lit for bump mapping is lit the same from every direction
			for page := 1; page < graphics.LightmapPages; page++ {
				pages[page] = pages[0]
			}
		}
		for page, pixel := range pages {
			offset := page*pageSize + idx*4
			colour[offset], colour[offset+1], colour[offset+2], colour[offset+3] = pixel[0], pixel[1], pixel[2], 255
		}
	}

	return width, height, colour
}

// linearToLightmap converts light, from 0 to 4 in 1024ths, to how the engine stores it in a lightmap texture: gamma
// corrected for a 2.2 gamma screen, and halved, so that shaders double it back to light a surface up to twice as
// brightly as its texture. This is mathlib's lineartolightmap table, built for the engine's default gamma and
// overbright.
var linearToLightmap = func() (table [4096]uint8) {
	const gamma = 2.2
	const overbrightFactor = 0.5
	for i := range table {
		table[i] = uint8(min(255, math.Round(math.Pow(float64(i)/1024, 1/gamma)*255*overbrightFactor)))
	}
	return table
}()

// linearToVertexLight converts light, from 0 to 4 in 1024ths, to how the engine stores it in a lightmap texture, as
// linearToLightmap does, from 0 to 1. This is mathlib's lineartovertex table.
var linearToVertexLight = func() (table [4096]float64) {
	const gamma = 2.2
	const overbrightFactor = 0.5
	for i := range table {
		table[i] = min(1, math.Pow(float64(i)/1024, 1/gamma)*overbrightFactor)
	}
	return table
}()

// luxelToLinear converts a colour channel of a lightmap sample, stored with a shared exponent, to linear light
func luxelToLinear(colour uint8, exponent int8) float64 {
	return float64(colour) * math.Pow(2, float64(exponent)) / 255
}

// luxelToLightmap converts a colour channel of a lightmap sample, stored with a shared exponent, to a lightmap texture
// value
func luxelToLightmap(colour uint8, exponent int8) uint8 {
	return linearToLightmap[int(min(4091, math.Round(luxelToLinear(colour, exponent)*1024)))]
}

// bumpedLuxelsToLightmap converts a luxel of a face lit for bump mapping, lit from each bump basis direction, to how
// the engine stores it in the lightmap of each direction. The engine gamma corrects the light that reaches the face,
// from flat, as it does for any lightmap, and scales the light from each direction to average it: a surface whose
// normal map faces straight out of it, which is lit equally by each direction's lightmap, is lit as it would be
// without a normal map.
func bumpedLuxelsToLightmap(flat common.ColorRGBExponent32, bumps [3]common.ColorRGBExponent32) (lightmaps [3][3]uint8) {
	flatChannels := [3]uint8{flat.R, flat.G, flat.B}
	for channel := 0; channel < 3; channel++ {
		var light [3]float64
		for direction, bump := range bumps {
			light[direction] = luxelToLinear([3]uint8{bump.R, bump.G, bump.B}[channel], bump.Exponent)
		}
		average := (light[0] + light[1] + light[2]) / 3
		goal := linearToVertexLight[int(min(4095, math.Round(luxelToLinear(flatChannels[channel], flat.Exponent)*1024)))]

		scale := 0.0
		if average != 0 {
			scale = goal / average
		}
		for direction := range light {
			lightmaps[direction][channel] = uint8(math.Round(min(1, light[direction]*scale) * 255))
		}
	}
	return lightmaps
}
