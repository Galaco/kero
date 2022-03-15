package graphics

import (
	"testing"
)

func TestMaterial_FilePath(t *testing.T) {
	sut := Material{
		filePath: "foo/bar.vmt",
	}

	if sut.FilePath() != "foo/bar.vmt" {
		t.Errorf("incorrect filepath returned. Expected %s, but received: %s", "foo/bar.vmt", sut.FilePath())
	}
}
