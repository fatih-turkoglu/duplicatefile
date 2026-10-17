package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"DuplicateFile/internal/db"
)

func openDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func write(t *testing.T, root, rel string, data []byte) string {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(strings.ReplaceAll(rel, `\`, "/")))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// bigFile, partialChunk'tan büyük; middle yalnızca ortadaki baytı değiştirir.
func bigFile(middle byte) []byte {
	b := bytes.Repeat([]byte("0123456789"), 2000) // 20000 bayt
	b[len(b)/2] = middle
	return b
}

// fixture, test klasör ağacını kurar.
//
//	hello x3       -> küçük kopya grubu
//	big1, big2     -> büyük kopya grubu
//	big3           -> big'lerle aynı baş/son, farklı orta: kısmi eşleşir, tam eşleşmez
//	s1, s2         -> aynı boyut, farklı içerik
//	u.txt          -> benzersiz boyut, hiç okunmamalı
//	empty1/2       -> 0 bayt, atlanmalı
//	$RECYCLE.BIN, haric -> atlanmalı
func fixture(t *testing.T) (root string, p map[string]string) {
	root = t.TempDir()
	hello := []byte("hello world")
	p = map[string]string{
		"dup1": write(t, root, `a\dup1.txt`, hello),
		"dup2": write(t, root, `b\dup2.txt`, hello),
		"dup3": write(t, root, `b\alt\dup3.txt`, hello),
		"big1": write(t, root, `big1.bin`, bigFile('A')),
		"big2": write(t, root, `c\big2.bin`, bigFile('A')),
		"big3": write(t, root, `c\big3.bin`, bigFile('B')),
		"s1":   write(t, root, `s1.txt`, []byte("aaaa")),
		"s2":   write(t, root, `s2.txt`, []byte("bbbb")),
		"u":    write(t, root, `u.txt`, []byte("benzersiz")),
	}
	write(t, root, `empty1.txt`, nil)
	write(t, root, `a\empty2.txt`, nil)
	write(t, root, `$RECYCLE.BIN\dup.txt`, hello)
	p["haric"] = filepath.Join(root, "haric")
	write(t, root, `haric\dup.txt`, hello)
	return root, p
}

func scan(t *testing.T, s *Scanner, opts Options) *Result {
	t.Helper()
	res, err := s.Scan(context.Background(), opts, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	return res
}

func groupPaths(res *Result) [][]string {
	var out [][]string
	for _, g := range res.Groups {
		var ps []string
		for _, f := range g.Files {
			ps = append(ps, f.Path)
		}
		out = append(out, ps)
	}
	return out
}

func checkGroups(t *testing.T, res *Result, want [][]string) {
	t.Helper()
	got := groupPaths(res)
	if len(got) != len(want) {
		t.Fatalf("gruplar = %v\nbeklenen  = %v", got, want)
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("grup %d = %v, beklenen %v", i, got[i], want[i])
		}
		for j := range want[i] {
			if !within(got[i][j], want[i][j]) || !within(want[i][j], got[i][j]) {
				t.Fatalf("grup %d = %v, beklenen %v", i, got[i], want[i])
			}
		}
	}
}

func TestScanFindsDuplicatesAndUsesCache(t *testing.T) {
	root, p := fixture(t)
	s := New(openDB(t))
	opts := Options{Roots: []string{root}, ExcludeDirs: []string{p["haric"]}}

	// İlk tarama: her şey okunur.
	res := scan(t, s, opts)
	if res.Status != db.ScanCompleted {
		t.Fatalf("durum = %s", res.Status)
	}
	if res.ErrorCount != 0 {
		t.Fatalf("hatalar: %+v", res.Errors)
	}
	if res.FilesSeen != 9 {
		t.Errorf("FilesSeen = %d, beklenen 9", res.FilesSeen)
	}
	// Okunanlar: hello x3, s1, s2, big x3 (u.txt benzersiz boyutlu, okunmaz)
	if res.Hashed != 8 || res.CacheHits != 0 {
		t.Errorf("Hashed = %d, CacheHits = %d; beklenen 8, 0", res.Hashed, res.CacheHits)
	}
	// Gruplar boyuta göre büyükten küçüğe, içerde yola göre sıralı.
	checkGroups(t, res, [][]string{
		{p["big1"], p["big2"]},
		{p["dup1"], p["dup3"], p["dup2"]},
	})

	// İkinci tarama: hiçbir dosya okunmamalı.
	res = scan(t, s, opts)
	if res.Hashed != 0 || res.CacheHits != 8 || res.BytesRead != 0 {
		t.Errorf("Hashed = %d, CacheHits = %d, BytesRead = %d; beklenen 0, 8, 0",
			res.Hashed, res.CacheHits, res.BytesRead)
	}
	if len(res.Groups) != 2 {
		t.Errorf("grup sayısı = %d, beklenen 2", len(res.Groups))
	}

	// big3 değişip big1'in kopyası olursa yalnızca o yeniden okunmalı.
	os.WriteFile(p["big3"], bigFile('A'), 0o644)
	future := time.Now().Add(time.Hour)
	os.Chtimes(p["big3"], future, future)
	res = scan(t, s, opts)
	if res.Hashed != 1 {
		t.Errorf("değişiklik sonrası Hashed = %d, beklenen 1", res.Hashed)
	}
	checkGroups(t, res, [][]string{
		{p["big1"], p["big2"], p["big3"]},
		{p["dup1"], p["dup3"], p["dup2"]},
	})

	// Silinen dosyanın kaydı temizlenmeli.
	os.Remove(p["dup2"])
	res = scan(t, s, opts)
	if res.Removed != 1 {
		t.Errorf("Removed = %d, beklenen 1", res.Removed)
	}
	if _, err := s.db.FileByPath(context.Background(), p["dup2"]); !errors.Is(err, db.ErrNotFound) {
		t.Errorf("silinen dosyanın kaydı duruyor: %v", err)
	}
	checkGroups(t, res, [][]string{
		{p["big1"], p["big2"], p["big3"]},
		{p["dup1"], p["dup3"]},
	})
}

func TestScanFilterKeepsCache(t *testing.T) {
	root, p := fixture(t)
	s := New(openDB(t))
	scan(t, s, Options{Roots: []string{root}})

	// Yalnızca .bin taraması .txt kayıtlarını silmemeli.
	res := scan(t, s, Options{Roots: []string{root}, Extensions: []string{"BIN"}})
	if res.Removed != 0 {
		t.Errorf("Removed = %d, beklenen 0", res.Removed)
	}
	checkGroups(t, res, [][]string{{p["big1"], p["big2"]}})
	if f, err := s.db.FileByPath(context.Background(), p["dup1"]); err != nil || f.FullHash == "" {
		t.Errorf("filtre dışı kaydın önbelleği kaybolmuş: %+v %v", f, err)
	}
}

func TestScanMinSize(t *testing.T) {
	root, p := fixture(t)
	s := New(openDB(t))
	res := scan(t, s, Options{Roots: []string{root}, MinSize: 100})
	checkGroups(t, res, [][]string{{p["big1"], p["big2"]}})
}

func TestScanNestedRoots(t *testing.T) {
	root, _ := fixture(t)
	s := New(openDB(t))
	res := scan(t, s, Options{Roots: []string{filepath.Join(root, "b"), root, root + string(filepath.Separator)}})
	if len(res.Roots) != 1 {
		t.Errorf("Roots = %v, iç içe kökler elenmeliydi", res.Roots)
	}
}

// Kopyası olmayan klasörde gruplar JSON'da null değil [] olmalı; arayüz
// null'da çöküyordu (kopyalar silindikten sonra yeniden tarama).
func TestScanNoDuplicatesGivesEmptyGroups(t *testing.T) {
	root := t.TempDir()
	write(t, root, "a.txt", []byte("bir"))
	write(t, root, "b.txt", []byte("iki"))
	s := New(openDB(t))
	for _, mode := range []string{ModeExact, ModeSimilar} {
		res := scan(t, s, Options{Roots: []string{root}, Mode: mode})
		data, _ := json.Marshal(res)
		if !strings.Contains(string(data), `"groups":[]`) {
			t.Errorf("%s modu: groups null olmamalı: %s", mode, data)
		}
	}
	if g, _ := s.db.DuplicateGroups(context.Background(), []string{root}, 0); g == nil {
		t.Error("db.DuplicateGroups nil döndü")
	}
}

func TestScanInvalidRoot(t *testing.T) {
	s := New(openDB(t))
	if _, err := s.Scan(context.Background(), Options{Roots: []string{`Z:\kesinlikle\yok`}}, nil); err == nil {
		t.Error("olmayan kök hata vermeliydi")
	}
	if _, err := s.Scan(context.Background(), Options{}, nil); err == nil {
		t.Error("boş kök listesi hata vermeliydi")
	}
}

func TestScanCancel(t *testing.T) {
	root, _ := fixture(t)
	s := New(openDB(t))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stages []string
	res, err := s.Scan(ctx, Options{Roots: []string{root}}, func(p Progress) {
		if len(stages) == 0 || stages[len(stages)-1] != p.Stage {
			stages = append(stages, p.Stage)
		}
		if p.Stage == StagePartial {
			cancel()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != db.ScanCancelled {
		t.Errorf("durum = %s, beklenen cancelled", res.Status)
	}
	if len(res.Groups) != 0 {
		t.Errorf("iptal edilen taramada grup dönmemeli: %d", len(res.Groups))
	}
	if stages[len(stages)-1] != StageSave {
		t.Errorf("iptalden sonra da kayıt yapılmalı; aşamalar: %v", stages)
	}
	sc, _ := s.db.ScanByID(context.Background(), res.ScanID)
	if sc.Status != db.ScanCancelled {
		t.Errorf("DB'deki durum = %s", sc.Status)
	}
}

func TestProgressStages(t *testing.T) {
	root, _ := fixture(t)
	s := New(openDB(t))
	var stages []string
	_, err := s.Scan(context.Background(), Options{Roots: []string{root}}, func(p Progress) {
		if len(stages) == 0 || stages[len(stages)-1] != p.Stage {
			stages = append(stages, p.Stage)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{StageWalk, StagePartial, StageFull, StageCleanup, StageSave}
	if len(stages) != len(want) {
		t.Fatalf("aşamalar = %v, beklenen %v", stages, want)
	}
	for i := range want {
		if stages[i] != want[i] {
			t.Fatalf("aşamalar = %v, beklenen %v", stages, want)
		}
	}
}
