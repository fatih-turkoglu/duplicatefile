//go:build darwin

package main

import "DuplicateFile/internal/macos"

func systemPrefersDark() bool { return macos.DarkMode() }
