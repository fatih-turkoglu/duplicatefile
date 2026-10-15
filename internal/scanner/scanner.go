// Package scanner, klasörleri gezip kopya dosyaları bulur. Hash'ler üç
// aşamada ve yalnızca gerektiğinde hesaplanır:
//
//  1. Boyuta göre gruplama: boyutu benzersiz olan dosyanın kopyası olamaz.
//  2. Kısmi hash: aynı boyuttakilerin yalnızca ilk ve son birkaç KB'ı okunur.
//  3. Tam hash: kısmi hash'i de eşleşenlerin tamamı okunur.
//
// Yolu, boyutu ve değişiklik zamanı veritabanındakiyle aynı olan dosyaların
// hash'leri yeniden hesaplanmaz.
package scanner

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"DuplicateFile/internal/apperr"
	"DuplicateFile/internal/db"
	"DuplicateFile/internal/imaging"
)

// Aşama adları; ilerleme olaylarında kullanılır.
const (
	StageWalk    = "walk"
	StagePartial = "partial"
	StageFull    = "full"
	StageImage   = "image" // yalnızca benzer görsel modunda
	StageCleanup = "cleanup"
	StageSave    = "save"
)

// Tarama modları.
const (
	ModeExact   = "exact"   // içeriği birebir aynı dosyalar (varsayılan)
	ModeSimilar = "similar" // algısal hash'i yakın görseller
)

// DefaultThreshold, benzer görsellerde iki dHash arasında izin verilen en
// fazla farklı bit sayısıdır (64 üzerinden).
const DefaultThreshold = 8

// Her zaman atlanan Windows sistem klasörleri (küçük harfle).
var skipDirNames = map[string]bool{
	"$recycle.bin":              true,
	"system volume information": true,
}

const (
	maxErrors     = 200
	saveBatchSize = 5000
	progressEvery = 100 * time.Millisecond
)

// Options tarama ayarlarıdır.
type Options struct {
	Roots       []string `json:"roots"`
	MinSize     int64    `json:"minSize"`     // bayt; 0 baytlık dosyalar her zaman atlanır
	Extensions  []string `json:"extensions"`  // ".jpg" ya da "jpg"; boşsa tüm dosyalar
	ExcludeDirs []string `json:"excludeDirs"` // tam yol; alt klasörleriyle birlikte atlanır
	SkipHidden  bool     `json:"skipHidden"`  // gizli ve sistem dosyalarını atla
	Workers     int      `json:"workers"`     // paralel hash sayısı; 0 ise otomatik
	Mode        string   `json:"mode"`        // ModeExact (boşsa) ya da ModeSimilar
	Threshold   int      `json:"threshold"`   // benzer görsel eşiği; 0 ise DefaultThreshold
}

// Progress, taramanın anlık durumudur.
type Progress struct {
	Stage      string `json:"stage"`
	Done       int64  `json:"done"`
	Total      int64  `json:"total"` // gezinme aşamasında bilinmediği için 0
	Bytes      int64  `json:"bytes"`
	TotalBytes int64  `json:"totalBytes"`
	Current    string `json:"current"`
}

// ScanError, taramayı durdurmayan dosya düzeyindeki bir hatadır. Code
// bilinen hatalarda doludur (changed, image_unreadable, image_too_large);
// arayüz onu çevirir, yoksa Message (İngilizce/işletim sistemi metni) gösterilir.
type ScanError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func scanErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrChanged):
		return "changed"
	case errors.Is(err, imaging.ErrTooLarge):
		return "image_too_large"
	case errors.Is(err, imaging.ErrUnreadable):
		return "image_unreadable"
	}
	return ""
}

// Result taramanın özetidir.
type Result struct {
	ScanID     int64               `json:"scanId"`
	Status     string              `json:"status"` // db.ScanCompleted ya da db.ScanCancelled
	Mode       string              `json:"mode"`
	Threshold  int                 `json:"threshold"` // yalnızca benzer görsel modunda
	Roots      []string            `json:"roots"`
	FilesSeen  int64               `json:"filesSeen"`
	CacheHits  int64               `json:"cacheHits"` // hash'i önbellekten gelen dosya
	Hashed     int64               `json:"hashed"`    // bu taramada okunan dosya
	BytesRead  int64               `json:"bytesRead"`
	Removed    int64               `json:"removed"` // diskten silindiği için DB'den temizlenen
	Groups     []db.DuplicateGroup `json:"groups"`  // yalnızca tamamlanan taramada dolu
	Errors     []ScanError         `json:"errors"`  // en fazla maxErrors adet
	ErrorCount int                 `json:"errorCount"`
	DurationMs int64               `json:"durationMs"`
}

// entry, taramada görülen bir dosyadır. Önbellekten gelen bütün alanları
// taşır; böylece bir moddaki kayıt diğer modun hesapladığı hash'leri silmez.
type entry struct {
	path          string
	size, mtime   int64
	partial, full string
	phash         string
	width, height int
	read, failed  bool
	dirty         bool // DB'deki kayıttan farklı; kaydedilmeli
}

// Scanner, tarama sonuçlarını verilen veritabanında saklar.
type Scanner struct {
	db *db.DB
}

func New(d *db.DB) *Scanner {
	return &Scanner{db: d}
}

// run, tek bir taramanın durumunu taşır.
type run struct {
	ctx      context.Context
	opts     Options
	res      *Result
	similar  bool
	exts     map[string]bool
	excludes map[string]bool
	progress reporter
	errMu    sync.Mutex
}

// Scan taramayı çalıştırır. ctx iptal edilirse o ana kadar hesaplanan
// hash'ler kaydedilir ve Status "cancelled" olarak döner. Hata yalnızca
// taramanın hiç yapılamadığı durumlarda (geçersiz kök, DB hatası) döner.
func (s *Scanner) Scan(ctx context.Context, opts Options, onProgress func(Progress)) (*Result, error) {
	start := time.Now()
	roots, err := normalizeRoots(opts.Roots)
	if err != nil {
		return nil, err
	}
	if opts.Workers <= 0 {
		opts.Workers = min(runtime.NumCPU(), 4)
	}
	similar := opts.Mode == ModeSimilar
	if !similar {
		opts.Mode, opts.Threshold = ModeExact, 0
	} else if opts.Threshold <= 0 {
		opts.Threshold = DefaultThreshold
	}

	scanID, err := s.db.StartScan(ctx, strings.Join(roots, ";"))
	if err != nil {
		return nil, err
	}
	res := &Result{ScanID: scanID, Status: db.ScanCompleted, Mode: opts.Mode, Threshold: opts.Threshold,
		Roots: roots, Errors: []ScanError{}}
	r := &run{
		ctx:      ctx,
		opts:     opts,
		res:      res,
		similar:  similar,
		exts:     extSet(opts.Extensions),
		excludes: lowerSet(opts.ExcludeDirs),
		progress: reporter{fn: onProgress},
	}
	// Kayıt işlemleri iptalden etkilenmesin: iptalde de sonuçlar saklanır.
	dbCtx := context.WithoutCancel(ctx)

	fail := func(err error) (*Result, error) {
		s.db.FinishScan(dbCtx, scanID, res.FilesSeen, db.ScanFailed)
		return nil, err
	}

	cache, err := s.loadCache(ctx, roots)
	if err != nil && ctx.Err() == nil {
		return fail(err)
	}

	entries, err := r.walk(roots, cache)
	if err != nil && ctx.Err() == nil {
		return fail(err)
	}
	res.FilesSeen = int64(len(entries))

	if similar {
		if ctx.Err() == nil {
			r.hashImages(entries)
		}
	} else {
		if ctx.Err() == nil {
			r.hashPartial(entries)
		}
		if ctx.Err() == nil {
			r.hashFull(entries)
		}
	}
	if ctx.Err() == nil {
		// cache'te kalanlar bu taramada görülmeyen kayıtlardır.
		missing := r.findMissing(cache)
		if err := s.db.DeleteFiles(dbCtx, missing); err != nil {
			return fail(err)
		}
		res.Removed = int64(len(missing))
	}

	if err := s.save(dbCtx, r, entries); err != nil {
		return fail(err)
	}
	for _, e := range entries {
		if e.read {
			res.Hashed++
		}
	}

	if ctx.Err() != nil {
		res.Status = db.ScanCancelled
	}
	if err := s.db.FinishScan(dbCtx, scanID, res.FilesSeen, res.Status); err != nil {
		return nil, err
	}
	res.Groups = []db.DuplicateGroup{}
	if res.Status == db.ScanCompleted {
		// Gruplar DB'den değil bu taramada görülen dosyalardan kurulur; böylece
		// önceki taramalardan kalan, bu taramanın filtrelerine uymayan kayıtlar
		// sonuca karışmaz.
		if similar {
			res.Groups = buildSimilarGroups(entries, opts.Threshold, opts.Workers)
		} else {
			res.Groups = buildGroups(entries)
		}
	}
	res.DurationMs = time.Since(start).Milliseconds()
	return res, nil
}

// loadCache, köklerin altındaki tüm kayıtları küçük harfli yol anahtarıyla
// yükler. Kökler iç içe olmadığından kayıtlar çakışmaz.
func (s *Scanner) loadCache(ctx context.Context, roots []string) (map[string]db.File, error) {
	cache := make(map[string]db.File)
	for _, root := range roots {
		files, err := s.db.FilesUnder(ctx, root)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			cache[strings.ToLower(f.Path)] = f
		}
	}
	return cache, nil
}

// walk, kökleri gezip filtrelere uyan dosyaları toplar. Görülen her dosya
// (filtreye uymasa bile) cache'ten çıkarılır; böylece cache'te yalnızca bu
// taramada hiç görülmeyen kayıtlar kalır.
func (r *run) walk(roots []string, cache map[string]db.File) ([]*entry, error) {
	var entries []*entry
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if ctxErr := r.ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			if err != nil {
				if path == root {
					return apperr.Errorf("root_unreadable", "could not read %s: %w", root, err)
				}
				r.addErr(path, err)
				return nil
			}

			if d.IsDir() {
				if path != root && r.skipDir(path, d) {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return nil // kısayol bağlantıları, aygıtlar vb.
			}

			key := strings.ToLower(path)
			cached, hasCached := cache[key]
			delete(cache, key)

			if len(r.exts) > 0 && !r.exts[strings.ToLower(filepath.Ext(path))] {
				return nil
			}
			if r.similar && !imaging.Supported(path) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				r.addErr(path, err)
				return nil
			}
			if r.opts.SkipHidden && isHiddenOrSystem(info) {
				return nil
			}
			size := info.Size()
			if size == 0 || size < r.opts.MinSize {
				return nil
			}

			e := &entry{path: path, size: size, mtime: info.ModTime().UnixNano(), dirty: true}
			if hasCached && cached.Size == e.size && cached.ModTime == e.mtime {
				e.partial, e.full = cached.PartialHash, cached.FullHash
				e.phash, e.width, e.height = cached.PHash, cached.Width, cached.Height
				// Yolun harf büyüklüğü değiştiyse (yeniden adlandırma) kayıt güncellenmeli.
				e.dirty = cached.Path != path
				if (r.similar && e.phash != "") || (!r.similar && e.partial != "") {
					r.res.CacheHits++
				}
			}
			entries = append(entries, e)
			r.progress.send(Progress{Stage: StageWalk, Done: int64(len(entries)), Current: path}, false)
			return nil
		})
		if err != nil {
			return entries, err
		}
	}
	r.progress.send(Progress{Stage: StageWalk, Done: int64(len(entries)), Total: int64(len(entries))}, true)
	return entries, nil
}

func (r *run) skipDir(path string, d fs.DirEntry) bool {
	if skipDirNames[strings.ToLower(d.Name())] || r.excludes[strings.ToLower(path)] {
		return true
	}
	if r.opts.SkipHidden {
		if info, err := d.Info(); err == nil && isHiddenOrSystem(info) {
			return true
		}
	}
	return false
}

// hashPartial, boyutu başka bir dosyayla aynı olup kısmi hash'i olmayanları
// hash'ler.
func (r *run) hashPartial(entries []*entry) {
	var todo []*entry
	for _, group := range groupBy(entries, func(e *entry) string { return fmt.Sprint(e.size) }) {
		for _, e := range group {
			if e.partial == "" {
				todo = append(todo, e)
			}
		}
	}
	r.hashStage(StagePartial, todo,
		func(e *entry) int64 { return min(e.size, 2*partialChunk) },
		func(_ context.Context, e *entry, add func(int64)) error {
			if err := partialHash(e); err != nil {
				return err
			}
			add(min(e.size, 2*partialChunk))
			return nil
		})
}

// hashFull, boyutu ve kısmi hash'i başka bir dosyayla aynı olup tam hash'i
// olmayanları hash'ler.
func (r *run) hashFull(entries []*entry) {
	var todo []*entry
	key := func(e *entry) string {
		if e.partial == "" {
			return "" // kısmi hash'i yok (hata ya da iptal): gruplanmaz
		}
		return fmt.Sprint(e.size, ":", e.partial)
	}
	for _, group := range groupBy(entries, key) {
		for _, e := range group {
			if e.full == "" {
				todo = append(todo, e)
			}
		}
	}
	r.hashStage(StageFull, todo,
		func(e *entry) int64 { return e.size },
		func(ctx context.Context, e *entry, add func(int64)) error { return fullHash(ctx, e, add) })
}

// hashStage, items'ı opts.Workers kadar paralel işçiyle fn'den geçirir.
func (r *run) hashStage(stage string, items []*entry, bytesOf func(*entry) int64,
	fn func(context.Context, *entry, func(int64)) error) {

	var totalBytes int64
	for _, e := range items {
		totalBytes += bytesOf(e)
	}
	total := int64(len(items))
	var done, bytes atomic.Int64
	snapshot := func(current string) Progress {
		return Progress{Stage: stage, Done: done.Load(), Total: total,
			Bytes: bytes.Load(), TotalBytes: totalBytes, Current: current}
	}
	r.progress.send(snapshot(""), true)

	ch := make(chan *entry)
	var wg sync.WaitGroup
	for range r.opts.Workers {
		wg.Go(func() {
			for e := range ch {
				add := func(n int64) {
					bytes.Add(n)
					r.progress.send(snapshot(e.path), false)
				}
				err := fn(r.ctx, e, add)
				switch {
				case err == nil:
					e.read, e.dirty = true, true
				case r.ctx.Err() == nil:
					e.failed, e.dirty = true, true
					e.partial, e.full, e.phash = "", "", ""
					r.addErr(e.path, err)
				}
				done.Add(1)
				r.progress.send(snapshot(e.path), false)
			}
		})
	}

feed:
	for _, e := range items {
		select {
		case ch <- e:
		case <-r.ctx.Done():
			break feed
		}
	}
	close(ch)
	wg.Wait()

	r.res.BytesRead += bytes.Load()
	r.progress.send(snapshot(""), true)
}

// findMissing, cache'te kalan (taramada görülmeyen) kayıtlardan diskte artık
// bulunmayanların yollarını döner. Filtre ya da hariç tutma nedeniyle
// görülmeyen ama diskte duran dosyaların kayıtları korunur.
func (r *run) findMissing(cache map[string]db.File) []string {
	total := int64(len(cache))
	var done int64
	var missing []string
	r.progress.send(Progress{Stage: StageCleanup, Total: total}, true)
	for _, f := range cache {
		if r.ctx.Err() != nil {
			return nil
		}
		if _, err := os.Lstat(f.Path); errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, f.Path)
		}
		done++
		r.progress.send(Progress{Stage: StageCleanup, Done: done, Total: total, Current: f.Path}, false)
	}
	r.progress.send(Progress{Stage: StageCleanup, Done: done, Total: total}, true)
	return missing
}

// save, yalnızca yeni ya da değişmiş kayıtları yazar. Değişmemiş kayıtlara
// dokunulmadığı için last_seen_scan "son güncellendiği tarama" anlamına gelir.
func (s *Scanner) save(ctx context.Context, r *run, entries []*entry) error {
	var dirty []*entry
	for _, e := range entries {
		if e.dirty {
			dirty = append(dirty, e)
		}
	}
	total := int64(len(dirty))
	r.progress.send(Progress{Stage: StageSave, Total: total}, true)
	batch := make([]db.File, 0, min(len(dirty), saveBatchSize))
	for i, e := range dirty {
		batch = append(batch, db.File{
			Path: e.path, Size: e.size, ModTime: e.mtime,
			PartialHash: e.partial, FullHash: e.full, LastSeenScan: r.res.ScanID,
			PHash: e.phash, Width: e.width, Height: e.height,
		})
		if len(batch) == saveBatchSize || i == len(dirty)-1 {
			if err := s.db.UpsertFiles(ctx, batch); err != nil {
				return err
			}
			batch = batch[:0]
			r.progress.send(Progress{Stage: StageSave, Done: int64(i + 1), Total: total}, i == len(dirty)-1)
		}
	}
	return nil
}

// buildGroups, tam hash'i eşleşen dosyaları gruplar. Sıralama
// db.DuplicateGroups ile aynıdır: boyut büyükten küçüğe, sonra hash ve yol.
func buildGroups(entries []*entry) []db.DuplicateGroup {
	key := func(e *entry) string {
		if e.full == "" {
			return ""
		}
		return fmt.Sprint(e.size, ":", e.full)
	}
	groups := []db.DuplicateGroup{} // nil değil: JSON'da null yerine [] olsun
	for _, g := range groupBy(entries, key) {
		dg := db.DuplicateGroup{Hash: g[0].full, Size: g[0].size}
		for _, e := range g {
			dg.Files = append(dg.Files, db.File{Path: e.path, Size: e.size, ModTime: e.mtime, FullHash: e.full,
				Width: e.width, Height: e.height})
		}
		sort.Slice(dg.Files, func(i, j int) bool { return dg.Files[i].Path < dg.Files[j].Path })
		groups = append(groups, dg)
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Size != groups[j].Size {
			return groups[i].Size > groups[j].Size
		}
		return groups[i].Hash < groups[j].Hash
	})
	return groups
}

func (r *run) addErr(path string, err error) {
	r.errMu.Lock()
	defer r.errMu.Unlock()
	r.res.ErrorCount++
	if len(r.res.Errors) < maxErrors {
		r.res.Errors = append(r.res.Errors, ScanError{Path: path, Message: err.Error(), Code: scanErrorCode(err)})
	}
}

// groupBy, aynı anahtarlı en az iki eleman içeren grupları döner. Boş anahtar
// gruplanmaz; hatalı dosyalar atlanır.
func groupBy(entries []*entry, key func(*entry) string) [][]*entry {
	m := make(map[string][]*entry)
	for _, e := range entries {
		if e.failed {
			continue
		}
		if k := key(e); k != "" {
			m[k] = append(m[k], e)
		}
	}
	var groups [][]*entry
	for _, g := range m {
		if len(g) > 1 {
			groups = append(groups, g)
		}
	}
	return groups
}

// normalizeRoots, kökleri mutlak yola çevirir, klasör olduklarını doğrular ve
// başka bir kökün altında kalanları eler (aynı dosya iki kez taranmasın).
func normalizeRoots(roots []string) ([]string, error) {
	var abs []string
	for _, r := range roots {
		if strings.TrimSpace(r) == "" {
			continue
		}
		p, err := filepath.Abs(r)
		if err != nil {
			return nil, apperr.Errorf("root_invalid", "invalid folder %q: %w", r, err)
		}
		info, err := os.Stat(p)
		if err != nil {
			return nil, apperr.Errorf("root_not_found", "folder not found %q: %w", r, err)
		}
		if !info.IsDir() {
			return nil, apperr.Errorf("root_not_dir", "%q is not a folder", r)
		}
		abs = append(abs, p)
	}
	if len(abs) == 0 {
		return nil, apperr.New("no_roots", "no folder selected to scan")
	}

	sort.Slice(abs, func(i, j int) bool { return len(abs[i]) < len(abs[j]) })
	var out []string
	for _, p := range abs {
		nested := false
		for _, o := range out {
			if within(p, o) {
				nested = true
				break
			}
		}
		if !nested {
			out = append(out, p)
		}
	}
	return out, nil
}

// within, path'in root'un kendisi ya da altında olup olmadığını büyük/küçük
// harf duyarsız olarak söyler.
func within(path, root string) bool {
	path, root = strings.ToLower(path), strings.ToLower(root)
	if path == root {
		return true
	}
	sep := string(filepath.Separator)
	if !strings.HasSuffix(root, sep) {
		root += sep
	}
	return strings.HasPrefix(path, root)
}

func extSet(exts []string) map[string]bool {
	m := make(map[string]bool)
	for _, e := range exts {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		m[e] = true
	}
	return m
}

func lowerSet(paths []string) map[string]bool {
	m := make(map[string]bool)
	for _, p := range paths {
		if abs, err := filepath.Abs(p); err == nil {
			m[strings.ToLower(abs)] = true
		}
	}
	return m
}

// reporter, ilerleme bildirimlerini progressEvery aralığına seyreltir;
// force verilen bildirimler (aşama başı/sonu) her zaman iletilir.
type reporter struct {
	fn   func(Progress)
	mu   sync.Mutex
	last time.Time
}

func (p *reporter) send(pr Progress, force bool) {
	if p.fn == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	if !force && now.Sub(p.last) < progressEvery {
		return
	}
	p.last = now
	p.fn(pr)
}
