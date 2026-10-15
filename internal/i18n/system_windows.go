//go:build windows

package i18n

import "golang.org/x/sys/windows"

// SystemLanguages, Windows'ta kullanıcının tercih ettiği arayüz dillerini
// öncelik sırasıyla döner (ör. ["tr-TR", "en-US"]).
func SystemLanguages() []string {
	langs, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil {
		return []string{}
	}
	return langs
}
