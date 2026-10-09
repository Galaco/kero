package graphics

import (
	"strings"

	"github.com/galaco/vmt"
)

// Material
type Material struct {
	filePath string
	// ShaderName
	ShaderName string
	// BaseTextureName
	BaseTextureName string
	// BaseTexture2Name - for 2-texture blend materials (WorldVertexTransition)
	BaseTexture2Name string
	// BumpMapName is the normal map the surface is lit with, from the lightmap of each bump basis direction
	BumpMapName string
	// BumpMap2Name is the normal map of a blend material's second texture, blended with the first's
	BumpMap2Name string
	// SSBump is true if the normal maps hold how much each bump basis direction lights the surface, rather than its
	// normal (self-shadowing bump maps)
	SSBump bool
	// Skip
	Skip bool
	// Alpha
	Alpha float32
	// Translucent
	Translucent bool
	// NoCull draws both sides of faces
	NoCull bool
	// AlphaTest discards the parts of the base texture whose alpha is below AlphaTestReference
	AlphaTest          bool
	AlphaTestReference float32
}

// FilePath returns this materials location in whatever
// filesystem it was found
func (mat *Material) FilePath() string {
	return mat.filePath
}

// IsTranslucent returns true if this material blends with what is behind it, so it must be drawn after everything
// opaque. $translucent uses the base texture's alpha, and an $alpha below 1 fades the whole material.
func (mat *Material) IsTranslucent() bool {
	return mat.Translucent || (mat.Alpha > 0 && mat.Alpha < 1)
}

// IsBlendMaterial returns true if this material uses 2-texture blending (WorldVertexTransition)
func (mat *Material) IsBlendMaterial() bool {
	return mat.BaseTexture2Name != ""
}

func NewMaterial(filePath string) *Material {
	return &Material{
		filePath: filePath,
	}
}

func LoadMaterial(fs VirtualFileSystem, filePath string) (mat *Material, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	rawProps, err := vmt.FromFilesystem(filePath, fs, vmt.NewProperties())
	if err != nil {
		return nil, err
	}
	props := rawProps.(*vmt.Properties)
	mat = NewMaterial(filePath)
	mat.ShaderName = props.ShaderName
	mat.BaseTextureName = props.BaseTexture
	mat.BaseTexture2Name = props.BaseTexture2
	// Only these shaders light a surface from its normal map with bumped lightmaps. Others use $bumpmap for something
	// else, such as Water's refraction.
	if shader := strings.ToLower(props.ShaderName); shader == "lightmappedgeneric" || shader == "worldvertextransition" {
		mat.BumpMapName = props.Bumpmap
		mat.BumpMap2Name = props.Bumpmap2
		mat.SSBump = props.SSBump == "1"
	}

	mat.Alpha = props.Alpha
	if props.Translucent == 1 {
		mat.Translucent = true
	}
	mat.NoCull = props.NoCull == "1"
	mat.AlphaTest = props.AlphaTest == "1"
	mat.AlphaTestReference = props.AlphaTestReference
	if mat.AlphaTestReference <= 0 {
		// The engine's default
		mat.AlphaTestReference = 0.5
	}

	if props.CompileSky == 1 || props.CompileNoDraw == 1 {
		mat.Skip = true
	}

	return mat, nil
}
