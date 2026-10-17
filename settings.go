package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"DuplicateFile/internal/appdirs"
	"DuplicateFile/internal/apperr"
	"DuplicateFile/internal/db"
	"DuplicateFile/internal/pathutil"
)

// config, uygulama ayarlarının veritabanı dışında tutulması gerekenleridir
// (veritabanının kendi konumu gibi). Platformun kalıcı uygulama veri klasöründedir.
type config struct {
	DBPath string `json:"dbPath,omitempty"` // boşsa varsayılan konum
}

// defaultDBPath ve configFile değişkendir: testler gerçek kullanıcı
// klasörüne dokunmasın diye geçici yollara yönlendirir.
var defaultDBPath = func() (string, error) {
	if dir := os.Getenv(dataDirEnv); dir != "" {
		return filepath.Join(dir, "index.db"), nil
	}
	return db.DefaultPath()
}

var configFile = func() (string, error) {
	if dir := os.Getenv(dataDirEnv); dir != "" {
		return filepath.Join(dir, "config.json"), nil
	}
	dir, err := appdirs.Data()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// dataDirEnv tanımlıysa varsayılan veritabanı ve ayar dosyası bu klasörde
// tutulur. Uçtan uca testler gerçek kullanıcı verisine dokunmamak için
// kullanır; taşınabilir kullanım için de işe yarar.
const dataDirEnv = "DUPLICATEFILE_DATA_DIR"

func loadConfig() config {
	var c config
	if p, err := configFile(); err == nil {
		if data, err := os.ReadFile(p); err == nil {
			json.Unmarshal(data, &c) // bozuksa varsayılanlarla devam
		}
	}
	return c
}

func saveConfig(c config) error {
	p, err := configFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// openConfiguredDB, ayarlanan konumdaki veritabanını açar. Açılamazsa (ör.
// sürücü takılı değil) varsayılan konuma döner ve bir uyarı metni verir;
// ayar değiştirilmez, sürücü geri gelince yine kullanılır.
func openConfiguredDB(ctx context.Context) (d *db.DB, warning string, err error) {
	def, err := defaultDBPath()
	if err != nil {
		return nil, "", err
	}
	if p := loadConfig().DBPath; p != "" && !samePath(p, def) {
		if d, err := db.Open(ctx, p); err == nil {
			return d, "", nil
		} else {
			// Arayüz "ayarlanan konum açılamadı, varsayılan kullanılıyor" cümlesini
			// seçili dilde kurar; burada yalnızca neden verilir.
			warning = err.Error()
		}
	}
	d, err = db.Open(ctx, def)
	return d, warning, err
}

func samePath(a, b string) bool {
	return pathutil.Same(a, b)
}

// moveDatabase, açık veritabanını newPath'e taşır ve orada açılmış hâlini
// döner. Sıra: WAL ana dosyaya yazılır → kapatılır → kopyalanır → yeni kopya
// açılır → ancak o zaman eski dosyalar silinir. Herhangi bir adımda hata
// olursa yeni kopya silinir ve eski konumda yeniden açılmış veritabanı döner.
func moveDatabase(ctx context.Context, cur *db.DB, newPath string) (*db.DB, error) {
	oldPath := cur.Path()
	if samePath(oldPath, newPath) {
		return cur, nil
	}
	if _, err := os.Stat(newPath); err == nil {
		return cur, apperr.Errorf("db_target_exists", "a database already exists in the target folder: %s", newPath)
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return cur, fmt.Errorf("could not create target folder: %w", err)
	}
	if err := cur.Checkpoint(ctx); err != nil {
		return cur, err
	}
	if err := cur.Close(); err != nil {
		return cur, err
	}

	reopen := func(cause error) (*db.DB, error) {
		for _, p := range []string{newPath, newPath + "-wal", newPath + "-shm"} {
			os.Remove(p)
		}
		d, err := db.Open(ctx, oldPath)
		if err != nil {
			return nil, fmt.Errorf("%v; the old database could not be reopened either: %w", cause, err)
		}
		return d, cause
	}

	if err := copyFile(oldPath, newPath); err != nil {
		return reopen(fmt.Errorf("could not copy database: %w", err))
	}
	nd, err := db.Open(ctx, newPath)
	if err != nil {
		return reopen(fmt.Errorf("could not open database at the new location: %w", err))
	}
	for _, p := range []string{oldPath, oldPath + "-wal", oldPath + "-shm"} {
		os.Remove(p)
	}
	return nd, nil
}

// copyFile, yarım kalmış bir kopyanın hedef adında görünmemesi için önce
// geçici dosyaya yazar, diske işler ve sonra adını değiştirir.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Sync(); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

// ---------------------------------------------------------------- Arayüz

// DatabaseInfo, Ayarlar ekranındaki veritabanı bölümünün bilgisidir.
type DatabaseInfo struct {
	Path      string   `json:"path"`
	IsDefault bool     `json:"isDefault"`
	Warning   string   `json:"warning,omitempty"`
	Stats     db.Stats `json:"stats"`
}

// DatabaseInfo veritabanının konumunu ve istatistiklerini döner.
func (a *App) DatabaseInfo() (DatabaseInfo, error) {
	d, err := a.getDB()
	if err != nil {
		return DatabaseInfo{}, err
	}
	def, _ := defaultDBPath()
	stats, err := d.Stats(a.ctx)
	if err != nil {
		return DatabaseInfo{}, err
	}
	a.dbMu.RLock()
	warning := a.dbWarning
	a.dbMu.RUnlock()
	return DatabaseInfo{Path: d.Path(), IsDefault: samePath(d.Path(), def), Warning: warning, Stats: stats}, nil
}

// MoveDatabase veritabanını dir klasörüne taşır; dir boşsa varsayılan konuma.
func (a *App) MoveDatabase(dir string) (DatabaseInfo, error) {
	newPath, err := defaultDBPath()
	if err != nil {
		return DatabaseInfo{}, err
	}
	if dir != "" {
		if !filepath.IsAbs(dir) {
			return DatabaseInfo{}, apperr.New("path_not_absolute", "folder must be an absolute path")
		}
		newPath = filepath.Join(dir, "index.db")
	}

	err = a.maintain(func(cur *db.DB) error {
		a.dbMu.Lock()
		defer a.dbMu.Unlock()
		nd, err := moveDatabase(a.ctx, cur, newPath)
		a.db = nd // hata olsa da eski konumda yeniden açılmış olabilir
		if nd == nil {
			a.dbErr = err
			return err
		}
		if err != nil {
			return err
		}
		a.dbWarning = ""
		cfg := config{}
		if dir != "" {
			cfg.DBPath = newPath
		}
		if err := saveConfig(cfg); err != nil {
			return fmt.Errorf("database moved but the setting could not be saved; the old location will be used on next start: %w", err)
		}
		return nil
	})
	if err != nil {
		return DatabaseInfo{}, err
	}
	return a.DatabaseInfo()
}

// CompactResult sıkıştırmadan önceki ve sonraki boyuttur (bayt).
type CompactResult struct {
	Before int64 `json:"before"`
	After  int64 `json:"after"`
}

// CompactDatabase veritabanındaki boş alanı geri verir (VACUUM).
func (a *App) CompactDatabase() (CompactResult, error) {
	var r CompactResult
	err := a.maintain(func(d *db.DB) (err error) {
		r.Before, r.After, err = d.Vacuum(a.ctx)
		return err
	})
	return r, err
}

// PruneDatabase diskte artık olmayan dosyaların kayıtlarını siler.
func (a *App) PruneDatabase() (int64, error) {
	var n int64
	err := a.maintain(func(d *db.DB) (err error) {
		n, err = d.PruneMissing(a.ctx)
		return err
	})
	return n, err
}

// ClearDatabase bütün önbelleği siler; sonraki taramada dosyalar yeniden okunur.
func (a *App) ClearDatabase() error {
	return a.maintain(func(d *db.DB) error {
		if err := d.Clear(a.ctx); err != nil {
			return err
		}
		_, _, err := d.Vacuum(a.ctx)
		return err
	})
}
