//go:build !windows

package main

import "errors"

func showInExplorer(path string) error {
	return errors.New("only supported on Windows")
}
