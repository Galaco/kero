package stringtable

import (
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/texinfo"
	"github.com/galaco/source-tools-common/texdatastringtable"
)

// NewTable returns a new StringTable
func NewTable(stringData *lumps.TexDataStringData, stringTable *lumps.TexDataStringTable) *texdatastringtable.TexDataStringTable {
	// Prepare texture lookup table
	return texdatastringtable.NewTable(stringData.GetData(), stringTable.GetData())
}

// SortUnique builds a unique list of materials in a StringTable
// referenced by BSP TexInfo lump data.
func SortUnique(stringTable *texdatastringtable.TexDataStringTable, texInfos *[]texinfo.TexInfo) []string {
	materialList := make([]string, 0)
	for _, ti := range *texInfos {
		target, _ := stringTable.GetString(int(ti.TexData))
		found := false
		for _, cur := range materialList {
			if cur == target {
				found = true
				break
			}
		}
		if !found {
			materialList = append(materialList, target)
		}
	}

	return materialList
}
