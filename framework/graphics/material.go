package graphics

import (
	"github.com/galaco/kero/framework/filesystem"
	"github.com/galaco/lambda-core/loader/material"
)

// Material
type Material struct {
	filePath string
	// ShaderName
	ShaderName string
	// BaseTextureName
	BaseTextureName string
	// BumpMapName
	BumpMapName string
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

func LoadMaterial(filePath string) (*Material, error) {
	props, err := material.LoadVmtFromFilesystem(filesystem.Singleton(), filePath)
	if err != nil {
		return nil, err
	}
	mat := NewMaterial(filePath)
	mat.BaseTextureName = props.BaseTexture
	mat.BumpMapName = props.Bumpmap

	return mat, nil
}
