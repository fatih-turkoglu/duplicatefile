// Package db, taranan dosyaların bilgilerini (yol, boyut, değişiklik zamanı,
// hash) SQLite dosyasında saklar. Tekrarlı taramalarda değişmemiş dosyaların
// hash'i yeniden hesaplanmadan buradan okunur.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"DuplicateFile/internal/appdirs"

	_ "modernc.org/sqlite" // saf Go SQLite sürücüsü, CGO gerektirmez
)

// ErrNotFound, aranan kayıt veritabanında yoksa döner.
var ErrNotFound = errors.New("db: record not found")

// DB, uygulamanın dosya indeksi veritabanıdır.
type DB struct {
	conn *sql.DB
	path string
}

// DefaultPath, veritabanının varsayılan konumunu döner:
// Windows: %LOCALAPPDATA%\DuplicateFile\index.db
// macOS: ~/Library/Application Support/DuplicateFile/index.db
func DefaultPath() (string, error) {
	dir, err := appdirs.Data()
	if err != nil {
		return "", fmt.Errorf("db: user data folder not found: %w", err)
	}
	return filepath.Join(dir, "index.db"), nil
}

// Open, verilen yoldaki veritabanını açar (yoksa oluşturur) ve şemayı
// güncel sürüme taşır.
func Open(ctx context.Context, path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("db: could not create folder: %w", err)
	}

	conn, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("db: could not open: %w", err)
	}
	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("db: could not connect: %w", err)
	}

	d := &DB{conn: conn, path: path}
	if err := d.migrate(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return d, nil
}

// Close veritabanı bağlantısını kapatır.
func (d *DB) Close() error {
	return d.conn.Close()
}

// Path veritabanı dosyasının yolunu döner.
func (d *DB) Path() string {
	return d.path
}

// dsn, her bağlantıda uygulanacak PRAGMA ayarlarıyla bağlantı dizesini kurar.
//   - WAL: tarama yazarken arayüz okuyabilsin
//   - synchronous=NORMAL: WAL ile güvenli ve çok daha hızlı
//   - busy_timeout: eşzamanlı yazmalarda hemen hata vermek yerine beklesin
func dsn(path string) string {
	uriPath := strings.NewReplacer("%", "%25", "?", "%3f", "#", "%23").Replace(filepath.ToSlash(path))
	return "file:" + uriPath +
		"?_pragma=journal_mode(WAL)" +
		"&_pragma=synchronous(NORMAL)" +
		"&_pragma=busy_timeout(5000)" +
		"&_pragma=foreign_keys(ON)"
}

// pathRange, root klasörünün altındaki tüm yolları kapsayan [lo, hi) aralığını
// döner. Böylece LIKE yerine path indeksini kullanan bir aralık sorgusu yapılır.
// Ayraç '\' (0x5C) olduğu için üst sınır bir sonraki karakter ']' (0x5D) olur.
func pathRange(root string) (lo, hi string) {
	sep := string(filepath.Separator)
	root = filepath.Clean(root)
	if !strings.HasSuffix(root, sep) {
		root += sep
	}
	return root, root[:len(root)-1] + string(filepath.Separator+1)
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
