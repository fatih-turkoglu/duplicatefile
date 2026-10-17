//go:build darwin

package fileops

import (
	"errors"
	"syscall"

	"DuplicateFile/internal/macos"
)

func moveToRecycleBin(path string) error {
	err := macos.Trash(path)
	if errors.Is(err, macos.ErrTrashUnavailable) {
		return ErrNoRecycleBin
	}
	return err
}

func isCrossDevice(err error) bool { return errors.Is(err, syscall.EXDEV) }
