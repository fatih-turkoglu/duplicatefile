//go:build !windows && !darwin

package main

import "errors"

func showInExplorer(path string) error {
	return errors.New("showing files is supported on Windows and macOS")
}
