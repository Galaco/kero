package scene

import (
	"fmt"
	"io"
	"sync"

	"github.com/galaco/kero/internal/framework/console"
	"github.com/galaco/kero/internal/framework/graphics"
	"github.com/galaco/kero/internal/framework/graphics/adapter"
	"github.com/galaco/kero/internal/framework/graphics/mesh"
	"github.com/go-gl/mathgl/mgl32"
)

type fileSystem interface {
	GetFile(string) (io.Reader, error)
}

type Skybox struct {
	SkyMaterialGpuID uint32
	SkyMesh          mesh.Mesh
	SkyMeshGpuID     adapter.GpuMesh
	SkyMeshTransform graphics.Transform
	Origin           mgl32.Vec3
}

func LoadSkybox(fs fileSystem, skyName string, origin mgl32.Vec3) *Skybox {
	textures, err := loadSkyboxTexture(fs, skyName)
	if err != nil {
		console.PrintString(console.LevelWarning, err.Error())
		return nil
	}

	skyCube := mesh.NewCube()
	t := graphics.Transform{}
	t.Orientation = mgl32.AnglesToQuat(mgl32.DegToRad(90), 0, 0, mgl32.XYZ)

	return &Skybox{
		SkyMaterialGpuID: adapter.UploadCubemap(textures),
		SkyMesh:          skyCube,
		SkyMeshGpuID:     adapter.UploadMesh(skyCube),
		Origin:           origin,
		SkyMeshTransform: t,
	}
}

func loadSkyboxTexture(fs fileSystem, skyName string) ([]adapter.Texture, error) {
	var errs [6]error
	sides := make([]adapter.Texture, 6)

	wg := sync.WaitGroup{}
	loadCubemapSide := func(idx int, path string) {
		rawMaterial, err := graphics.LoadMaterial(fs, "skybox/"+path)
		if err != nil {
			errs[idx] = fmt.Errorf("failed to load material: %s, %s", "skybox/"+path, err.Error())
			wg.Done()
			return
		}
		sides[idx], errs[idx] = graphics.LoadTexture(fs, rawMaterial.BaseTextureName)
		wg.Done()
	}

	names := [6]string{"ft", "bk", "up", "dn", "rt", "lf"}
	wg.Add(6)
	for i := 0; i < 6; i++ {
		go loadCubemapSide(i, skyName+names[i])
	}
	wg.Wait()

	for idx, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("failed to load texture: %s, %s", "skybox/"+skyName+names[idx], err.Error())
		}
	}

	return sides, nil
}
