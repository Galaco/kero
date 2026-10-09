package renderer

import (
	"reflect"
	"testing"

	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/scene/vis"
	"github.com/galaco/kero/renderer/cache"
	"github.com/go-gl/mathgl/mgl32"
)

func TestDisplacementBatches(t *testing.T) {
	// A camera at the origin looks down -z
	inFront := []float32{-5, -5, -10, 5, -5, -10, 0, 5, -10}
	behind := []float32{-5, -5, 10, 5, -5, 10, 0, 5, 10}
	placed := [][]float32{inFront, inFront, behind, inFront, inFront}
	materials := []string{"sand", "sand_blend", "sand", "sand_blend", "rock"}
	vertices := make([]float32, 0)
	for _, p := range placed {
		vertices = append(vertices, p...)
	}
	faces := make([]*graphics.BspFace, len(placed))
	for idx := range placed {
		face := graphics.NewMeshFace(int32(idx*3), 3, nil, nil, vertices)
		face.SetMaterial(materials[idx])
		faces[idx] = &face
	}

	materialCache := cache.NewMaterialCache()
	sand := cache.NewGpuMaterial(1, &graphics.Material{BaseTextureName: "sand"})
	blend := cache.NewGpuMaterial(2, &graphics.Material{BaseTextureName: "sand", BaseTexture2Name: "path"})
	rock := cache.NewGpuMaterial(3, &graphics.Material{BaseTextureName: "rock"})
	materialCache.Add("sand", sand)
	materialCache.Add("sand_blend", blend)
	materialCache.Add("rock", rock)

	d := newDisplacementBatches(faces, &materialCache)

	// Blended materials are drawn last, so the shader is switched once
	if len(d.batches) != 3 || d.batches[0].material != sand || d.batches[1].material != rock || d.batches[2].material != blend {
		t.Fatalf("got batches %+v, want sand, rock then the blended material", d.batches)
	}
	if d.batches[0].blend || d.batches[1].blend || !d.batches[2].blend {
		t.Errorf("only the material with a second texture should blend")
	}

	frustum := graphics.FrustumFromCamera(graphics.NewCamera(mgl32.DegToRad(90), 1))
	expect := func(pass string, batch *displacementBatch, offsets []int, counts []int32) {
		t.Helper()
		if len(batch.offsets) != len(offsets) || (len(offsets) > 0 && (!reflect.DeepEqual(batch.offsets, offsets) || !reflect.DeepEqual(batch.counts, counts))) {
			t.Errorf("%s: got offsets %v and counts %v, want %v and %v", pass, batch.offsets, batch.counts, offsets, counts)
		}
	}

	// The second displacement is in both clusters, the third is behind the camera and the last is in neither
	first := &vis.ClusterLeaf{DispFaces: []int{0, 1, 2}}
	second := &vis.ClusterLeaf{DispFaces: []int{1, 3}}
	d.collect([]*vis.ClusterLeaf{first, second}, frustum)
	expect("first pass", d.batches[0], []int{0}, []int32{3})
	expect("first pass", d.batches[1], nil, nil)
	expect("first pass", d.batches[2], []int{3, 9}, []int32{3, 3})

	// Each pass lists only its own displacements
	d.collect([]*vis.ClusterLeaf{second}, frustum)
	expect("second pass", d.batches[0], nil, nil)
	expect("second pass", d.batches[2], []int{3, 9}, []int32{3, 3})
}
