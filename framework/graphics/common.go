package graphics

import "io"

const (
	ExtensionVtf     = ".vtf"
	BasePathMaterial = "materials/"
	BasePathModel    = "models/"
)

type VirtualFileSystem interface {
	GetFile(string) (io.Reader, error)
}
