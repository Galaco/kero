package graphics

import (
	"github.com/galaco/vmt"
)

// Material
type Material struct {
	filePath string
	// ShaderName
	ShaderName string
	// BaseTextureName
	BaseTextureName string
	// Skip
	Skip bool
	// Alpha
	Alpha float32
	// Translucent
	Translucent bool
}

// FilePath returns this materials location in whatever
// filesystem it was found
func (mat *Material) FilePath() string {
	return mat.filePath
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
	mat.BaseTextureName = props.BaseTexture

	mat.Alpha = props.Alpha
	if props.Translucent == 1 {
		mat.Translucent = true
	}

	if props.CompileSky == 1 || props.CompileNoDraw == 1 {
		mat.Skip = true
	}

	return mat, nil
}
