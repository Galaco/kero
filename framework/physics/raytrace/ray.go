package raytrace

import (
	"github.com/galaco/kero/framework/scene"
	"github.com/galaco/kero/framework/scene/vis"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"sort"
)

type RayCastResult struct {
	T float64
	Hit bool
	Point mgl32.Vec3
}

type Ray struct {

}

func TraceRayBetween(scene *scene.StaticScene, origin mgl32.Vec3, destination mgl32.Vec3) RayCastResult {
	return TraceRay(scene, origin, destination.Sub(origin))
}

func TraceRay(scene *scene.StaticScene, origin mgl32.Vec3, direction mgl32.Vec3) (r RayCastResult) {
	// Step 1: Determine leafs that the ray intersects
	intersectedLeafs := make([]*vis.ClusterLeaf, 0)
	for idx, l := range scene.ClusterLeafs {
		if r := RayIntersectsAxisAlignedBoundingBox(origin, direction, l.Mins, l.Maxs); r.Hit == true {
			intersectedLeafs = append(intersectedLeafs, &scene.ClusterLeafs[idx])
		}
	}
	// Step 2: Sort by closeness to the origin
	sort.Slice(intersectedLeafs, func(i, j int) bool {
		return intersectedLeafs[i].Origin.Sub(origin).Len() < intersectedLeafs[j].Origin.Sub(origin).Len()
	})
	// Step 3: Find all faces within leafs that intersect
	var verts []float32
	var triangle [3]mgl32.Vec3
	for _,l := range intersectedLeafs {
		for _,f := range l.Faces {
			verts = scene.RawBsp.Mesh().Vertices()[f.Offset() : f.Offset()+f.Length()]
			if len(verts) < 9 {
				// Something is broken here; how can a triangle not have 3 verts?
				continue
			}
			triangle = [3]mgl32.Vec3{
				{
					verts[0],
					verts[1],
					verts[2],
				},
				{
					verts[3],
					verts[4],
					verts[5],
				},
				{
					verts[6],
					verts[7],
					verts[8],
				},
			}

			if r = RayIntersectsTriangle(origin, direction, triangle); r.Hit == true {
				return r
			}
		}
	}

	return r
}

func RayIntersectsAxisAlignedBoundingBox(origin, direction, min, max mgl32.Vec3) (r RayCastResult) {
	// Any component of direction could be 0!
	// Address this by using a small number, close to
	// 0 in case any of directions components are 0
	dir := direction
	if dir[0] == 0 {
		dir[0] = 0.00001
	}
	if dir[1] == 0 {
		dir[1] = 0.00001
	}
	if dir[2] == 0 {
		dir[2] = 0.00001
	}

	t1 := float64((min[0] - origin[0]) / dir[0])
	t2 := float64((max[0] - origin[0]) / dir[0])
	t3 := float64((min[1] - origin[1]) / dir[1])
	t4 := float64((max[1] - origin[1]) / dir[1])
	t5 := float64((min[2] - origin[2]) / dir[2])
	t6 := float64((max[2] - origin[2]) / dir[2])

	tmin := math.Max(math.Max(math.Min(t1, t2), math.Min(t3, t4)), math.Min(t5, t6))
	tmax := math.Min(math.Min(math.Max(t1, t2), math.Max(t3, t4)), math.Max(t5, t6))

	// if tmax < 0, ray is intersecting AABB
	// but entire AABB is behing it's origin
	if tmax < 0 {
		return r
	}

	// if tmin > tmax, ray doesn't intersect AABB
	if tmin > tmax {
		return r
	}

	t_result := tmin

	// If tmin is < 0, tmax is closer
	if tmin < 0.0 {
		t_result = tmax
	}

	r.Hit = true
	r.T = t_result
	r.Point = origin.Add(direction).Mul(float32(t_result))
	//	outResult->t = t_result;
	//	outResult->hit = true;

	//normals := []mgl32.Vec3{
	//	{-1, 0, 0},
	//	{1, 0, 0},
	//	{0, -1, 0},
	//	{0, 1, 0},
	//	{0, 0, -1},
	//	{0, 0, 1},
	//}
	//t := []float32{ t1, t2, t3, t4, t5, t6 }
	//
	//for i := 0; i < 6; i++ {
	//	if CMP(t_result, t[i]) {
	//		outResult->normal = normals[i]
	//	}
	//}


	return r
}

const mollerTrumboreEpsilon = float32(0.0000001)
func RayIntersectsTriangle(rayOrigin mgl32.Vec3, rayVector mgl32.Vec3, inTriangle [3]mgl32.Vec3) (r RayCastResult) {
	// Uses mollerTrumboreRayTriangleIntersection
	vertex0 := inTriangle[0]
	vertex1 := inTriangle[1]
	vertex2 := inTriangle[2]
	var edge1, edge2, h, s, q mgl32.Vec3
	var a, f, u, v float32
	edge1 = vertex1.Sub(vertex0)
	edge2 = vertex2.Sub(vertex0)
	h = rayVector.Cross(edge2)
	a = edge1.Dot(h)
	if a > -mollerTrumboreEpsilon && a < mollerTrumboreEpsilon {
		return r // This ray is parallel to this triangle.
	}
	f = 1.0 / a
	s = rayOrigin.Sub(vertex0)
	u = f * s.Dot(h)

	if u < 0.0 || u > 1.0 {
		return r
	}
	q = s.Cross(edge1)
	v = f * rayVector.Dot(q)

	if v < 0.0 || u+v > 1.0 {
		return r
	}
	// At this stage we can compute t to find out where the intersection point is on the line.
	t := f * edge2.Dot(q)

	if t > mollerTrumboreEpsilon { // ray intersection
		r.Hit = true
		r.Point = rayOrigin.Add(rayVector.Mul(t))
		return r
	} else { // This means that there is a line intersection but not a ray intersection.
		return r
	}
}
