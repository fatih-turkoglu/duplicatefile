//go:build darwin

package fileops

import (
	"os"
	"path/filepath"
	"testing"

	"DuplicateFile/internal/db"
)

func TestMacTrashFailurePreservesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keep me.txt")
	if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	// A NUL must be rejected rather than silently truncating the native path.
	if err := moveToRecycleBin(path + "\x00ignored"); err == nil {
		t.Fatal("invalid path accepted")
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "keep" {
		t.Fatalf("source changed: %q %v", data, err)
	}
}

func TestMacRecycleIntegration(t *testing.T) {
	if os.Getenv("DUPLICATEFILE_TEST_RECYCLE") != "1" {
		t.Skip("set DUPLICATEFILE_TEST_RECYCLE=1 for native Trash integration")
	}
	path := filepath.Join(t.TempDir(), "DuplicateFile-trash-test.txt")
	if err := os.WriteFile(path, []byte("temporary DuplicateFile test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := moveToRecycleBin(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
	// The native API reports success only after moving the item to Trash.
	// The test is opt-in because macOS keeps the disposable fixture in Trash.
}

func TestMacRenameCaseCollision(t *testing.T) {
	e := setup(t)
	source := e.file("file.txt", "source")
	target := filepath.Join(e.dir, "FILE.txt")
	if _, err := os.Stat(target); err == nil {
		t.Skip("requires a case-sensitive volume")
	}
	if err := os.WriteFile(target, []byte("target"), 0o600); err != nil {
		t.Fatal(err)
	}
	report := e.svc.Rename(e.ctx, []RenameItem{{Path: source, NewName: "FILE.txt"}}, nil)
	if report.Failed != 1 || report.Results[0].Code != "exists" {
		t.Fatalf("collision not refused: %+v", report)
	}
	e.mustHave(source, "source")
	if data, _ := os.ReadFile(target); string(data) != "target" {
		t.Fatalf("target overwritten: %q", data)
	}
}

func TestMacRenameDoesNotReplaceSymlink(t *testing.T) {
	e := setup(t)
	source := e.file("source.txt", "source")
	target := filepath.Join(e.dir, "link.txt")
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	report := e.svc.Rename(e.ctx, []RenameItem{{Path: source, NewName: "link.txt"}}, nil)
	if report.Failed != 1 || report.Results[0].Code != "exists" {
		t.Fatalf("symlink replaced: %+v", report)
	}
	if link, err := os.Readlink(target); err != nil || link != source {
		t.Fatalf("symlink changed: %s %v", link, err)
	}
	e.mustHave(source, "source")
}

func TestMacCachedApplicationFilesCannotBeChanged(t *testing.T) {
	e := setup(t)
	appFile := e.file("AdBlock.app/Contents/Resources/rules.json", "rules")
	copy := e.file("backup.json", "rules")
	for _, permanent := range []bool{false, true} {
		report := e.svc.Delete(e.ctx, []string{appFile}, DeleteOptions{Permanent: permanent}, nil)
		if report.Failed != 1 || report.Results[0].Code != "protected_path" {
			t.Fatalf("bundle deletion allowed: %+v", report)
		}
	}
	report := e.svc.Rename(e.ctx, []RenameItem{{Path: appFile, NewName: "renamed.json"}}, nil)
	if report.Failed != 1 || report.Results[0].Code != "protected_path" {
		t.Fatalf("bundle rename allowed: %+v", report)
	}
	report = e.svc.Move(e.ctx, []string{copy}, filepath.Dir(appFile), ConflictRename, nil)
	if report.Failed != 1 || report.Results[0].Code != "protected_path" {
		t.Fatalf("move into bundle allowed: %+v", report)
	}
	report = e.svc.Delete(e.ctx, []string{copy}, DeleteOptions{Permanent: true}, nil)
	if report.Failed != 1 || report.Results[0].Code != "last_copy" {
		t.Fatalf("protected app resource counted as backup: %+v", report)
	}
	e.mustHave(appFile, "rules")
	e.mustHave(copy, "rules")
}

func TestMacPhysicalAliasDoesNotCountAsLastCopy(t *testing.T) {
	e := setup(t)
	source := e.file("source.txt", "same")
	alias := filepath.Join(e.dir, "alias.txt")
	if err := os.Link(source, alias); err != nil {
		t.Fatal(err)
	}
	f, err := e.db.FileByPath(e.ctx, source)
	if err != nil {
		t.Fatal(err)
	}
	f.Path = alias
	if err := e.db.UpsertFiles(e.ctx, []db.File{f}); err != nil {
		t.Fatal(err)
	}
	report := e.svc.Delete(e.ctx, []string{source}, DeleteOptions{Permanent: true}, nil)
	if report.Failed != 1 || report.Results[0].Code != "last_copy" {
		t.Fatalf("physical alias counted as backup: %+v", report)
	}
	e.mustHave(source, "same")
}
