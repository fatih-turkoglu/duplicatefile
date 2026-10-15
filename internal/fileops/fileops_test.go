package fileops

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"DuplicateFile/internal/db"
)

type env struct {
	t    *testing.T
	dir  string
	db   *db.DB
	svc  *Service
	ctx  context.Context
	past time.Time
}

func setup(t *testing.T) *env {
	t.Helper()
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return &env{t: t, dir: t.TempDir(), db: d, svc: New(d), ctx: ctx,
		past: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)}
}

// file, içerikli bir dosya oluşturup DB'ye hash'iyle kaydeder.
func (e *env) file(rel, content string) string {
	e.t.Helper()
	p := filepath.Join(e.dir, rel)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		e.t.Fatal(err)
	}
	os.Chtimes(p, e.past, e.past)
	if err := e.db.UpsertFiles(e.ctx, []db.File{{Path: p, Size: int64(len(content)),
		ModTime: e.past.UnixNano(), FullHash: "h-" + content}}); err != nil {
		e.t.Fatal(err)
	}
	return p
}

func (e *env) mustNotExist(p string) {
	e.t.Helper()
	if _, err := os.Lstat(p); !errors.Is(err, os.ErrNotExist) {
		e.t.Errorf("%s hâlâ var", p)
	}
	if _, err := e.db.FileByPath(e.ctx, p); !errors.Is(err, db.ErrNotFound) {
		e.t.Errorf("%s DB kaydı hâlâ var", p)
	}
}

func (e *env) mustHave(p, content string) {
	e.t.Helper()
	b, err := os.ReadFile(p)
	if err != nil || string(b) != content {
		e.t.Errorf("%s içeriği = %q, %v; beklenen %q", p, b, err, content)
	}
}

func (e *env) mustRecord(p, content string) {
	e.t.Helper()
	f, err := e.db.FileByPath(e.ctx, p)
	if err != nil || f.FullHash != "h-"+content {
		e.t.Errorf("%s DB kaydı = %+v, %v", p, f, err)
	}
}

func TestDeletePermanent(t *testing.T) {
	e := setup(t)
	a := e.file("a.txt", "aaa")
	kopya := e.file(`yedek\a.txt`, "aaa") // a silinebilsin diye kopyası
	b := e.file(`alt\b.txt`, "bbb")

	var calls int
	rep := e.svc.Delete(e.ctx, []string{a, filepath.Join(e.dir, "yok.txt"), filepath.Dir(b), "goreli.txt"},
		DeleteOptions{Permanent: true}, func(done, total int) { calls++ })

	if rep.Succeeded != 1 || rep.Failed != 3 {
		t.Fatalf("rapor: %+v", rep)
	}
	if calls != 4 {
		t.Errorf("ilerleme çağrısı = %d, beklenen 4", calls)
	}
	e.mustNotExist(a)
	e.mustHave(kopya, "aaa")
	e.mustHave(b, "bbb") // klasör silinmemeli
	for _, r := range rep.Results[1:] {
		if r.Code != "unverified" {
			t.Errorf("%s: kod = %q, beklenen unverified", r.Path, r.Code)
		}
	}
}

func TestDeleteLastCopyGuard(t *testing.T) {
	e := setup(t)
	a1 := e.file(`a\x.txt`, "aaa")
	a2 := e.file(`b\x.txt`, "aaa")
	a3 := e.file(`c\x.txt`, "aaa")
	b1 := e.file(`a\y.txt`, "bbb")
	b2 := e.file(`b\y.txt`, "bbb")

	// a'nın bütün kopyaları istenirse hiçbiri silinmemeli; b'nin biri kalıyor.
	rep := e.svc.Delete(e.ctx, []string{a1, a2, a3, b1}, DeleteOptions{Permanent: true}, nil)
	if rep.Succeeded != 1 || rep.Failed != 3 {
		t.Fatalf("rapor: %+v", rep)
	}
	for _, r := range rep.Results[:3] {
		if r.Code != "last_copy" {
			t.Errorf("%s: kod = %q, beklenen last_copy", r.Path, r.Code)
		}
	}
	e.mustHave(a1, "aaa")
	e.mustHave(a2, "aaa")
	e.mustHave(a3, "aaa")
	e.mustNotExist(b1)

	// Artık b'nin tek kopyası var: o da silinmemeli.
	if rep := e.svc.Delete(e.ctx, []string{b2}, DeleteOptions{Permanent: true}, nil); rep.Results[0].Code != "last_copy" {
		t.Errorf("tek kalan kopya silindi: %+v", rep)
	}

	// Kalan kopya taramadan sonra değiştiyse güvenilir kopya sayılmamalı.
	future := time.Now().Add(time.Hour)
	os.Chtimes(a3, future, future)
	os.Remove(a2)
	if rep := e.svc.Delete(e.ctx, []string{a1}, DeleteOptions{Permanent: true}, nil); rep.Results[0].Code != "last_copy" {
		t.Errorf("değişmiş ya da silinmiş kopyaya güvenildi: %+v", rep)
	}
	e.mustHave(a1, "aaa")
}

func TestMoveSameVolume(t *testing.T) {
	e := setup(t)
	a := e.file("a.txt", "aaa")
	hedef := filepath.Join(e.dir, "hedef")
	os.Mkdir(hedef, 0o755)

	rep := e.svc.Move(e.ctx, []string{a}, hedef, ConflictSkip, nil)
	if rep.Succeeded != 1 {
		t.Fatalf("rapor: %+v", rep)
	}
	moved := filepath.Join(hedef, "a.txt")
	if rep.Results[0].NewPath != moved {
		t.Errorf("NewPath = %s", rep.Results[0].NewPath)
	}
	e.mustNotExist(a)
	e.mustHave(moved, "aaa")
	e.mustRecord(moved, "aaa")
}

func TestMoveConflict(t *testing.T) {
	e := setup(t)
	hedef := filepath.Join(e.dir, "hedef")
	mevcut := e.file(`hedef\a.txt`, "eski")
	e.file(`hedef\a (2).txt`, "eski2")
	a := e.file("a.txt", "yeni")

	// Skip: hedefe dokunulmamalı, kaynak yerinde kalmalı.
	rep := e.svc.Move(e.ctx, []string{a}, hedef, ConflictSkip, nil)
	if rep.Failed != 1 || !strings.Contains(rep.Results[0].Error, ErrExists.Error()) {
		t.Fatalf("rapor: %+v", rep)
	}
	e.mustHave(mevcut, "eski")
	e.mustHave(a, "yeni")

	// Rename: ilk boş ad "a (3).txt" olmalı.
	rep = e.svc.Move(e.ctx, []string{a}, hedef, ConflictRename, nil)
	want := filepath.Join(hedef, "a (3).txt")
	if rep.Succeeded != 1 || rep.Results[0].NewPath != want {
		t.Fatalf("rapor: %+v", rep)
	}
	e.mustHave(mevcut, "eski")
	e.mustHave(want, "yeni")
	e.mustRecord(want, "yeni")
	e.mustRecord(mevcut, "eski")
}

func TestMoveInvalidTargets(t *testing.T) {
	e := setup(t)
	a := e.file("a.txt", "aaa")

	for _, dest := range []string{e.dir, filepath.Join(e.dir, "yok"), a} {
		rep := e.svc.Move(e.ctx, []string{a}, dest, ConflictRename, nil)
		if rep.Failed != 1 {
			t.Errorf("hedef %s kabul edilmemeliydi: %+v", dest, rep)
		}
	}
	e.mustHave(a, "aaa")
}

func TestCopyThenRemove(t *testing.T) {
	e := setup(t)
	a := e.file("a.txt", "içerik")
	dest := filepath.Join(e.dir, "kopya.txt")

	if err := copyThenRemove(a, dest); err != nil {
		t.Fatal(err)
	}
	e.mustHave(dest, "içerik")
	if _, err := os.Lstat(a); !errors.Is(err, os.ErrNotExist) {
		t.Error("kaynak silinmemiş")
	}
	info, _ := os.Stat(dest)
	if !info.ModTime().Equal(e.past) {
		t.Errorf("değişiklik zamanı korunmamış: %v", info.ModTime())
	}

	// Hedef varsa: üzerine yazmamalı, kaynağa dokunmamalı.
	b := e.file("b.txt", "bbb")
	if err := copyThenRemove(b, dest); !errors.Is(err, ErrExists) {
		t.Errorf("err = %v, beklenen ErrExists", err)
	}
	e.mustHave(dest, "içerik")
	e.mustHave(b, "bbb")
}

func TestCopyThenRemoveRollsBackWhenSourceLocked(t *testing.T) {
	e := setup(t)
	a := e.file("a.txt", "kilitli")
	dest := filepath.Join(e.dir, "kopya.txt")

	// Windows'ta açık dosya silinemez; kopya geri alınmalı.
	f, err := os.Open(a)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := copyThenRemove(a, dest); err == nil {
		t.Skip("bu platformda açık dosya silinebiliyor")
	}
	if _, err := os.Lstat(dest); !errors.Is(err, os.ErrNotExist) {
		t.Error("başarısız taşımada kopya geri alınmamış")
	}
	e.mustHave(a, "kilitli")
}

func TestRename(t *testing.T) {
	e := setup(t)
	a := e.file("a.txt", "aaa")
	b := e.file("b.txt", "bbb")

	rep := e.svc.Rename(e.ctx, []RenameItem{
		{Path: a, NewName: "yeni ad.txt"},
		{Path: b, NewName: "yeni ad.txt"}, // az önce alınan ad: çakışma
		{Path: b, NewName: `alt\c.txt`},   // klasör ayracı yasak
	}, nil)
	if rep.Succeeded != 1 || rep.Failed != 2 {
		t.Fatalf("rapor: %+v", rep)
	}
	yeni := filepath.Join(e.dir, "yeni ad.txt")
	e.mustNotExist(a)
	e.mustHave(yeni, "aaa")
	e.mustRecord(yeni, "aaa")
	e.mustHave(b, "bbb")

	// Yalnızca harf büyüklüğü değişimi.
	rep = e.svc.Rename(e.ctx, []RenameItem{{Path: b, NewName: "B.txt"}}, nil)
	if rep.Succeeded != 1 {
		t.Fatalf("harf büyüklüğü değişimi: %+v", rep)
	}
	entries, _ := os.ReadDir(e.dir)
	var names []string
	for _, de := range entries {
		names = append(names, de.Name())
	}
	if !strings.Contains(strings.Join(names, "|"), "B.txt") {
		t.Errorf("diskteki adlar: %v", names)
	}
	if f, err := e.db.FileByPath(e.ctx, b); err != nil || f.Path != filepath.Join(e.dir, "B.txt") {
		t.Errorf("DB kaydı: %+v %v", f, err)
	}
}

func TestValidateName(t *testing.T) {
	for _, name := range []string{"a.txt", "çok güzel dosya (2).jpg", ".gitignore", "console.txt", "COM10.txt"} {
		if err := ValidateName(name); err != nil {
			t.Errorf("%q geçerli olmalı: %v", name, err)
		}
	}
	for _, name := range []string{"", ".", "..", "a/b", `a\b`, "a:b", "a?", "a*", `a"`, "a<", "a>", "a|",
		"son nokta.", "son boşluk ", "CON", "con.txt", "Lpt1.log", "nul", "a\tb"} {
		if err := ValidateName(name); !errors.Is(err, ErrInvalidName) {
			t.Errorf("%q geçersiz olmalı: %v", name, err)
		}
	}
}

func TestCancelled(t *testing.T) {
	e := setup(t)
	a := e.file("a.txt", "aaa")
	ctx, cancel := context.WithCancel(e.ctx)
	cancel()

	rep := e.svc.Delete(ctx, []string{a}, DeleteOptions{Permanent: true}, nil)
	if rep.Failed != 1 || rep.Results[0].Error != ErrCancelled.Error() {
		t.Fatalf("rapor: %+v", rep)
	}
	e.mustHave(a, "aaa")
}

// Benzer görsel silme: birebir kopyası olmayan bir görsel, eşik içinde benzer
// ve diskte duran bir eşi kaldıkça silinebilir.
func TestDeleteSimilarGuard(t *testing.T) {
	e := setup(t)
	img := func(rel, content, phash string) string {
		p := e.file(rel, content) // FullHash: "h-"+content (her biri benzersiz)
		f, _ := e.db.FileByPath(e.ctx, p)
		f.PHash = phash
		e.db.UpsertFiles(e.ctx, []db.File{f})
		return p
	}
	buyuk := img("buyuk.jpg", "1", "0000000000000000")
	kucuk := img("kucuk.jpg", "2", "0000000000000003") // 2 bit farklı
	uzak := img("uzak.jpg", "3", "00000000000000ff")   // 8 bit farklı
	similar := DeleteOptions{Permanent: true, Similar: true, Threshold: 4}

	// Birebir kopya korumasıyla (Similar olmadan) silinemez.
	if rep := e.svc.Delete(e.ctx, []string{kucuk}, DeleteOptions{Permanent: true}, nil); rep.Results[0].Code != "last_copy" {
		t.Fatalf("benzer mod olmadan silinmemeliydi: %+v", rep)
	}
	// Eşik dışındaki tek "eş" yeterli değil.
	if rep := e.svc.Delete(e.ctx, []string{uzak}, similar, nil); rep.Results[0].Code != "last_copy" {
		t.Errorf("eşik dışındaki görsel eş sayıldı: %+v", rep)
	}
	// İkisi birden silinmek istenirse hiçbiri silinmemeli.
	rep := e.svc.Delete(e.ctx, []string{buyuk, kucuk}, similar, nil)
	if rep.Succeeded != 0 {
		t.Errorf("benzer grubun tamamı silindi: %+v", rep)
	}
	// Yalnızca küçük olan: büyük kaldığı için silinir.
	rep = e.svc.Delete(e.ctx, []string{kucuk}, similar, nil)
	if rep.Succeeded != 1 {
		t.Fatalf("küçük görsel silinmedi: %+v", rep)
	}
	e.mustNotExist(kucuk)
	e.mustHave(buyuk, "1")
}

// Gerçek Geri Dönüşüm Kutusu'nu kullandığı için yalnızca istenince çalışır:
//
//	$env:DUPLICATEFILE_TEST_RECYCLE = "1"; go test ./internal/fileops -run Recycle
func TestRecycleBin(t *testing.T) {
	if os.Getenv("DUPLICATEFILE_TEST_RECYCLE") != "1" {
		t.Skip("DUPLICATEFILE_TEST_RECYCLE=1 değil")
	}
	e := setup(t)
	a := e.file("duplicatefile-test.txt", "geri dönüşüm testi")
	e.file(`kopya\duplicatefile-test.txt`, "geri dönüşüm testi")

	rep := e.svc.Delete(e.ctx, []string{a}, DeleteOptions{}, nil)
	if rep.Succeeded != 1 {
		t.Fatalf("rapor: %+v", rep)
	}
	e.mustNotExist(a)
}
