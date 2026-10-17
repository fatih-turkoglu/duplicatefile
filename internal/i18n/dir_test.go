package i18n

import (
	"path/filepath"
	"testing"
)

func TestBundleLanguageDir(t *testing.T) {
	contents := filepath.Join(t.TempDir(), "DuplicateFile.app", "Contents")
	exe := filepath.Join(contents, "MacOS", "DuplicateFile")
	if got := languageDir(exe); got != filepath.Join(contents, "Resources", "lang") {
		t.Fatalf("bundle translations: %s", got)
	}
	exe = filepath.Join(t.TempDir(), "DuplicateFile")
	if got := languageDir(exe); got != filepath.Join(filepath.Dir(exe), "lang") {
		t.Fatalf("standalone translations: %s", got)
	}
	override := t.TempDir()
	t.Setenv("DUPLICATEFILE_LANG_DIR", override)
	if got := Dir(); got != override {
		t.Fatalf("override = %s, want %s", got, override)
	}
}
