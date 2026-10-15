//go:build !windows

package scanner

import (
	"io/fs"
	"strings"
)

// isHiddenOrSystem, Windows dışında nokta ile başlayan adları gizli sayar.
func isHiddenOrSystem(info fs.FileInfo) bool {
	return strings.HasPrefix(info.Name(), ".")
}
