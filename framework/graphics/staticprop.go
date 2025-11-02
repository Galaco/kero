package graphics

import (
	"github.com/galaco/bsp/lump/primitive/game"
	"github.com/galaco/kero/framework/graphics/mesh"
	"github.com/go-gl/mathgl/mgl32"
)

// StaticProp is a somewhat specialised model
// that implements a few core entity features (largely because
// it is basically a renderable entity that cannot do anything or be reference)
type StaticProp struct {
	Transform       Transform
	leafList        []uint16
	fadeMinDistance float32
	fadeMaxDistance float32
	model           mesh.ModelInstance
}

// Model returns props model
func (prop *StaticProp) Model() *mesh.ModelInstance {
	return &prop.model
}

// LeafList returrns all leafs that this props is in
func (prop *StaticProp) LeafList() []uint16 {
	return prop.leafList
}

func (prop *StaticProp) FadeMinDistance() float32 {
	return prop.fadeMinDistance
}

func (prop *StaticProp) FadeMaxDistance() float32 {
	return prop.fadeMaxDistance
}

// GetTransformedBounds returns the world-space axis-aligned bounding box for this prop
// by transforming the model's local bounds by the prop's transform
func (prop *StaticProp) GetTransformedBounds() (mgl32.Vec3, mgl32.Vec3) {
	if prop.model.Model == nil {
		// Return a small default box if model not loaded
		pos := prop.Transform.Translation
		return pos.Sub(mgl32.Vec3{10, 10, 10}), pos.Add(mgl32.Vec3{10, 10, 10})
	}

	// Get model-space bounds
	modelMins, modelMaxs := prop.model.Model.Bounds()

	// Transform all 8 corners of the bounding box
	transformMatrix := prop.Transform.TransformationMatrix()
	corners := [8]mgl32.Vec3{
		modelMins,
		{modelMaxs.X(), modelMins.Y(), modelMins.Z()},
		{modelMins.X(), modelMaxs.Y(), modelMins.Z()},
		{modelMaxs.X(), modelMaxs.Y(), modelMins.Z()},
		{modelMins.X(), modelMins.Y(), modelMaxs.Z()},
		{modelMaxs.X(), modelMins.Y(), modelMaxs.Z()},
		{modelMins.X(), modelMaxs.Y(), modelMaxs.Z()},
		modelMaxs,
	}

	// Initialize world-space mins/maxs with transformed first corner
	firstCorner := transformMatrix.Mul4x1(corners[0].Vec4(1)).Vec3()
	worldMins := firstCorner
	worldMaxs := firstCorner

	// Transform remaining corners and expand bounds
	for i := 1; i < 8; i++ {
		transformed := transformMatrix.Mul4x1(corners[i].Vec4(1)).Vec3()

		if transformed.X() < worldMins.X() {
			worldMins[0] = transformed.X()
		}
		if transformed.Y() < worldMins.Y() {
			worldMins[1] = transformed.Y()
		}
		if transformed.Z() < worldMins.Z() {
			worldMins[2] = transformed.Z()
		}

		if transformed.X() > worldMaxs.X() {
			worldMaxs[0] = transformed.X()
		}
		if transformed.Y() > worldMaxs.Y() {
			worldMaxs[1] = transformed.Y()
		}
		if transformed.Z() > worldMaxs.Z() {
			worldMaxs[2] = transformed.Z()
		}
	}

	return worldMins, worldMaxs
}

// NewStaticProp returns new StaticProp
func NewStaticProp(lumpProp game.IStaticPropDataLump, propLeafs *game.StaticPropLeafLump, renderable *mesh.Model) *StaticProp {
	prop := StaticProp{
		model: mesh.ModelInstance{
			Model: renderable,
		},
		leafList: make([]uint16, lumpProp.GetLeafCount()),
	}
	for i := uint16(0); i < lumpProp.GetLeafCount(); i++ {
		prop.leafList[i] = propLeafs.Leaf[lumpProp.GetFirstLeaf()+i]
	}

	angles := lumpProp.GetAngles()
	prop.Transform.Translation = lumpProp.GetOrigin()
	prop.Transform.Orientation = mgl32.AnglesToQuat(mgl32.DegToRad(angles[0]), mgl32.DegToRad(angles[1]), mgl32.DegToRad(angles[2]), mgl32.YZX)
	prop.fadeMinDistance = lumpProp.GetFadeMinDist()
	prop.fadeMaxDistance = lumpProp.GetFadeMaxDist()

	return &prop
}
