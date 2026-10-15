package db

import (
	"context"
	"fmt"
	"strings"
)

// DuplicateGroup, içeriği birebir aynı olan (boyut ve tam hash'i eşleşen)
// dosyalar kümesidir. Similar ise benzer görsel grubudur: dosyaların boyutu
// farklı olabilir, Size en büyük dosyanın boyutu, Hash referans görselin
// algısal hash'idir.
type DuplicateGroup struct {
	Hash    string `json:"hash"`
	Size    int64  `json:"size"`
	Files   []File `json:"files"`
	Similar bool   `json:"similar,omitempty"`
}

// FilesWithPHash, algısal hash'i hesaplanmış bütün kayıtları döner.
func (d *DB) FilesWithPHash(ctx context.Context) ([]File, error) {
	rows, err := d.conn.QueryContext(ctx, `SELECT `+fileColumns+` FROM files WHERE phash IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("db: could not read image records: %w", err)
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

// FilesByContent, boyutu ve tam hash'i verilenle aynı olan tüm kayıtları döner.
func (d *DB) FilesByContent(ctx context.Context, size int64, fullHash string) ([]File, error) {
	rows, err := d.conn.QueryContext(ctx,
		`SELECT `+fileColumns+` FROM files WHERE full_hash = ? AND size = ?`, fullHash, size)
	if err != nil {
		return nil, fmt.Errorf("db: could not read copies: %w", err)
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

// DuplicateGroups, verilen kök klasörlerin altındaki kopya gruplarını döner.
// Gruplar dosya boyutuna göre büyükten küçüğe sıralanır. minSize'dan küçük
// dosyalar dahil edilmez.
func (d *DB) DuplicateGroups(ctx context.Context, roots []string, minSize int64) ([]DuplicateGroup, error) {
	if len(roots) == 0 {
		return nil, nil
	}

	// Her kök için "path >= ? AND path < ?" koşulu, OR ile birleştirilir.
	conds := make([]string, len(roots))
	var rangeArgs []any
	for i, r := range roots {
		lo, hi := pathRange(r)
		conds[i] = "(path >= ? AND path < ?)"
		rangeArgs = append(rangeArgs, lo, hi)
	}
	inRoots := "(" + strings.Join(conds, " OR ") + ")"

	query := `
		SELECT ` + fileColumns + ` FROM files
		WHERE ` + inRoots + ` AND (size, full_hash) IN (
			SELECT size, full_hash FROM files
			WHERE full_hash IS NOT NULL AND size >= ? AND ` + inRoots + `
			GROUP BY size, full_hash
			HAVING COUNT(*) > 1
		)
		ORDER BY size DESC, full_hash, path`

	args := append([]any{}, rangeArgs...)
	args = append(args, minSize)
	args = append(args, rangeArgs...)

	rows, err := d.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db: could not read duplicate groups: %w", err)
	}
	defer rows.Close()

	groups := []DuplicateGroup{} // nil değil: JSON'da null yerine [] olsun
	for rows.Next() {
		f, err := scanFile(rows)
		if err != nil {
			return nil, fmt.Errorf("db: could not read file row: %w", err)
		}
		n := len(groups)
		if n == 0 || groups[n-1].Hash != f.FullHash || groups[n-1].Size != f.Size {
			groups = append(groups, DuplicateGroup{Hash: f.FullHash, Size: f.Size})
			n++
		}
		groups[n-1].Files = append(groups[n-1].Files, f)
	}
	return groups, rows.Err()
}
