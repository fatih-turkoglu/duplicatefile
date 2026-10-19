package db

import (
	"context"
	"runtime"
	"testing"
)

func TestPathCollationMatchesPlatform(t *testing.T) {
	d := openTest(t)
	ctx := context.Background()
	a, b := testPath(`D:\Photos\a.jpg`), testPath(`D:\Photos\A.jpg`)
	if err := d.UpsertFiles(ctx, []File{{Path: a, Size: 10}, {Path: b, Size: 20}}); err != nil {
		t.Fatal(err)
	}
	files := mustFilesUnder(t, d, testPath(`D:\Photos`))
	want := 2
	if runtime.GOOS == "windows" {
		want = 1
	}
	if len(files) != want {
		t.Fatalf("case-sensitive entries: %d, want %d", len(files), want)
	}
	if runtime.GOOS != "windows" {
		if err := d.DeleteFile(ctx, b); err != nil {
			t.Fatal(err)
		}
		f, err := d.FileByPath(ctx, a)
		if err != nil || f.Size != 10 {
			t.Fatalf("deleting another case changed the original: %+v %v", f, err)
		}
	}
}
