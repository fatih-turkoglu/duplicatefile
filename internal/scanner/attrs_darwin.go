//go:build darwin

package scanner

import (
	"io/fs"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func isHiddenOrSystem(info fs.FileInfo) bool {
	if strings.HasPrefix(info.Name(), ".") {
		return true
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Flags&unix.UF_HIDDEN != 0
}
