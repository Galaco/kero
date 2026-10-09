package graphics

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lump/primitive/face"
	"github.com/galaco/bsp/lump/primitive/texinfo"
	mesh2 "github.com/galaco/kero/framework/graphics/mesh"
	"github.com/go-gl/mathgl/mgl32"
)

// TexCoordsForFaceFromTexInfo Generate texturecoordinates for face data
func TexCoordsForFaceFromTexInfo(vertexes []float32, tx *texinfo.TexInfo, width int, height int) []float32 {
	uvs := make([]float32, (len(vertexes)/3)*2)
	for idx := 0; idx < len(vertexes)/3; idx++ {
		//u = tv0,0 * x + tv0,1 * y + tv0,2 * z + tv0,3
		uvs[idx*2] = ((tx.TextureVecsTexelsPerWorldUnits[0][0] * vertexes[(idx*3)]) +
			(tx.TextureVecsTexelsPerWorldUnits[0][1] * vertexes[(idx*3)+1]) +
			(tx.TextureVecsTexelsPerWorldUnits[0][2] * vertexes[(idx*3)+2]) +
			tx.TextureVecsTexelsPerWorldUnits[0][3]) / float32(width)

		//v = tv1,0 * x + tv1,1 * y + tv1,2 * z + tv1,3
		uvs[(idx*2)+1] = ((tx.TextureVecsTexelsPerWorldUnits[1][0] * vertexes[(idx*3)]) +
			(tx.TextureVecsTexelsPerWorldUnits[1][1] * vertexes[(idx*3)+1]) +
			(tx.TextureVecsTexelsPerWorldUnits[1][2] * vertexes[(idx*3)+2]) +
			tx.TextureVecsTexelsPerWorldUnits[1][3]) / float32(height)
	}

	return uvs
}

// LightmapCoordsForFaceFromTexInfo create lightmap coordinates from TexInfo
func LightmapCoordsForFaceFromTexInfo(vertexes []float32,
	faceInfo *face.Face,
	tx *texinfo.TexInfo,
	lightmapWidth float32,
	lightmapHeight float32,
	lightmapOffsetX float32,
	lightmapOffsetY float32) []float32 {
	//vert.lightCoord[0] = DotProduct (vec, MSurf_TexInfo( surfID )->lightmapVecsLuxelsPerWorldUnits[0].AsVector3D()) +
	//	MSurf_TexInfo( surfID )->lightmapVecsLuxelsPerWorldUnits[0][3];
	//vert.lightCoord[0] -= MSurf_LightmapMins( surfID )[0];
	//vert.lightCoord[0] += 0.5f;
	//vert.lightCoord[0] /= ( float )MSurf_LightmapExtents( surfID )[0]; //pSurf->texinfo->texture->width;
	//
	//vert.lightCoord[1] = DotProduct (vec, MSurf_TexInfo( surfID )->lightmapVecsLuxelsPerWorldUnits[1].AsVector3D()) +
	//	MSurf_TexInfo( surfID )->lightmapVecsLuxelsPerWorldUnits[1][3];
	//vert.lightCoord[1] -= MSurf_LightmapMins( surfID )[1];
	//vert.lightCoord[1] += 0.5f;
	//vert.lightCoord[1] /= ( float )MSurf_LightmapExtents( surfID )[1]; //pSurf->texinfo->texture->height;
	//
	//vert.lightCoord[0] = sOffset + vert.lightCoord[0] * sScale;
	//vert.lightCoord[1] = tOffset + vert.lightCoord[1] * tScale;

	uvs := make([]float32, (len(vertexes)/3)*2)

	// Scale calculation uses actual texture dimensions (luxels + 1)
	sScale := 1 / lightmapWidth
	sOffset := lightmapOffsetX * sScale
	sScale = float32(faceInfo.LightmapTextureSizeInLuxels[0] + 1) * sScale

	tScale := 1 / lightmapHeight
	tOffset := lightmapOffsetY * tScale
	tScale = float32(faceInfo.LightmapTextureSizeInLuxels[1] + 1) * tScale

	// 0x00000001 = SURFDRAW_NOLIGHT
	if tx.Flags&0x00000001 != 0 {
		for idx := 0; idx < len(vertexes)/3; idx++ {
			uvs[(idx*2)+0] = 0.5
			uvs[(idx*2)+1] = 0.5
		}
		return uvs
	}

	if faceInfo.LightmapTextureSizeInLuxels[0] == 0 {
		for idx := 0; idx < len(vertexes)/3; idx++ {
			uvs[(idx*2)+0] = sOffset
			uvs[(idx*2)+1] = tOffset
		}
		return uvs
	}

	for idx := 0; idx < len(vertexes)/3; idx++ {
		uvs[(idx*2)+0] =
			(mgl32.Vec3{vertexes[(idx*3)+0], vertexes[(idx*3)+1], vertexes[(idx*3)+2]}).Dot(
				mgl32.Vec3{tx.LightmapVecsLuxelsPerWorldUnits[0][0], tx.LightmapVecsLuxelsPerWorldUnits[0][1], tx.LightmapVecsLuxelsPerWorldUnits[0][2]}) +
				tx.LightmapVecsLuxelsPerWorldUnits[0][3]
		uvs[(idx*2)+0] -= float32(faceInfo.LightmapTextureMinsInLuxels[0])
		uvs[(idx*2)+0] += 0.5
		// Divide by actual texture dimensions (luxels + 1), not luxel count
		uvs[(idx*2)+0] /= float32(faceInfo.LightmapTextureSizeInLuxels[0] + 1)

		uvs[(idx*2)+1] =
			(mgl32.Vec3{vertexes[(idx*3)+0], vertexes[(idx*3)+1], vertexes[(idx*3)+2]}).Dot(
				mgl32.Vec3{tx.LightmapVecsLuxelsPerWorldUnits[1][0], tx.LightmapVecsLuxelsPerWorldUnits[1][1], tx.LightmapVecsLuxelsPerWorldUnits[1][2]}) +
				tx.LightmapVecsLuxelsPerWorldUnits[1][3]
		uvs[(idx*2)+1] -= float32(faceInfo.LightmapTextureMinsInLuxels[1])
		uvs[(idx*2)+1] += 0.5
		// Divide by actual texture dimensions (luxels + 1), not luxel count
		uvs[(idx*2)+1] /= float32(faceInfo.LightmapTextureSizeInLuxels[1] + 1)

		uvs[(idx*2)+0] = sOffset + uvs[(idx*2)+0]*sScale
		uvs[(idx*2)+1] = tOffset + uvs[(idx*2)+1]*tScale
	}

	return uvs
}

// DisplacementLightmapCoords creates lightmap coordinates for a displacement's vertices. A displacement's lightmap
// covers its grid evenly, from the luxel centre at its start corner to the luxel centre at the opposite corner, however
// the displacement is shaped.
func DisplacementLightmapCoords(vertices []DisplacementVertex,
	faceInfo *face.Face,
	tx *texinfo.TexInfo,
	lightmapWidth float32,
	lightmapHeight float32,
	lightmapOffsetX float32,
	lightmapOffsetY float32) []float32 {
	uvs := make([]float32, len(vertices)*2)

	// 0x00000001 = SURFDRAW_NOLIGHT
	if tx.Flags&0x00000001 != 0 {
		for idx := range vertices {
			uvs[(idx*2)+0] = 0.5
			uvs[(idx*2)+1] = 0.5
		}
		return uvs
	}

	for idx, vertex := range vertices {
		uvs[(idx*2)+0] = (lightmapOffsetX + 0.5 + vertex.Grid.X()*float32(faceInfo.LightmapTextureSizeInLuxels[0])) / lightmapWidth
		uvs[(idx*2)+1] = (lightmapOffsetY + 0.5 + vertex.Grid.Y()*float32(faceInfo.LightmapTextureSizeInLuxels[1])) / lightmapHeight
	}

	return uvs
}

// DisplacementVertex is where a displacement's vertex is on the face the displacement is built from
type DisplacementVertex struct {
	// Base is the vertex's position on the face, before it is displaced. Texture coordinates are mapped from it, so a
	// displacement's texture stretches with its shape as it does in the engine.
	Base mgl32.Vec3
	// Grid is the vertex's position across the face, from 0 to 1 along each edge from the start corner: X towards the
	// corner before it, Y towards the corner after it
	Grid mgl32.Vec2
}

// Bsp
type Bsp struct {
	file *bsp.Bsp

	mesh             *mesh2.BasicMesh // Regular BSP faces (no blend weights)
	displacementMesh *mesh2.BasicMesh // Displacement surfaces (with blend weights)
	faces            []BspFace
	dispFaces        []int

	materialDictionary map[string]*Material
	textureInfos       []texinfo.TexInfo

	StaticPropDictionary map[string]*mesh2.Model
	StaticProps          []StaticProp

	EntityPropDictionary map[string]*mesh2.Model

	camera *Camera

	lightmapAtlas *TextureAtlas
}

// BasicMesh returns the regular BSP mesh (no displacements)
func (bsp *Bsp) Mesh() *mesh2.BasicMesh {
	return bsp.mesh
}

// DisplacementMesh returns the displacement mesh (with blend weights)
func (bsp *Bsp) DisplacementMesh() *mesh2.BasicMesh {
	return bsp.displacementMesh
}

// Faces
func (bsp *Bsp) Faces() []BspFace {
	return bsp.faces
}

// DispFaces
func (bsp *Bsp) DispFaces() []int {
	return bsp.dispFaces
}

// MaterialDictionary
func (bsp *Bsp) MaterialDictionary() map[string]*Material {
	return bsp.materialDictionary
}

func (bsp *Bsp) TexInfos() []texinfo.TexInfo {
	return bsp.textureInfos
}

func (bsp *Bsp) Camera() *Camera {
	return bsp.camera
}

func (bsp *Bsp) SetCamera(camera *Camera) {
	bsp.camera = camera
}

func (bsp *Bsp) File() *bsp.Bsp {
	return bsp.file
}

func (bsp *Bsp) LightmapAtlas() *TextureAtlas {
	return bsp.lightmapAtlas
}

// NewBsp
func NewBsp(
	file *bsp.Bsp,
	mesh *mesh2.BasicMesh,
	displacementMesh *mesh2.BasicMesh,
	faces []BspFace,
	dispFaces []int,
	materialDictionary map[string]*Material,
	textureInfos []texinfo.TexInfo,
	lightmapAtlas *TextureAtlas) *Bsp {
	return &Bsp{
		file:               file,
		mesh:               mesh,
		displacementMesh:   displacementMesh,
		faces:              faces,
		dispFaces:          dispFaces,
		materialDictionary: materialDictionary,
		textureInfos:       textureInfos,
		lightmapAtlas:      lightmapAtlas,
	}
}

// BspFace
type BspFace struct {
	offset     int
	length     int
	center     mgl32.Vec3
	mins, maxs mgl32.Vec3
	material   string
	texInfo    *texinfo.TexInfo
	bspFace    *face.Face
	// displacementVertices describes each of a displacement's vertices; it is nil for other faces
	displacementVertices []DisplacementVertex
}

// Offset
func (face *BspFace) Offset() int {
	return face.offset
}

// Length
func (face *BspFace) Length() int {
	return face.length
}

func (face *BspFace) Material() string {
	return face.material
}

func (face *BspFace) SetMaterial(materialPath string) {
	face.material = materialPath
}

func (face *BspFace) TexInfo() *texinfo.TexInfo {
	return face.texInfo
}

func (face *BspFace) RawFace() *face.Face {
	return face.bspFace
}

// DisplacementVertices describes each of a displacement's vertices, in the order they are in the mesh. It is nil for
// other faces.
func (face *BspFace) DisplacementVertices() []DisplacementVertex {
	return face.displacementVertices
}

func (face *BspFace) SetDisplacementVertices(vertices []DisplacementVertex) {
	face.displacementVertices = vertices
}

// NewFace
// NewMeshFace creates a face drawn with length vertices of a mesh's vertices, starting at offset
func NewMeshFace(offset int32, length int32, texInfo *texinfo.TexInfo, bspFace *face.Face, vertices []float32) BspFace {
	center := mgl32.Vec3{}
	var mins, maxs mgl32.Vec3
	for v := offset; v < offset+length; v++ {
		vertex := mgl32.Vec3{vertices[v*3], vertices[v*3+1], vertices[v*3+2]}
		center = center.Add(vertex)
		if v == offset {
			mins, maxs = vertex, vertex
			continue
		}
		for axis := 0; axis < 3; axis++ {
			mins[axis] = min(mins[axis], vertex[axis])
			maxs[axis] = max(maxs[axis], vertex[axis])
		}
	}
	if length > 0 {
		center = center.Mul(1 / float32(length))
	}
	return BspFace{
		offset:  int(offset),
		length:  int(length),
		center:  center,
		mins:    mins,
		maxs:    maxs,
		texInfo: texInfo,
		bspFace: bspFace,
	}
}

// Center returns the average position of the face's vertices
func (face *BspFace) Center() mgl32.Vec3 {
	return face.center
}

// Bounds returns the smallest box that holds the face's vertices
func (face *BspFace) Bounds() (mins, maxs mgl32.Vec3) {
	return face.mins, face.maxs
}
