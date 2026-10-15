package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"DuplicateFile/internal/db"
)

// isolate, varsayılan veritabanı ve ayar dosyası yollarını geçici klasöre
// yönlendirir; testler gerçek kullanıcı verisine dokunmaz.
func isolate(t *testing.T) (defPath string) {
	t.Helper()
	dir := t.TempDir()
	defPath = filepath.Join(dir, "varsayilan", "index.db")
	oldDef, oldCfg := defaultDBPath, configFile
	defaultDBPath = func() (string, error) { return defPath, nil }
	configFile = func() (string, error) { return filepath.Join(dir, "config.json"), nil }
	t.Cleanup(func() { defaultDBPath, configFile = oldDef, oldCfg })
	return defPath
}

func TestDataDirEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(dataDirEnv, dir)
	if p, _ := defaultDBPath(); p != filepath.Join(dir, "index.db") {
		t.Errorf("defaultDBPath = %s", p)
	}
	if p, _ := configFile(); p != filepath.Join(dir, "config.json") {
		t.Errorf("configFile = %s", p)
	}
}

func TestMoveDatabase(t *testing.T) {
	ctx := context.Background()
	defPath := isolate(t)
	cur, err := db.Open(ctx, defPath)
	if err != nil {
		t.Fatal(err)
	}
	cur.UpsertFiles(ctx, []db.File{{Path: `D:\a.txt`, Size: 5, FullHash: "h"}})

	newPath := filepath.Join(t.TempDir(), "yeni klasör", "index.db")
	nd, err := moveDatabase(ctx, cur, newPath)
	if err != nil {
		t.Fatal(err)
	}
	defer nd.Close()
	if nd.Path() != newPath {
		t.Errorf("yeni yol = %s", nd.Path())
	}
	if f, err := nd.FileByPath(ctx, `D:\a.txt`); err != nil || f.FullHash != "h" {
		t.Errorf("kayıt taşınmadı: %+v %v", f, err)
	}
	for _, p := range []string{defPath, defPath + "-wal", defPath + "-shm"} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("eski dosya kalmış: %s", p)
		}
	}
	if _, err := os.Stat(newPath + ".tmp"); err == nil {
		t.Error("geçici kopya kalmış")
	}
}

func TestMoveDatabaseRefusesExistingTarget(t *testing.T) {
	ctx := context.Background()
	defPath := isolate(t)
	cur, _ := db.Open(ctx, defPath)
	defer cur.Close()

	target := filepath.Join(t.TempDir(), "index.db")
	os.WriteFile(target, []byte("başka bir veritabanı"), 0o644)
	nd, err := moveDatabase(ctx, cur, target)
	if err == nil {
		t.Fatal("var olan hedefin üzerine yazılmamalıydı")
	}
	if nd != cur {
		t.Error("hata durumunda mevcut veritabanı dönmeliydi")
	}
	if data, _ := os.ReadFile(target); string(data) != "başka bir veritabanı" {
		t.Error("hedef dosya değiştirilmiş")
	}
}

func TestMoveDatabaseRollsBack(t *testing.T) {
	ctx := context.Background()
	defPath := isolate(t)
	cur, _ := db.Open(ctx, defPath)
	cur.UpsertFiles(ctx, []db.File{{Path: `D:\a.txt`, Size: 1}})

	// Hedef klasör yerinde bir dosya var: klasör oluşturulamaz.
	blocker := filepath.Join(t.TempDir(), "engel")
	os.WriteFile(blocker, nil, 0o644)
	nd, err := moveDatabase(ctx, cur, filepath.Join(blocker, "index.db"))
	if err == nil {
		t.Fatal("hata bekleniyordu")
	}
	defer nd.Close()
	if _, err := nd.FileByPath(ctx, `D:\a.txt`); err != nil {
		t.Errorf("eski veritabanı kullanılamıyor: %v", err)
	}
}

func TestOpenConfiguredDBFallsBack(t *testing.T) {
	ctx := context.Background()
	defPath := isolate(t)

	// Ayar yokken varsayılan.
	d, warning, err := openConfiguredDB(ctx)
	if err != nil || d.Path() != defPath || warning != "" {
		t.Fatalf("varsayılan: %v %q %v", d, warning, err)
	}
	d.Close()

	// Ayarlanan konum geçerliyse o.
	custom := filepath.Join(t.TempDir(), "index.db")
	saveConfig(config{DBPath: custom})
	d, warning, err = openConfiguredDB(ctx)
	if err != nil || d.Path() != custom || warning != "" {
		t.Fatalf("ayarlı: %v %q %v", d, warning, err)
	}
	d.Close()

	// Ayarlanan konum açılamıyorsa varsayılana dönülür, uyarı verilir, ayar korunur.
	blocker := filepath.Join(t.TempDir(), "engel")
	os.WriteFile(blocker, nil, 0o644)
	saveConfig(config{DBPath: filepath.Join(blocker, "index.db")})
	d, warning, err = openConfiguredDB(ctx)
	if err != nil || d.Path() != defPath || warning == "" {
		t.Fatalf("geri dönüş: %v %q %v", d, warning, err)
	}
	d.Close()
	if loadConfig().DBPath == "" {
		t.Error("ayar silinmemeliydi (sürücü geri gelince kullanılsın)")
	}
}
