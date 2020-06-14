package scene

import (
	"fmt"
	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/entity"
	"github.com/galaco/kero/framework/graphics"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
	"github.com/go-gl/mathgl/mgl32"
	"io"
	"sync"
)

type fileSystem interface {
	GetFile(string) (io.Reader, error)
}

type Skybox struct {
	SkyMaterialGpuID uint32
	SkyMesh          graphics.Mesh
	SkyMeshGpuID     graphics.GpuMesh
	SkyMeshTransform graphics3d.Transform
	Origin           mgl32.Vec3
}

func LoadSkybox(fs fileSystem, worldspawn entity.IEntity) *Skybox {
	skyName := worldspawn.ValueForKey("skyname")
	textures, err := loadSkyboxTexture(fs, skyName)
	if err != nil {
		console.PrintString(console.LevelWarning, err.Error())
		return nil
	}

	cameraPosition := worldspawn.VectorForKey("origin")
	// renderScale := worldspawn.FloatForKey("scale")
	skyCube := graphics.NewCube()
	t := graphics3d.Transform{}
	t.Rotation[0] = 90

	return &Skybox{
		SkyMaterialGpuID: graphics.UploadCubemap(textures),
		SkyMesh:          skyCube,
		SkyMeshGpuID:     graphics.UploadMesh(skyCube),
		Origin:           cameraPosition,
		SkyMeshTransform: t,
	}
}

func loadSkyboxTexture(fs fileSystem, skyName string) ([]*graphics.Texture2D, error) {
	var errs [6]error
	sides := make([]*graphics.Texture2D, 6)

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
