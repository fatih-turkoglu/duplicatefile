package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestStatsClearVacuum(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)

	var files []File
	for i := range 2000 {
		files = append(files, File{Path: fmt.Sprintf(`D:\klasör\%04d.txt`, i), Size: int64(i), FullHash: "h"})
	}
	if err := d.UpsertFiles(ctx, files); err != nil {
		t.Fatal(err)
	}
	d.StartScan(ctx, `D:\klasör`)

	s, err := d.Stats(ctx)
	if err != nil || s.Files != 2000 || s.Scans != 1 || s.SizeBytes == 0 {
		t.Fatalf("Stats = %+v, %v", s, err)
	}

	if err := d.Clear(ctx); err != nil {
		t.Fatal(err)
	}
	if s, _ := d.Stats(ctx); s.Files != 0 || s.Scans != 0 {
		t.Errorf("Clear sonrası: %+v", s)
	}

	before, after, err := d.Vacuum(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if after >= before {
		t.Errorf("Vacuum küçültmedi: %d -> %d", before, after)
	}
	// Şema korunmalı: yeni kayıt eklenebilmeli.
	if err := d.UpsertFiles(ctx, files[:1]); err != nil {
		t.Errorf("Clear sonrası yazılamadı: %v", err)
	}
}

func TestPruneMissing(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)
	dir := t.TempDir()
	present := filepath.Join(dir, "var.txt")
	os.WriteFile(present, []byte("x"), 0o644)

	d.UpsertFiles(ctx, []File{
		{Path: present},
		{Path: filepath.Join(dir, "yok.txt")},
		{Path: filepath.Join(dir, "yok-klasor", "a.txt")},
	})
	n, err := d.PruneMissing(ctx)
	if err != nil || n != 2 {
		t.Fatalf("PruneMissing = %d, %v; beklenen 2", n, err)
	}
	if _, err := d.FileByPath(ctx, present); err != nil {
		t.Errorf("var olan dosyanın kaydı silindi: %v", err)
	}
}

func TestCheckpointMakesSingleFile(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)
	d.UpsertFiles(ctx, []File{{Path: `D:\a.txt`, Size: 1}})
	if err := d.Checkpoint(ctx); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(d.Path() + "-wal"); err == nil && info.Size() != 0 {
		t.Errorf("checkpoint sonrası WAL boş değil: %d bayt", info.Size())
	}
}
