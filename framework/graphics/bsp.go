package graphics

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/primitives/face"
	"github.com/galaco/bsp/primitives/texinfo"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
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

	sScale := 1 / lightmapWidth
	sOffset := lightmapOffsetX * sScale
	sScale = float32(faceInfo.LightmapTextureSizeInLuxels[0]) * sScale

	tScale := 1 / lightmapHeight
	tOffset := lightmapOffsetY * tScale
	tScale = float32(faceInfo.LightmapTextureSizeInLuxels[1]) * tScale

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
		uvs[(idx*2)+0] /= float32(faceInfo.LightmapTextureSizeInLuxels[0])

		uvs[(idx*2)+1] =
			(mgl32.Vec3{vertexes[(idx*3)+0], vertexes[(idx*3)+1], vertexes[(idx*3)+2]}).Dot(
				mgl32.Vec3{tx.LightmapVecsLuxelsPerWorldUnits[1][0], tx.LightmapVecsLuxelsPerWorldUnits[1][1], tx.LightmapVecsLuxelsPerWorldUnits[1][2]}) +
				tx.LightmapVecsLuxelsPerWorldUnits[1][3]
		uvs[(idx*2)+1] -= float32(faceInfo.LightmapTextureMinsInLuxels[1])
		uvs[(idx*2)+1] += 0.5
		uvs[(idx*2)+1] /= float32(faceInfo.LightmapTextureSizeInLuxels[1])

		uvs[(idx*2)+0] = sOffset + uvs[(idx*2)+0]*sScale
		uvs[(idx*2)+1] = tOffset + uvs[(idx*2)+1]*tScale
	}

	return uvs
}

// Bsp
type Bsp struct {
	file *bsp.Bsp

	mesh      *BasicMesh
	faces     []BspFace
	dispFaces []int

	materialDictionary map[string]*Material
	textureInfos       []texinfo.TexInfo

	StaticPropDictionary map[string]*Model
	StaticProps          []StaticProp

	camera *graphics3d.Camera

	lightmapAtlas *TextureAtlas
}

// BasicMesh
func (bsp *Bsp) Mesh() *BasicMesh {
	return bsp.mesh
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

func (bsp *Bsp) Camera() *graphics3d.Camera {
	return bsp.camera
}

func (bsp *Bsp) SetCamera(camera *graphics3d.Camera) {
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
	mesh *BasicMesh,
	faces []BspFace,
	dispFaces []int,
	materialDictionary map[string]*Material,
	textureInfos []texinfo.TexInfo,
	lightmapAtlas *TextureAtlas) *Bsp {
	return &Bsp{
		file:               file,
		mesh:               mesh,
		faces:              faces,
		dispFaces:          dispFaces,
		materialDictionary: materialDictionary,
		textureInfos:       textureInfos,
		lightmapAtlas:      lightmapAtlas,
	}
}

// BspFace
type BspFace struct {
	offset   int
	length   int
	material string
	texInfo  *texinfo.TexInfo
	bspFace  *face.Face
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

// NewFace
func NewMeshFace(offset int32, length int32, texInfo *texinfo.TexInfo, bspFace *face.Face) BspFace {
	return BspFace{
		offset:  int(offset),
		length:  int(length),
		texInfo: texInfo,
		bspFace: bspFace,
	}
}
