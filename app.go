package main

import (
	"context"
	"sync"

	"DuplicateFile/internal/apperr"
	"DuplicateFile/internal/db"
	"DuplicateFile/internal/fileops"
	"DuplicateFile/internal/scanner"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Arayüze gönderilen tarama olayları.
const (
	EventScanProgress = "scan:progress" // scanner.Progress
	EventScanDone     = "scan:done"     // scanner.Result
	EventScanError    = "scan:error"    // string
	EventOpsProgress  = "ops:progress"  // {done, total}
)

// App struct
type App struct {
	ctx context.Context

	// Veritabanı ayarlardan taşınabildiği için erişim kilitlidir.
	dbMu      sync.RWMutex
	db        *db.DB
	dbErr     error
	dbWarning string // ayarlanan konum açılamayıp varsayılana dönüldüyse

	// Tarama, dosya işlemleri ve veritabanı bakımı aynı anda çalışmaz.
	busyMu     sync.Mutex
	busy       bool
	scanCancel context.CancelFunc // tarama sürüyorsa nil değil
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.db, a.dbWarning, a.dbErr = openConfiguredDB(ctx)
	if a.dbErr != nil {
		// Uygulama yine açılsın; arayüz hatayı Ayarlar'da gösterir.
		runtime.LogErrorf(ctx, "veritabanı açılamadı: %v", a.dbErr)
		return
	}
	if a.dbWarning != "" {
		runtime.LogWarning(ctx, a.dbWarning)
	}
	runtime.LogInfof(ctx, "veritabanı: %s", a.db.Path())
}

// shutdown is called when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	a.CancelScan()
	a.dbMu.Lock()
	defer a.dbMu.Unlock()
	if a.db != nil {
		a.db.Close()
	}
}

func (a *App) getDB() (*db.DB, error) {
	a.dbMu.RLock()
	defer a.dbMu.RUnlock()
	if a.db == nil {
		if a.dbErr != nil {
			return nil, a.dbErr
		}
		return nil, apperr.New("db_unavailable", "database is not open")
	}
	return a.db, nil
}

var errBusy = apperr.New("busy", "another scan, file operation or database maintenance is in progress")

// acquire, meşgul bayrağını alır; bırakmak için dönen fonksiyon çağrılır.
func (a *App) acquire() (release func(), err error) {
	a.busyMu.Lock()
	defer a.busyMu.Unlock()
	if a.busy {
		return nil, errBusy
	}
	a.busy = true
	return func() {
		a.busyMu.Lock()
		a.busy = false
		a.busyMu.Unlock()
	}, nil
}

// SelectFolder Windows klasör seçme diyaloğunu açar. Kullanıcı vazgeçerse
// boş metin döner.
func (a *App) SelectFolder(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: title})
}

// ShowInExplorer Gezgin'i açıp dosyayı seçili gösterir.
func (a *App) ShowInExplorer(path string) error {
	return showInExplorer(path)
}

// StartScan taramayı arka planda başlatır ve hemen döner. İlerleme
// EventScanProgress, sonuç EventScanDone, hata EventScanError olayıyla gelir.
func (a *App) StartScan(opts scanner.Options) error {
	d, err := a.getDB()
	if err != nil {
		return err
	}
	release, err := a.acquire()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.busyMu.Lock()
	a.scanCancel = cancel
	a.busyMu.Unlock()

	go func() {
		defer func() {
			a.busyMu.Lock()
			a.scanCancel = nil
			a.busyMu.Unlock()
			cancel()
			release()
		}()
		res, err := scanner.New(d).Scan(ctx, opts, func(p scanner.Progress) {
			runtime.EventsEmit(a.ctx, EventScanProgress, p)
		})
		if err != nil {
			runtime.EventsEmit(a.ctx, EventScanError, err.Error())
			return
		}
		runtime.EventsEmit(a.ctx, EventScanDone, res)
	}()
	return nil
}

// CancelScan sürmekte olan taramayı iptal eder; tarama yoksa bir şey yapmaz.
func (a *App) CancelScan() {
	a.busyMu.Lock()
	defer a.busyMu.Unlock()
	if a.scanCancel != nil {
		a.scanCancel()
	}
}

// IsScanning bir taramanın sürüp sürmediğini döner.
func (a *App) IsScanning() bool {
	a.busyMu.Lock()
	defer a.busyMu.Unlock()
	return a.scanCancel != nil
}

// DuplicateGroups, veritabanında kayıtlı kopya gruplarını döner (örneğin
// uygulama yeniden açıldığında son sonuçları göstermek için).
func (a *App) DuplicateGroups(roots []string, minSize int64) ([]db.DuplicateGroup, error) {
	d, err := a.getDB()
	if err != nil {
		return nil, err
	}
	return d.DuplicateGroups(a.ctx, roots, minSize)
}

// runOps, dosya işlemini tarama ve diğer işlemlerle çakışmayacak şekilde
// çalıştırır; ilerlemeyi EventOpsProgress olayıyla bildirir.
func (a *App) runOps(op func(svc *fileops.Service, progress fileops.Progress) fileops.Report) (fileops.Report, error) {
	d, err := a.getDB()
	if err != nil {
		return fileops.Report{}, err
	}
	release, err := a.acquire()
	if err != nil {
		return fileops.Report{}, err
	}
	defer release()

	return op(fileops.New(d), func(done, total int) {
		runtime.EventsEmit(a.ctx, EventOpsProgress, map[string]int{"done": done, "total": total})
	}), nil
}

// DeleteFiles dosyaları Geri Dönüşüm Kutusu'na gönderir; opts.Permanent ise
// kalıcı siler. Benzer görsel sonuçlarında opts.Similar ve tarama eşiği verilir.
func (a *App) DeleteFiles(paths []string, opts fileops.DeleteOptions) (fileops.Report, error) {
	return a.runOps(func(svc *fileops.Service, p fileops.Progress) fileops.Report {
		return svc.Delete(a.ctx, paths, opts, p)
	})
}

// MoveFiles dosyaları destDir klasörüne taşır. conflict: "skip" ya da "rename".
// (Parametre düz string: Wails adlandırılmış string tiplerini TS'e aktarmıyor.)
func (a *App) MoveFiles(paths []string, destDir string, conflict string) (fileops.Report, error) {
	return a.runOps(func(svc *fileops.Service, p fileops.Progress) fileops.Report {
		return svc.Move(a.ctx, paths, destDir, fileops.Conflict(conflict), p)
	})
}

// RenameFiles dosyaları kendi klasörlerinde yeniden adlandırır.
func (a *App) RenameFiles(items []fileops.RenameItem) (fileops.Report, error) {
	return a.runOps(func(svc *fileops.Service, p fileops.Progress) fileops.Report {
		return svc.Rename(a.ctx, items, p)
	})
}

// ValidateFileName, arayüzün yeniden adlandırma kutusunda anlık doğrulama
// yapabilmesi içindir; geçerliyse boş metin döner.
func (a *App) ValidateFileName(name string) string {
	if err := fileops.ValidateName(name); err != nil {
		return err.Error()
	}
	return ""
}

// maintain, veritabanı bakım işini diğer işlerle çakışmadan çalıştırır.
func (a *App) maintain(fn func(d *db.DB) error) error {
	d, err := a.getDB()
	if err != nil {
		return err
	}
	release, err := a.acquire()
	if err != nil {
		return err
	}
	defer release()
	return fn(d)
}
