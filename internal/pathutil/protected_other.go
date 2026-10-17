//go:build !darwin

package pathutil

func IsProtected(path string) bool { return false }
