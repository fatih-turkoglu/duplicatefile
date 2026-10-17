package db

import (
	"context"
	"fmt"
	"runtime"
)

// migrations, şemanın sürüm sürüm değişiklikleridir. Veritabanının mevcut
// sürümü PRAGMA user_version'da tutulur; yalnızca eksik olanlar sırayla
// uygulanır. Var olan bir migration asla değiştirilmez, yenisi eklenir.
var migrations = []string{
	// v1: ilk şema
	`
	CREATE TABLE scans (
		id          INTEGER PRIMARY KEY,
		root_path   TEXT    NOT NULL,
		started_at  INTEGER NOT NULL,          -- Unix saniye
		finished_at INTEGER,                   -- Unix saniye, sürüyorsa NULL
		file_count  INTEGER NOT NULL DEFAULT 0,
		status      TEXT    NOT NULL DEFAULT 'running'
	);

	CREATE TABLE files (
		id             INTEGER PRIMARY KEY,
		path           TEXT    NOT NULL UNIQUE COLLATE NOCASE, -- Windows yolları büyük/küçük harf duyarsız
		size           INTEGER NOT NULL,
		mtime          INTEGER NOT NULL,       -- Unix nanosaniye
		partial_hash   TEXT,                   -- dosyanın başı + sonu, hex
		full_hash      TEXT,                   -- tüm içerik, hex
		last_seen_scan INTEGER NOT NULL DEFAULT 0
	);

	CREATE INDEX idx_files_size      ON files(size);
	CREATE INDEX idx_files_full_hash ON files(full_hash) WHERE full_hash IS NOT NULL;
	`,

	// v2: benzer görseller için algısal hash ve çözünürlük
	`
	ALTER TABLE files ADD COLUMN phash  TEXT;                       -- dHash, 16 karakter hex
	ALTER TABLE files ADD COLUMN width  INTEGER NOT NULL DEFAULT 0;
	ALTER TABLE files ADD COLUMN height INTEGER NOT NULL DEFAULT 0;
	`,

	// v3: Unix disks can contain separate files differing only in case.
	// Rebuild the table without losing IDs, cached hashes or scan history.
	`
	CREATE TABLE files_v3 (
		id INTEGER PRIMARY KEY,
		path TEXT NOT NULL UNIQUE COLLATE ` + pathCollation() + `,
		size INTEGER NOT NULL,
		mtime INTEGER NOT NULL,
		partial_hash TEXT,
		full_hash TEXT,
		last_seen_scan INTEGER NOT NULL DEFAULT 0,
		phash TEXT,
		width INTEGER NOT NULL DEFAULT 0,
		height INTEGER NOT NULL DEFAULT 0
	);
	INSERT INTO files_v3 SELECT id, path, size, mtime, partial_hash, full_hash,
		last_seen_scan, phash, width, height FROM files;
	DROP TABLE files;
	ALTER TABLE files_v3 RENAME TO files;
	CREATE INDEX idx_files_size ON files(size);
	CREATE INDEX idx_files_full_hash ON files(full_hash) WHERE full_hash IS NOT NULL;
	`,
}

func pathCollation() string {
	if runtime.GOOS == "windows" {
		return "NOCASE"
	}
	return "BINARY"
}

// SchemaVersion, bu sürümün beklediği şema sürümüdür.
func SchemaVersion() int {
	return len(migrations)
}

func (d *DB) migrate(ctx context.Context) error {
	var current int
	if err := d.conn.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("db: could not read schema version: %w", err)
	}
	if current > len(migrations) {
		return fmt.Errorf("db: database was created by a newer version (v%d, supported v%d)", current, len(migrations))
	}

	for v := current; v < len(migrations); v++ {
		tx, err := d.conn.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("db: could not start migration v%d: %w", v+1, err)
		}
		if _, err := tx.ExecContext(ctx, migrations[v]); err != nil {
			tx.Rollback()
			return fmt.Errorf("db: could not apply migration v%d: %w", v+1, err)
		}
		// PRAGMA parametre kabul etmez; v bizim sabitimiz olduğu için güvenli.
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", v+1)); err != nil {
			tx.Rollback()
			return fmt.Errorf("db: could not write schema version: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("db: could not commit migration v%d: %w", v+1, err)
		}
	}
	return nil
}
