//go:build !windows

package i18n

import (
	"os"
	"strings"
)

// SystemLanguages, Windows dışında LANG ortam değişkeninden okur (ör. tr_TR.UTF-8).
func SystemLanguages() []string {
	l := os.Getenv("LANG")
	if l == "" {
		return []string{}
	}
	l, _, _ = strings.Cut(l, ".")
	return []string{strings.ReplaceAll(l, "_", "-")}
}
