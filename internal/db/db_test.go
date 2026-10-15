package db

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sort"
	"testing"
)

func openTest(t *testing.T) *DB {
	t.Helper()
	// Klasör adındaki boşluk ve '#' URI kaçışlarını da sınar.
	d, err := Open(context.Background(), filepath.Join(t.TempDir(), "alt klasör #1", "index.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func paths(files []File) []string {
	var p []string
	for _, f := range files {
		p = append(p, f.Path)
	}
	sort.Strings(p)
	return p
}

func TestOpenMigratesAndReopens(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "index.db")

	d, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("ilk Open: %v", err)
	}
	var v int
	d.conn.QueryRow("PRAGMA user_version").Scan(&v)
	if v != SchemaVersion() {
		t.Fatalf("user_version = %d, beklenen %d", v, SchemaVersion())
	}
	var mode string
	d.conn.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if mode != "wal" {
		t.Errorf("journal_mode = %q, beklenen wal", mode)
	}
	d.Close()

	// İkinci açılış migration'ları tekrar uygulamamalı.
	d, err = Open(ctx, path)
	if err != nil {
		t.Fatalf("ikinci Open: %v", err)
	}
	d.Close()
}

// v1 şemasıyla oluşturulmuş bir veritabanı açılınca kayıtlar korunarak v2'ye
// geçmeli (kullanıcıların mevcut önbelleği kaybolmasın).
func TestMigrateV1ToV2KeepsData(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "index.db")

	conn, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(migrations[0]); err != nil {
		t.Fatal(err)
	}
	conn.Exec("PRAGMA user_version = 1")
	conn.Exec(`INSERT INTO files (path, size, mtime, full_hash) VALUES ('D:\eski.txt', 5, 7, 'h')`)
	conn.Close()

	d, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("v1'den açılış: %v", err)
	}
	defer d.Close()
	f, err := d.FileByPath(ctx, `D:\eski.txt`)
	if err != nil || f.FullHash != "h" || f.PHash != "" || f.Width != 0 {
		t.Errorf("taşınan kayıt: %+v %v", f, err)
	}
}

func TestImageFieldsRoundTrip(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)
	d.UpsertFiles(ctx, []File{
		{Path: `D:\foto.jpg`, Size: 10, PHash: "00ff00ff00ff00ff", Width: 4000, Height: 3000},
		{Path: `D:\belge.pdf`, Size: 10},
	})
	f, _ := d.FileByPath(ctx, `D:\foto.jpg`)
	if f.PHash != "00ff00ff00ff00ff" || f.Width != 4000 || f.Height != 3000 {
		t.Errorf("görsel alanları: %+v", f)
	}
	imgs, err := d.FilesWithPHash(ctx)
	if err != nil || len(imgs) != 1 || imgs[0].Path != `D:\foto.jpg` {
		t.Errorf("FilesWithPHash = %+v, %v", imgs, err)
	}
}

func TestOpenRejectsNewerSchema(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "index.db")
	d, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	d.conn.Exec("PRAGMA user_version = 999")
	d.Close()

	if _, err := Open(ctx, path); err == nil {
		t.Fatal("daha yeni şemalı veritabanı açılmamalıydı")
	}
}

func TestUpsertInsertAndUpdate(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)

	err := d.UpsertFiles(ctx, []File{
		{Path: `D:\a\1.txt`, Size: 10, ModTime: 100, PartialHash: "p1", LastSeenScan: 1},
		{Path: `D:\a\2.txt`, Size: 20, ModTime: 200, LastSeenScan: 1},
	})
	if err != nil {
		t.Fatalf("UpsertFiles: %v", err)
	}

	f, err := d.FileByPath(ctx, `D:\a\2.txt`)
	if err != nil {
		t.Fatal(err)
	}
	if f.PartialHash != "" || f.FullHash != "" {
		t.Errorf("hash'ler boş olmalı: %+v", f)
	}

	// Aynı yol (farklı harf büyüklüğüyle) güncelleme yapmalı, yeni satır değil.
	err = d.UpsertFiles(ctx, []File{
		{Path: `d:\A\1.TXT`, Size: 11, ModTime: 101, PartialHash: "p2", FullHash: "f2", LastSeenScan: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	f, err = d.FileByPath(ctx, `D:\a\1.txt`)
	if err != nil {
		t.Fatal(err)
	}
	if f.Size != 11 || f.ModTime != 101 || f.PartialHash != "p2" || f.FullHash != "f2" || f.LastSeenScan != 2 {
		t.Errorf("güncellenmemiş kayıt: %+v", f)
	}
	all, _ := d.FilesUnder(ctx, `D:\`)
	if len(all) != 2 {
		t.Errorf("kayıt sayısı = %d, beklenen 2", len(all))
	}
}

func TestFileByPathNotFound(t *testing.T) {
	d := openTest(t)
	if _, err := d.FileByPath(context.Background(), `C:\yok.txt`); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, beklenen ErrNotFound", err)
	}
}

func TestFilesUnderRange(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)
	d.UpsertFiles(ctx, []File{
		{Path: `D:\Foo\a.txt`},
		{Path: `D:\Foo\alt\b.txt`},
		{Path: `D:\foo\C.txt`},     // harf büyüklüğü farklı, aynı klasör
		{Path: `D:\Foo2\x.txt`},    // önek benzer ama başka klasör
		{Path: `D:\Foo.bak\y.txt`}, // '.' < '\' olan kardeş klasör
		{Path: `D:\Fooz\z.txt`},
		{Path: `E:\Foo\a.txt`},
	})

	for _, root := range []string{`D:\Foo`, `D:\Foo\`, `d:\FOO`} {
		got := paths(mustFilesUnder(t, d, root))
		want := []string{`D:\Foo\a.txt`, `D:\Foo\alt\b.txt`, `D:\foo\C.txt`}
		if !equal(got, want) {
			t.Errorf("FilesUnder(%q) = %v, beklenen %v", root, got, want)
		}
	}

	if got := mustFilesUnder(t, d, `D:\`); len(got) != 6 {
		t.Errorf("FilesUnder(D:\\) = %d kayıt, beklenen 6", len(got))
	}
}

func TestDeleteFiles(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)
	d.UpsertFiles(ctx, []File{{Path: `D:\a.txt`}, {Path: `D:\b.txt`}, {Path: `D:\c.txt`}})

	if err := d.DeleteFiles(ctx, []string{`D:\a.txt`, `d:\C.TXT`, `D:\yok.txt`}); err != nil {
		t.Fatal(err)
	}
	got := paths(mustFilesUnder(t, d, `D:\`))
	if want := []string{`D:\b.txt`}; !equal(got, want) {
		t.Errorf("kalan = %v, beklenen %v", got, want)
	}
}

func TestDuplicateGroups(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)
	d.UpsertFiles(ctx, []File{
		// Büyük grup: 3 kopya, biri kök dışında
		{Path: `D:\r1\big1.bin`, Size: 1000, FullHash: "B"},
		{Path: `D:\r2\big2.bin`, Size: 1000, FullHash: "B"},
		{Path: `E:\disari\big3.bin`, Size: 1000, FullHash: "B"},
		// Küçük grup
		{Path: `D:\r1\s1.txt`, Size: 10, FullHash: "S"},
		{Path: `D:\r1\alt\s2.txt`, Size: 10, FullHash: "S"},
		// Kopyası yalnızca kök dışında olan: grup oluşturmamalı
		{Path: `D:\r1\tek.txt`, Size: 50, FullHash: "T"},
		{Path: `E:\disari\tek.txt`, Size: 50, FullHash: "T"},
		// Hash'i hesaplanmamış aynı boyutlular: grup oluşturmamalı
		{Path: `D:\r1\h1.txt`, Size: 70},
		{Path: `D:\r1\h2.txt`, Size: 70},
	})

	groups, err := d.DuplicateGroups(ctx, []string{`D:\r1`, `D:\r2`}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("grup sayısı = %d, beklenen 2: %+v", len(groups), groups)
	}
	if g := groups[0]; g.Hash != "B" || g.Size != 1000 || len(g.Files) != 2 {
		t.Errorf("ilk grup (en büyük) yanlış: %+v", g)
	}
	if g := groups[1]; g.Hash != "S" || len(g.Files) != 2 {
		t.Errorf("ikinci grup yanlış: %+v", g)
	}

	// minSize küçük grubu elemeli.
	groups, _ = d.DuplicateGroups(ctx, []string{`D:\r1`, `D:\r2`}, 100)
	if len(groups) != 1 || groups[0].Hash != "B" {
		t.Errorf("minSize=100 sonucu yanlış: %+v", groups)
	}

	// Tek kökte "B" grubu yalnızca 1 dosya içerir, grup sayılmamalı.
	groups, _ = d.DuplicateGroups(ctx, []string{`D:\r1`}, 0)
	if len(groups) != 1 || groups[0].Hash != "S" {
		t.Errorf("tek kök sonucu yanlış: %+v", groups)
	}
}

func TestDeleteFile(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)
	d.UpsertFiles(ctx, []File{{Path: `D:\a.txt`}})

	if err := d.DeleteFile(ctx, `D:\a.txt`); err != nil {
		t.Fatal(err)
	}
	if _, err := d.FileByPath(ctx, `D:\a.txt`); !errors.Is(err, ErrNotFound) {
		t.Errorf("kayıt silinmemiş: %v", err)
	}
	if err := d.DeleteFile(ctx, `D:\a.txt`); err != nil {
		t.Errorf("olmayan kaydı silmek hata vermemeli: %v", err)
	}
}

func TestMoveFile(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)
	d.UpsertFiles(ctx, []File{
		{Path: `D:\a.txt`, Size: 1, FullHash: "h1"},
		{Path: `D:\hedef\b.txt`, Size: 2, FullHash: "h2"},
	})

	// Var olan hedefin üzerine taşıma: hedefin eski kaydı silinmeli.
	if err := d.MoveFile(ctx, `D:\a.txt`, `D:\hedef\b.txt`); err != nil {
		t.Fatal(err)
	}
	f, err := d.FileByPath(ctx, `D:\hedef\b.txt`)
	if err != nil {
		t.Fatal(err)
	}
	if f.FullHash != "h1" {
		t.Errorf("taşınan kayıt yanlış: %+v", f)
	}
	if got := mustFilesUnder(t, d, `D:\`); len(got) != 1 {
		t.Errorf("kayıt sayısı = %d, beklenen 1", len(got))
	}

	// Yalnızca harf büyüklüğü değişimi kaydı kaybetmemeli.
	if err := d.MoveFile(ctx, `D:\hedef\b.txt`, `D:\hedef\B.txt`); err != nil {
		t.Fatal(err)
	}
	f, err = d.FileByPath(ctx, `D:\hedef\b.txt`)
	if err != nil || f.Path != `D:\hedef\B.txt` {
		t.Errorf("harf büyüklüğü değişimi: f=%+v err=%v", f, err)
	}

	if err := d.MoveFile(ctx, `D:\yok.txt`, `D:\z.txt`); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, beklenen ErrNotFound", err)
	}
}

func TestScanLifecycle(t *testing.T) {
	ctx := context.Background()
	d := openTest(t)

	id1, err := d.StartScan(ctx, `D:\r`)
	if err != nil {
		t.Fatal(err)
	}
	id2, _ := d.StartScan(ctx, `D:\r`)
	if id2 <= id1 {
		t.Errorf("tarama kimlikleri artmalı: %d, %d", id1, id2)
	}

	s, err := d.ScanByID(ctx, id1)
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != ScanRunning || !s.FinishedAt.IsZero() {
		t.Errorf("yeni tarama: %+v", s)
	}

	if err := d.FinishScan(ctx, id1, 42, ScanCompleted); err != nil {
		t.Fatal(err)
	}
	s, _ = d.ScanByID(ctx, id1)
	if s.Status != ScanCompleted || s.FileCount != 42 || s.FinishedAt.IsZero() {
		t.Errorf("biten tarama: %+v", s)
	}

	if _, err := d.ScanByID(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, beklenen ErrNotFound", err)
	}
}

func mustFilesUnder(t *testing.T, d *DB, root string) []File {
	t.Helper()
	files, err := d.FilesUnder(context.Background(), root)
	if err != nil {
		t.Fatalf("FilesUnder(%q): %v", root, err)
	}
	return files
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
