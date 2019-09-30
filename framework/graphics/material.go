package graphics

import (
	"github.com/golang-source-engine/vmt"
)

// Material
type Material struct {
	filePath string
	// ShaderName
	ShaderName string
	// BaseTextureName
	BaseTextureName string
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

func LoadMaterial(fs VirtualFileSystem, filePath string) (*Material, error) {
	rawProps, err := vmt.FromFilesystem(filePath, fs, vmt.NewProperties())
	if err != nil {
		return nil, err
	}
	props := rawProps.(*vmt.Properties)
	mat := NewMaterial(filePath)
	mat.BaseTextureName = props.BaseTexture
	return mat, nil
}
