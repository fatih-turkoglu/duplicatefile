package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWithin(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Photos")
	if !Within(filepath.Join(root, "a.jpg"), root) || Within(root+"2", root) {
		t.Fatal("folder boundary was lost")
	}
	if runtime.GOOS != "windows" && Within(filepath.Join(filepath.Dir(root), "photos", "a.jpg"), root) {
		t.Fatal("Unix paths were case folded")
	}
	if !Within(root, string(filepath.Separator)) && runtime.GOOS != "windows" {
		t.Fatal("root does not contain path")
	}
}

func TestSameUsesFileIdentity(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "source")
	if err := os.WriteFile(p, []byte("same"), 0o600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(dir, "alias")
	if err := os.Link(p, alias); err != nil {
		t.Skip(err)
	}
	if !Same(p, alias) {
		t.Fatal("same file was not recognised")
	}
	other := filepath.Join(dir, "other")
	if err := os.WriteFile(other, []byte("same"), 0o600); err != nil {
		t.Fatal(err)
	}
	if Same(p, other) {
		t.Fatal("distinct files treated as one")
	}
}
