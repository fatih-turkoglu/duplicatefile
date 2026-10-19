//go:build darwin

package i18n

import "DuplicateFile/internal/macos"

// Finder-launched applications do not inherit the terminal's LANG variable.
func SystemLanguages() []string { return macos.PreferredLanguages() }
