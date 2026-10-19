// Package pathutil keeps Windows path matching separate from Unix paths.
package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Key preserves case on Unix, including case-sensitive APFS volumes.
func Key(path string) string {
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}

func Same(a, b string) bool {
	if Key(a) == Key(b) {
		return true
	}
	ai, ae := os.Lstat(a)
	bi, be := os.Lstat(b)
	return ae == nil && be == nil && os.SameFile(ai, bi)
}

func Within(path, root string) bool {
	path, root = Key(path), Key(root)
	if path == root {
		return true
	}
	if !strings.HasSuffix(root, string(filepath.Separator)) {
		root += string(filepath.Separator)
	}
	return strings.HasPrefix(path, root)
}
