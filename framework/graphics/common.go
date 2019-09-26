package graphics

import "io"

const (
	ExtensionVtf     = ".vtf"
	BasePathMaterial = "materials/"
)

type VirtualFileSystem interface {
	GetFile(string) (io.Reader, error)
}
