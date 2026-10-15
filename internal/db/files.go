package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// File, indekslenmiş bir dosyanın kaydıdır. Boş hash alanları "henüz
// hesaplanmadı" anlamına gelir ve veritabanında NULL olarak saklanır.
type File struct {
	ID           int64  `json:"id"`
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	ModTime      int64  `json:"modTime"` // Unix nanosaniye
	PartialHash  string `json:"-"`
	FullHash     string `json:"fullHash"`
	LastSeenScan int64  `json:"-"`
	// Görseller için (benzer görsel taraması hesaplar; diğer dosyalarda boş/0).
	PHash  string `json:"-"` // dHash, 16 karakter hex
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	// Similarity saklanmaz: benzer görsel grubunda grubun referans görseline
	// benzerlik yüzdesidir (100 = aynı).
	Similarity int `json:"similarity,omitempty"`
}

const fileColumns = `id, path, size, mtime, partial_hash, full_hash, last_seen_scan, phash, width, height`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanFile(r rowScanner) (File, error) {
	var f File
	var partial, full, phash sql.NullString
	err := r.Scan(&f.ID, &f.Path, &f.Size, &f.ModTime, &partial, &full, &f.LastSeenScan, &phash, &f.Width, &f.Height)
	f.PartialHash = partial.String
	f.FullHash = full.String
	f.PHash = phash.String
	return f, err
}

// FileByPath, yolu verilen dosyanın kaydını döner; yoksa ErrNotFound.
func (d *DB) FileByPath(ctx context.Context, path string) (File, error) {
	row := d.conn.QueryRowContext(ctx, `SELECT `+fileColumns+` FROM files WHERE path = ?`, path)
	f, err := scanFile(row)
	if errors.Is(err, sql.ErrNoRows) {
		return File{}, ErrNotFound
	}
	if err != nil {
		return File{}, fmt.Errorf("db: could not read file record: %w", err)
	}
	return f, nil
}

// FilesUnder, root klasörü altındaki (alt klasörler dahil) tüm kayıtları döner.
// Tarama başında önbellek olarak tek seferde yüklenmek içindir.
func (d *DB) FilesUnder(ctx context.Context, root string) ([]File, error) {
	lo, hi := pathRange(root)
	rows, err := d.conn.QueryContext(ctx,
		`SELECT `+fileColumns+` FROM files WHERE path >= ? AND path < ?`, lo, hi)
	if err != nil {
		return nil, fmt.Errorf("db: could not read file records: %w", err)
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		f, err := scanFile(rows)
		if err != nil {
			return nil, fmt.Errorf("db: could not read file row: %w", err)
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// UpsertFiles, kayıtları tek bir transaction içinde ekler ya da (aynı yol
// varsa) günceller. Tarama sonuçlarını toplu yazmak için kullanılır.
func (d *DB) UpsertFiles(ctx context.Context, files []File) error {
	if len(files) == 0 {
		return nil
	}
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO files (path, size, mtime, partial_hash, full_hash, last_seen_scan, phash, width, height)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			path           = excluded.path,
			size           = excluded.size,
			mtime          = excluded.mtime,
			partial_hash   = excluded.partial_hash,
			full_hash      = excluded.full_hash,
			last_seen_scan = excluded.last_seen_scan,
			phash          = excluded.phash,
			width          = excluded.width,
			height         = excluded.height`)
	if err != nil {
		return fmt.Errorf("db: could not prepare query: %w", err)
	}
	defer stmt.Close()

	for _, f := range files {
		if _, err := stmt.ExecContext(ctx, f.Path, f.Size, f.ModTime,
			nullString(f.PartialHash), nullString(f.FullHash), f.LastSeenScan,
			nullString(f.PHash), f.Width, f.Height); err != nil {
			return fmt.Errorf("db: could not write %s: %w", f.Path, err)
		}
	}
	return tx.Commit()
}

// DeleteFiles, verilen yolların kayıtlarını tek transaction içinde siler.
// Diskten kaybolmuş dosyaları temizlemek için kullanılır.
func (d *DB) DeleteFiles(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `DELETE FROM files WHERE path = ?`)
	if err != nil {
		return fmt.Errorf("db: could not prepare query: %w", err)
	}
	defer stmt.Close()

	for _, p := range paths {
		if _, err := stmt.ExecContext(ctx, p); err != nil {
			return fmt.Errorf("db: could not delete %s: %w", p, err)
		}
	}
	return tx.Commit()
}

// DeleteFile, tek bir dosyanın kaydını siler. Kayıt yoksa hata vermez.
func (d *DB) DeleteFile(ctx context.Context, path string) error {
	if _, err := d.conn.ExecContext(ctx, `DELETE FROM files WHERE path = ?`, path); err != nil {
		return fmt.Errorf("db: could not delete %s: %w", path, err)
	}
	return nil
}

// MoveFile, taşınan ya da yeniden adlandırılan dosyanın kaydını yeni yola
// aktarır. Hedef yolda eski bir kayıt varsa (üzerine yazılmış dosya) silinir.
// Kaynak kayıt yoksa ErrNotFound döner.
func (d *DB) MoveFile(ctx context.Context, oldPath, newPath string) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Yalnızca harf büyüklüğü değişiyorsa (a.txt -> A.txt) NOCASE nedeniyle
	// hedef kayıt kaynağın kendisidir; silinmemeli.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM files WHERE path = ? AND id <> (SELECT id FROM files WHERE path = ?)`,
		newPath, oldPath); err != nil {
		return fmt.Errorf("db: could not clear target record: %w", err)
	}
	res, err := tx.ExecContext(ctx, `UPDATE files SET path = ? WHERE path = ?`, newPath, oldPath)
	if err != nil {
		return fmt.Errorf("db: could not move %s: %w", oldPath, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}
