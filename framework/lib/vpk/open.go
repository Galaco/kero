package vpk

import (
	"github.com/galaco/vpk2"
)

// OpenVPK Basic wrapper around vpk library.
// Just opens a multi-part vpk (ver 2 only)
func OpenVPK(filepath string) (*vpk.VPK, error) {
	return vpk.Open(vpk.MultiVPK(filepath))
}
