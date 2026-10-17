//go:build !windows && !darwin

package main

func systemPrefersDark() bool {
	return false
}
