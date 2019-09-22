package graphics

import (
	"errors"
	keyvalues "github.com/galaco/KeyValues"
	"github.com/galaco/kero/framework/filesystem"
	"strings"
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
	if !strings.HasSuffix(filePath, filesystem.ExtensionVmt) {
		filePath += filesystem.ExtensionVmt
	}
	filePath = filesystem.BasePathMaterial + filePath

	return readVmt(filePath)
}

func readVmt(path string) (*Material, error) {
	root, err := filesystem.ReadKeyValues(path)
	if err != nil {
		return nil, err
	}

	include, err := root.Find("include")
	if include != nil && err == nil {
		includePath, _ := include.AsString()
		root, err = mergeIncludedVmtRecursive(root, includePath)
		if err != nil {
			return nil, err
		}
	}

	// @NOTE this will be replaced with a proper kv->material builder
	mat, err := materialFromKeyValues(root, path)
	if err != nil {
		return nil, err
	}
	return mat, nil
}

func mergeIncludedVmtRecursive(base *keyvalues.KeyValue, includePath string) (*keyvalues.KeyValue, error) {
	parent, err := filesystem.ReadKeyValues(includePath)
	if err != nil {
		return base, errors.New("failed to read included vmt")
	}
	parents, _ := parent.Children()
	result, err := base.Patch(parents[0])
	if err != nil {
		return base, errors.New("failed to merge included vmt")
	}
	include, err := result.Find("include")
	if err == nil {
		newIncludePath, _ := include.AsString()
		if newIncludePath == includePath {
			err = result.RemoveChild("include")
			return &result, err
		}
		return mergeIncludedVmtRecursive(&result, newIncludePath)
	}
	return &result, nil
}

func materialFromKeyValues(kv *keyvalues.KeyValue, path string) (*Material, error) {
	shaderName := kv.Key()

	// $basetexture
	baseTexture := findKeyValueAsString(kv, "$basetexture")

	// $bumpmap
	bumpMapTexture := findKeyValueAsString(kv, "$bumpmap")

	mat := NewMaterial(path)
	mat.ShaderName = shaderName
	mat.BaseTextureName = baseTexture
	mat.BumpMapName = bumpMapTexture
	return mat, nil
}

func findKeyValueAsString(kv *keyvalues.KeyValue, keyName string) string {
	target, err := kv.Find(keyName)
	if err != nil {
		return ""
	}
	ret, err := target.AsString()
	if err != nil {
		return ""
	}

	return ret
}
