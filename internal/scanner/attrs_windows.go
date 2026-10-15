//go:build windows

package scanner

import (
	"io/fs"
	"syscall"
)

// isHiddenOrSystem, Windows'ta gizli ya da sistem özniteliği taşıyan
// dosya/klasörleri ayırt eder.
func isHiddenOrSystem(info fs.FileInfo) bool {
	if a, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		return a.FileAttributes&(syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_SYSTEM) != 0
	}
	return false
}
