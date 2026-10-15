package db

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// Stats veritabanının özetidir.
type Stats struct {
	Files     int64 `json:"files"`
	Scans     int64 `json:"scans"`
	SizeBytes int64 `json:"sizeBytes"` // ana dosya + WAL + SHM
}

// Stats kayıt sayılarını ve diskteki boyutu döner.
func (d *DB) Stats(ctx context.Context) (Stats, error) {
	var s Stats
	if err := d.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM files`).Scan(&s.Files); err != nil {
		return s, fmt.Errorf("db: could not count records: %w", err)
	}
	if err := d.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM scans`).Scan(&s.Scans); err != nil {
		return s, fmt.Errorf("db: could not count scans: %w", err)
	}
	s.SizeBytes = d.SizeOnDisk()
	return s, nil
}

// SizeOnDisk, veritabanının diskte kapladığı toplam bayttır.
func (d *DB) SizeOnDisk() int64 {
	var total int64
	for _, p := range d.Files() {
		if info, err := os.Stat(p); err == nil {
			total += info.Size()
		}
	}
	return total
}

// Files, veritabanına ait dosyaların yollarıdır (ana dosya, WAL, SHM).
func (d *DB) Files() []string {
	return []string{d.path, d.path + "-wal", d.path + "-shm"}
}

// Checkpoint, WAL'daki değişiklikleri ana dosyaya yazar ve WAL'ı boşaltır;
// böylece veritabanı tek dosya olarak kopyalanabilir.
func (d *DB) Checkpoint(ctx context.Context) error {
	if _, err := d.conn.ExecContext(ctx, `PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		return fmt.Errorf("db: checkpoint failed: %w", err)
	}
	return nil
}

// Vacuum, silinen kayıtlardan kalan boş alanı geri verir ve diskteki boyutun
// öncesini ve sonrasını döner.
func (d *DB) Vacuum(ctx context.Context) (before, after int64, err error) {
	before = d.SizeOnDisk()
	if _, err = d.conn.ExecContext(ctx, `VACUUM`); err != nil {
		return before, before, fmt.Errorf("db: vacuum failed: %w", err)
	}
	if err = d.Checkpoint(ctx); err != nil {
		return before, d.SizeOnDisk(), err
	}
	return before, d.SizeOnDisk(), nil
}

// Clear bütün dosya ve tarama kayıtlarını siler (önbelleği sıfırlar). Şema
// korunur.
func (d *DB) Clear(ctx context.Context) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: could not begin transaction: %w", err)
	}
	defer tx.Rollback()
	for _, q := range []string{`DELETE FROM files`, `DELETE FROM scans`} {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("db: could not clear: %w", err)
		}
	}
	return tx.Commit()
}

// PruneMissing, diskte artık bulunmayan dosyaların kayıtlarını siler ve
// silinen sayıyı döner. Erişilemeyen (ör. bağlı olmayan sürücü) dosyaların
// kayıtları silinmez; yalnızca "yok" yanıtı alınanlar silinir.
func (d *DB) PruneMissing(ctx context.Context) (int64, error) {
	rows, err := d.conn.QueryContext(ctx, `SELECT path FROM files`)
	if err != nil {
		return 0, fmt.Errorf("db: could not read records: %w", err)
	}
	var missing []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			rows.Close()
			return 0, err
		}
		if _, err := os.Lstat(p); errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, p)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return int64(len(missing)), d.DeleteFiles(ctx, missing)
}
