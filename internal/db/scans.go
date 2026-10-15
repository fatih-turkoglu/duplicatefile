package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Tarama durumları.
const (
	ScanRunning   = "running"
	ScanCompleted = "completed"
	ScanCancelled = "cancelled"
	ScanFailed    = "failed"
)

// Scan, bir tarama çalıştırmasının kaydıdır.
type Scan struct {
	ID         int64
	RootPath   string
	StartedAt  time.Time
	FinishedAt time.Time // sürüyorsa sıfır değer
	FileCount  int64
	Status     string
}

// StartScan yeni bir tarama kaydı açar ve kimliğini döner. Kimlikler artan
// olduğu için dosyaların last_seen_scan alanında kullanılır.
func (d *DB) StartScan(ctx context.Context, root string) (int64, error) {
	res, err := d.conn.ExecContext(ctx,
		`INSERT INTO scans (root_path, started_at, status) VALUES (?, ?, ?)`,
		root, time.Now().Unix(), ScanRunning)
	if err != nil {
		return 0, fmt.Errorf("db: could not create scan record: %w", err)
	}
	return res.LastInsertId()
}

// FinishScan taramayı verilen durumla kapatır.
func (d *DB) FinishScan(ctx context.Context, id, fileCount int64, status string) error {
	_, err := d.conn.ExecContext(ctx,
		`UPDATE scans SET finished_at = ?, file_count = ?, status = ? WHERE id = ?`,
		time.Now().Unix(), fileCount, status, id)
	if err != nil {
		return fmt.Errorf("db: could not finish scan record: %w", err)
	}
	return nil
}

// ScanByID tek bir tarama kaydını döner; yoksa ErrNotFound.
func (d *DB) ScanByID(ctx context.Context, id int64) (Scan, error) {
	var s Scan
	var started int64
	var finished sql.NullInt64
	err := d.conn.QueryRowContext(ctx,
		`SELECT id, root_path, started_at, finished_at, file_count, status FROM scans WHERE id = ?`, id).
		Scan(&s.ID, &s.RootPath, &started, &finished, &s.FileCount, &s.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return Scan{}, ErrNotFound
	}
	if err != nil {
		return Scan{}, fmt.Errorf("db: could not read scan record: %w", err)
	}
	s.StartedAt = time.Unix(started, 0)
	if finished.Valid {
		s.FinishedAt = time.Unix(finished.Int64, 0)
	}
	return s, nil
}
