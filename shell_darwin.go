//go:build darwin

package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"DuplicateFile/internal/apperr"
)

// open -R selects the item in Finder; arguments are never passed to a shell.
func showInExplorer(path string) error {
	if !filepath.IsAbs(path) {
		return apperr.Errorf("path_not_absolute", "path must be absolute: %s", path)
	}
	if _, err := os.Lstat(path); err != nil {
		return err
	}
	return exec.Command("/usr/bin/open", "-R", path).Run()
}
