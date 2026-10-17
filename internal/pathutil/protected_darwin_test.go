//go:build darwin

package pathutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMacProtectedPaths(t *testing.T) {
	for _, path := range []string{
		"/Applications/AdBlock.app/Contents/Resources/rules.json",
		"/System/Library/file", "/System/Volumes/Data/Users/me/file",
		"/Library/Application Support/App/file", "/usr/lib/file", "/private", "/private/var/db/file",
		"/Users/me/Library/Caches/App/file", "/Users/me/Applications/App/file",
		"/Users/me/Downloads/Test.APP/Contents/file", "/Users/me/Pictures/Photos.photoslibrary/file",
		"/Volumes/External/Library/file", "/Volumes/External/Applications/App.app/file",
	} {
		if !IsProtected(path) {
			t.Errorf("not protected: %s", path)
		}
	}
	for _, path := range []string{
		"/", "/Users/me/Documents/a.txt", "/Users/me/Downloads/a.jpg", "/Volumes/External/Photos/a.jpg",
		"/Users/me/Documents/Library/a.txt", "/Users/me/Documents/application-notes.txt",
		filepath.Join(t.TempDir(), "test.txt"),
	} {
		if IsProtected(path) {
			t.Errorf("personal path blocked: %s", path)
		}
	}
}

func TestMacProtectedAlias(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "Test.app", "Contents")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(app, "rules.json")
	if err := os.WriteFile(file, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(dir, "alias")
	if err := os.Symlink(app, alias); err != nil {
		t.Fatal(err)
	}
	if !IsProtected(filepath.Join(alias, "rules.json")) {
		t.Fatal("symlink bypassed bundle protection")
	}
}
