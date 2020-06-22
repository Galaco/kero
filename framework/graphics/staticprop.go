package graphics

import (
	"github.com/galaco/bsp/primitives/game"
	graphics3d "github.com/galaco/kero/framework/graphics/3d"
)

// StaticProp is a somewhat specialised model
// that implements a few core entity features (largely because
// it is basically a renderable entity that cannot do anything or be reference)
type StaticProp struct {
	Transform       graphics3d.Transform
	leafList        []uint16
	model           *Model
	fadeMinDistance float32
	fadeMaxDistance float32
}

// Model returns props model
func (prop *StaticProp) Model() *Model {
	return prop.model
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

// NewStaticProp returns new StaticProp
func NewStaticProp(lumpProp game.IStaticPropDataLump, propLeafs *game.StaticPropLeafLump, renderable *Model) *StaticProp {
	prop := StaticProp{
		model: renderable,
		leafList: make([]uint16, lumpProp.GetLeafCount()),
	}
	for i := uint16(0); i < lumpProp.GetLeafCount(); i++ {
		prop.leafList[i] = propLeafs.Leaf[lumpProp.GetFirstLeaf()+i]
	}
	prop.Transform.Position = lumpProp.GetOrigin()
	prop.Transform.Rotation = lumpProp.GetAngles()
	prop.fadeMinDistance = lumpProp.GetFadeMinDist()
	prop.fadeMaxDistance = lumpProp.GetFadeMaxDist()

	return &prop
}
