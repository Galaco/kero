package renderer

import (
	"fmt"
	"sort"

	"github.com/galaco/kero/framework/console"
	"github.com/galaco/kero/framework/graphics"
	"github.com/galaco/kero/framework/scene/vis"
	"github.com/galaco/kero/renderer/cache"
)

// displacementBatch is the displacements drawn with one material
type displacementBatch struct {
	material *cache.GpuMaterial
	// blend is true for a material that blends two textures across its displacements (WorldVertexTransition)
	blend bool
	// firsts and counts are the vertices of the batch's displacements to draw in a pass
	firsts, counts []int32
}

// displacementBatches groups displacements by material, so that a pass draws all of a material's displacements
// that it sees at once
type displacementBatches struct {
	faces []*graphics.BspFace
	// batches holds every blended material after the others, so that a pass switches shader once
	batches []*displacementBatch
	// batchOf is the batch of each displacement, or -1 for one with no material
	batchOf []int
	// listed is the pass each displacement was last listed in. A displacement in more than one cluster is listed once.
	listed []uint32
	pass   uint32
}

// newDisplacementBatches groups displacement faces by their material
func newDisplacementBatches(faces []*graphics.BspFace, materials *cache.Material) *displacementBatches {
	d := &displacementBatches{
		faces:   faces,
		batchOf: make([]int, len(faces)),
		listed:  make([]uint32, len(faces)),
	}

	byMaterial := map[*cache.GpuMaterial]*displacementBatch{}
	for _, face := range faces {
		mat := materials.Find(face.Material())
		if mat == nil || byMaterial[mat] != nil {
			continue
		}
		batch := &displacementBatch{material: mat, blend: mat.Properties != nil && mat.Properties.IsBlendMaterial()}
		if batch.blend && mat.Diffuse2 == 0 {
			console.PrintString(console.LevelWarning, fmt.Sprintf("Blended displacement material %s has no second texture", face.Material()))
		}
		byMaterial[mat] = batch
		d.batches = append(d.batches, batch)
	}
	sort.SliceStable(d.batches, func(i, j int) bool {
		return !d.batches[i].blend && d.batches[j].blend
	})

	batchIndex := map[*displacementBatch]int{}
	for idx, batch := range d.batches {
		batchIndex[batch] = idx
	}
	for idx, face := range faces {
		d.batchOf[idx] = -1
		if mat := materials.Find(face.Material()); mat != nil {
			d.batchOf[idx] = batchIndex[byMaterial[mat]]
		}
	}

	return d
}

// collect lists the displacements in clusters that are inside frustum in their material's batch, replacing those
// listed for the previous pass
func (d *displacementBatches) collect(clusters []*vis.ClusterLeaf, frustum *graphics.Frustum) {
	d.pass++
	for _, batch := range d.batches {
		batch.firsts = batch.firsts[:0]
		batch.counts = batch.counts[:0]
	}

	for _, cluster := range clusters {
		for _, idx := range cluster.DispFaces {
			if d.listed[idx] == d.pass {
				continue
			}
			d.listed[idx] = d.pass
			if d.batchOf[idx] == -1 {
				continue
			}
			face := d.faces[idx]
			if !frustum.IsCuboidInFrustum(face.Bounds()) {
				continue
			}
			batch := d.batches[d.batchOf[idx]]
			batch.firsts = append(batch.firsts, int32(face.Offset()))
			batch.counts = append(batch.counts, int32(face.Length()))
		}
	}
}
