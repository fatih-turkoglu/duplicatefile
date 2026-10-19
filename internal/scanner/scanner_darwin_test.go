//go:build darwin

package scanner

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func TestMacSystemFilterSkipsAppAndExplicitBundleRoot(t *testing.T) {
	root := t.TempDir()
	write(t, root, "a.txt", []byte("same"))
	write(t, root, "b.txt", []byte("same"))
	app := write(t, root, "AdBlock.app/Contents/Resources/rules.json", []byte("same"))
	s := New(openDB(t))
	// Populate the cache first; enabling the filter must also exclude old entries.
	all := scan(t, s, Options{Roots: []string{root}})
	if all.FilesSeen != 3 {
		t.Fatalf("unfiltered: %d", all.FilesSeen)
	}
	filtered := scan(t, s, Options{Roots: []string{root}, SkipHidden: true})
	if filtered.FilesSeen != 2 || len(filtered.Groups) != 1 || len(filtered.Groups[0].Files) != 2 {
		t.Fatalf("bundle leaked into results: %+v", filtered)
	}
	for _, selectedRoot := range []string{filepath.Join(root, "AdBlock.app"), filepath.Dir(app)} {
		res, err := s.Scan(context.Background(), Options{Roots: []string{selectedRoot}, SkipHidden: true}, nil)
		if err != nil || res.FilesSeen != 0 {
			t.Fatalf("explicit bundle root was scanned: %+v %v", res, err)
		}
	}
}

func TestMacHiddenFlag(t *testing.T) {
	root := t.TempDir()
	p := write(t, root, "hidden.txt", []byte("same"))
	write(t, root, "visible.txt", []byte("same"))
	if err := syscall.Chflags(p, unix.UF_HIDDEN); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil || !isHiddenOrSystem(info) {
		t.Fatalf("Finder hidden flag ignored: %v", err)
	}
	s := New(openDB(t))
	res := scan(t, s, Options{Roots: []string{root}, SkipHidden: true})
	if res.FilesSeen != 1 {
		t.Fatalf("hidden file scanned: %d", res.FilesSeen)
	}
	res = scan(t, s, Options{Roots: []string{root}})
	if len(res.Groups) != 1 {
		t.Fatal("show-hidden scan did not include both copies")
	}
}

func TestMacTrashExcludedEvenWhenShowingHidden(t *testing.T) {
	root := t.TempDir()
	write(t, root, "visible.txt", []byte("same"))
	for _, dir := range []string{".Trash", ".Trashes", ".Spotlight-V100", ".fseventsd"} {
		write(t, root, filepath.Join(dir, "copy.txt"), []byte("same"))
	}
	res := scan(t, New(openDB(t)), Options{Roots: []string{root}})
	if res.FilesSeen != 1 || len(res.Groups) != 0 {
		t.Fatalf("system metadata or Trash scanned: %+v", res)
	}
}
