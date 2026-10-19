//go:build !windows && !darwin

package fileops

import (
	"errors"
	"syscall"
)

func moveToRecycleBin(path string) error {
	return ErrNoRecycleBin
}

func isCrossDevice(err error) bool {
	return errors.Is(err, syscall.EXDEV)
}
