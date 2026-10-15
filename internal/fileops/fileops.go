// Package fileops, kopya dosyalar üzerinde silme, taşıma ve yeniden
// adlandırma işlemlerini yapar ve her işlemden sonra veritabanı kaydını
// günceller. Hiçbir işlem var olan bir dosyanın üzerine yazmaz.
package fileops

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"DuplicateFile/internal/db"
	"DuplicateFile/internal/imaging"
)

var (
	ErrExists       = errors.New("a file with the same name exists at the target")
	ErrInvalidName  = errors.New("invalid file name")
	ErrNoRecycleBin = errors.New("this drive has no Recycle Bin")
	ErrCancelled    = errors.New("operation cancelled")
	ErrLastCopy     = errors.New("not deleted: no other copy of this file would remain")
	ErrUnverified   = errors.New("could not verify that this file has a copy; rescan the folder")
	ErrSameFolder   = errors.New("file is already in the target folder")
	ErrNotRegular   = errors.New("only regular files can be processed")
	ErrDestInvalid  = errors.New("invalid target folder")
	ErrNotAbsolute  = errors.New("path must be absolute")
)

// Arayüzün hataları metinden değil koddan tanıması için.
var errorCodes = []struct {
	err  error
	code string
}{
	{ErrExists, "exists"},
	{ErrInvalidName, "invalid_name"},
	{ErrNoRecycleBin, "no_recycle_bin"},
	{ErrCancelled, "cancelled"},
	{ErrLastCopy, "last_copy"},
	{ErrUnverified, "unverified"},
	{ErrSameFolder, "same_folder"},
	{ErrNotRegular, "not_regular"},
	{ErrDestInvalid, "dest_invalid"},
	{ErrNotAbsolute, "path_not_absolute"},
	{fs.ErrNotExist, "not_found"},
	{fs.ErrPermission, "access_denied"},
}

func codeOf(err error) string {
	for _, c := range errorCodes {
		if errors.Is(err, c.err) {
			return c.code
		}
	}
	return "error"
}

// Conflict, taşımada hedefte aynı adlı dosya varsa ne yapılacağıdır.
type Conflict string

const (
	ConflictSkip   Conflict = "skip"   // dosyayı taşıma, hata olarak bildir
	ConflictRename Conflict = "rename" // "ad (2).uzantı" biçiminde yeni ad ver
)

// Result tek bir dosyanın işlem sonucudur.
type Result struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath,omitempty"`
	Error   string `json:"error,omitempty"`
	// Code, hatanın türüdür: exists, invalid_name, no_recycle_bin, cancelled,
	// last_copy, unverified ya da error. Başarılıysa boştur.
	Code string `json:"code,omitempty"`
	// Warning, dosya işlemi başarılı olduğu halde DB kaydı güncellenemediyse
	// doludur; kayıt bir sonraki taramada düzelir.
	Warning string `json:"warning,omitempty"`
}

func (r *Result) fail(err error) {
	r.Error = err.Error()
	r.Code = codeOf(err)
}

// Report toplu bir işlemin özetidir.
type Report struct {
	Succeeded int      `json:"succeeded"`
	Failed    int      `json:"failed"`
	Results   []Result `json:"results"`
}

func (r *Report) add(res Result) {
	if res.Error == "" {
		r.Succeeded++
	} else {
		r.Failed++
	}
	r.Results = append(r.Results, res)
}

// RenameItem, yeniden adlandırılacak dosya ve yeni adıdır (klasör olmadan).
type RenameItem struct {
	Path    string `json:"path"`
	NewName string `json:"newName"`
}

// Progress toplu işlemin ilerlemesini bildirir; nil olabilir.
type Progress func(done, total int)

// Service, dosya işlemlerini yapar ve sonuçları veritabanına yansıtır.
type Service struct {
	db *db.DB
}

func New(d *db.DB) *Service {
	return &Service{db: d}
}

// each, items'ın her biri için op'u çalıştırır; iptalde kalanları
// ErrCancelled ile işaretler.
func each[T any](ctx context.Context, items []T, progress Progress, op func(T) Result, pathOf func(T) string) Report {
	rep := Report{Results: make([]Result, 0, len(items))}
	for i, it := range items {
		if ctx.Err() != nil {
			res := Result{Path: pathOf(it)}
			res.fail(ErrCancelled)
			rep.add(res)
		} else {
			rep.add(op(it))
		}
		if progress != nil {
			progress(i+1, len(items))
		}
	}
	return rep
}

func self(p string) string { return p }

// DeleteOptions silme ayarlarıdır; sıfır değeri en güvenli olanıdır
// (Geri Dönüşüm Kutusu, yalnızca birebir kopya koruması).
type DeleteOptions struct {
	Permanent bool `json:"permanent"` // Geri Dönüşüm Kutusu'nu kullanma
	// Similar: dosyalar benzer görsel taramasından geliyor; birebir kopyası
	// yerine, algısal hash'i en fazla Threshold bit farklı bir görselin
	// kalması da yeterli sayılır.
	Similar   bool `json:"similar"`
	Threshold int  `json:"threshold"`
}

// Delete dosyaları Geri Dönüşüm Kutusu'na gönderir; opts.Permanent ise kalıcı
// siler. Diskte bir eşi (birebir kopyası ya da Similar ise benzer görseli)
// kalmayacak dosyalar silinmez (ErrLastCopy).
func (s *Service) Delete(ctx context.Context, paths []string, opts DeleteOptions, progress Progress) Report {
	dbCtx := context.WithoutCancel(ctx)
	permanent := opts.Permanent
	refused := s.lastCopyCheck(dbCtx, paths)
	if opts.Similar {
		refused = s.lastSimilarCheck(dbCtx, paths, refused, opts.Threshold)
	}
	return each(ctx, paths, progress, func(p string) Result {
		res := Result{Path: p}
		if err := refused[p]; err != nil {
			res.fail(err)
			return res
		}
		if err := checkFile(p); err != nil {
			res.fail(err)
			return res
		}
		var err error
		if permanent {
			err = os.Remove(p)
		} else {
			err = moveToRecycleBin(p)
		}
		if err != nil {
			res.fail(err)
			return res
		}
		if err := s.db.DeleteFile(dbCtx, p); err != nil {
			res.Warning = err.Error()
		}
		return res
	}, self)
}

// lastCopyCheck, silinmesi reddedilecek dosyaları nedeniyle döner:
//   - ErrUnverified: DB'de tam hash'i yok, kopya olduğu bilinmiyor
//   - ErrLastCopy: aynı içerikli (boyut + tam hash) dosyalardan silme listesinde
//     olmayan, diskte duran ve taramadan beri değişmemiş hiçbiri yok
//
// İkinci durumda o içeriğe ait bütün silme istekleri reddedilir; hangisinin
// tutulacağına kullanıcı karar versin diye biri kendiliğinden seçilmez.
func (s *Service) lastCopyCheck(ctx context.Context, paths []string) map[string]error {
	type content struct {
		size int64
		hash string
	}
	refused := make(map[string]error)
	deleting := make(map[string]bool, len(paths))
	for _, p := range paths {
		deleting[strings.ToLower(p)] = true
	}

	groups := make(map[content][]string)
	for _, p := range paths {
		f, err := s.db.FileByPath(ctx, p)
		if err != nil || f.FullHash == "" {
			refused[p] = ErrUnverified
			continue
		}
		k := content{f.Size, f.FullHash}
		groups[k] = append(groups[k], p)
	}

	for k, members := range groups {
		copies, err := s.db.FilesByContent(ctx, k.size, k.hash)
		survivor := false
		for _, c := range copies {
			if err != nil || deleting[strings.ToLower(c.Path)] {
				continue
			}
			info, statErr := os.Stat(c.Path)
			if statErr == nil && info.Mode().IsRegular() && info.Size() == k.size && info.ModTime().UnixNano() == c.ModTime {
				survivor = true
				break
			}
		}
		if !survivor {
			for _, p := range members {
				refused[p] = ErrLastCopy
			}
		}
	}
	return refused
}

// lastSimilarCheck, birebir kopya kontrolünden geçemeyen (refused) dosyalara
// ikinci bir şans verir: algısal hash'i en fazla threshold bit farklı, silme
// listesinde olmayan, diskte duran ve taramadan beri değişmemiş bir görsel
// varsa dosya silinebilir. Benzerlik geçişli olmadığından kontrol her dosya
// için ayrı yapılır.
func (s *Service) lastSimilarCheck(ctx context.Context, paths []string, refused map[string]error, threshold int) map[string]error {
	out := make(map[string]error)
	var pending []string
	for _, p := range paths {
		if refused[p] != nil {
			pending = append(pending, p)
		}
	}
	if len(pending) == 0 {
		return out
	}

	all, err := s.db.FilesWithPHash(ctx)
	if err != nil {
		for _, p := range pending {
			out[p] = ErrUnverified
		}
		return out
	}
	deleting := make(map[string]bool, len(paths))
	for _, p := range paths {
		deleting[strings.ToLower(p)] = true
	}
	type candidate struct {
		f    db.File
		hash uint64
	}
	var cands []candidate
	for _, f := range all {
		if deleting[strings.ToLower(f.Path)] {
			continue
		}
		if h, ok := imaging.ParseHash(f.PHash); ok {
			cands = append(cands, candidate{f, h})
		}
	}

	alive := make(map[string]bool) // diskte değişmeden duruyor mu (önbellekli)
	isAlive := func(f db.File) bool {
		v, ok := alive[f.Path]
		if !ok {
			info, err := os.Stat(f.Path)
			v = err == nil && info.Mode().IsRegular() && info.Size() == f.Size && info.ModTime().UnixNano() == f.ModTime
			alive[f.Path] = v
		}
		return v
	}

	for _, p := range pending {
		f, err := s.db.FileByPath(ctx, p)
		h, ok := imaging.ParseHash(f.PHash)
		if err != nil || !ok {
			out[p] = ErrUnverified
			continue
		}
		out[p] = ErrLastCopy
		for _, c := range cands {
			if imaging.Distance(h, c.hash) <= threshold && isAlive(c.f) {
				delete(out, p)
				break
			}
		}
	}
	return out
}

// Move dosyaları destDir klasörüne taşır.
func (s *Service) Move(ctx context.Context, paths []string, destDir string, conflict Conflict, progress Progress) Report {
	dbCtx := context.WithoutCancel(ctx)
	destErr := checkDir(destDir)
	return each(ctx, paths, progress, func(p string) Result {
		res := Result{Path: p}
		if destErr != nil {
			res.fail(destErr)
			return res
		}
		if err := checkFile(p); err != nil {
			res.fail(err)
			return res
		}
		dest := filepath.Join(destDir, filepath.Base(p))
		if strings.EqualFold(filepath.Clean(dest), filepath.Clean(p)) {
			res.fail(ErrSameFolder)
			return res
		}
		if exists(dest) {
			if conflict != ConflictRename {
				res.fail(ErrExists)
				return res
			}
			var err error
			if dest, err = uniquePath(dest); err != nil {
				res.fail(err)
				return res
			}
		}
		if err := moveFile(p, dest); err != nil {
			res.fail(err)
			return res
		}
		res.NewPath = dest
		res.Warning = s.recordMove(dbCtx, p, dest)
		return res
	}, self)
}

// Rename dosyaları aynı klasör içinde yeni adlarına taşır.
func (s *Service) Rename(ctx context.Context, items []RenameItem, progress Progress) Report {
	dbCtx := context.WithoutCancel(ctx)
	return each(ctx, items, progress, func(it RenameItem) Result {
		res := Result{Path: it.Path}
		if err := ValidateName(it.NewName); err != nil {
			res.fail(err)
			return res
		}
		if err := checkFile(it.Path); err != nil {
			res.fail(err)
			return res
		}
		dest := filepath.Join(filepath.Dir(it.Path), it.NewName)
		if dest == it.Path {
			res.NewPath = dest // ad değişmedi
			return res
		}
		// Yalnızca harf büyüklüğü değişiyorsa hedef dosya kaynağın kendisidir.
		if exists(dest) && !strings.EqualFold(dest, it.Path) {
			res.fail(ErrExists)
			return res
		}
		if err := os.Rename(it.Path, dest); err != nil {
			res.fail(err)
			return res
		}
		res.NewPath = dest
		res.Warning = s.recordMove(dbCtx, it.Path, dest)
		return res
	}, func(it RenameItem) string { return it.Path })
}

// recordMove DB kaydını yeni yola taşır. İndekslenmemiş dosya hata değildir.
func (s *Service) recordMove(ctx context.Context, from, to string) string {
	if err := s.db.MoveFile(ctx, from, to); err != nil && !errors.Is(err, db.ErrNotFound) {
		return err.Error()
	}
	return ""
}

// moveFile, aynı sürücüde yeniden adlandırır; farklı sürücüde kopyalayıp
// kaynağı siler. Çağıran, dest'in var olmadığını önceden doğrulamış olmalıdır.
func moveFile(src, dest string) error {
	err := os.Rename(src, dest)
	if err == nil || !isCrossDevice(err) {
		return err
	}
	return copyThenRemove(src, dest)
}

// copyThenRemove, src'yi dest'e kopyalar, boyutu doğrular, değişiklik
// zamanını korur (önbellek geçerli kalsın) ve ancak ondan sonra src'yi siler.
// Kopyalama yarıda kalırsa yarım dest silinir; src'ye dokunulmaz.
func copyThenRemove(src, dest string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrExists
		}
		return err
	}
	defer func() {
		if err != nil {
			out.Close()
			os.Remove(dest)
		}
	}()

	n, err := io.Copy(out, in)
	if err != nil {
		return err
	}
	if n != info.Size() {
		return fmt.Errorf("copy incomplete (%d/%d bytes)", n, info.Size())
	}
	if err = out.Sync(); err != nil {
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	if err = os.Chtimes(dest, time.Now(), info.ModTime()); err != nil {
		return err
	}

	in.Close()
	if rmErr := os.Remove(src); rmErr != nil {
		// Kaynak silinemediyse (örneğin başka bir programda açık) kopya da
		// silinir; böylece disk işlem öncesindeki hâline döner.
		return fmt.Errorf("source file could not be deleted; the move was rolled back: %w", rmErr)
	}
	return nil
}

// ValidateName, Windows dosya adı kurallarını denetler.
func ValidateName(name string) error {
	invalid := func(why string) error { return fmt.Errorf("%w: %s", ErrInvalidName, why) }
	if name == "" || name == "." || name == ".." {
		return invalid("name cannot be empty")
	}
	if strings.TrimRight(name, " .") != name {
		return invalid("name cannot end with a space or a dot")
	}
	for _, r := range name {
		if r < 32 || strings.ContainsRune(`<>:"/\|?*`, r) {
			return invalid(`name cannot contain: < > : " / \ | ? *`)
		}
	}
	stem := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	switch stem {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return invalid(stem + " is a reserved name in Windows")
	}
	return nil
}

// uniquePath, "ad.uzantı" var ise "ad (2).uzantı", "ad (3).uzantı"... içinden
// ilk boş olanı döner.
func uniquePath(p string) (string, error) {
	dir, base := filepath.Split(p)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	for i := 2; i < 10000; i++ {
		cand := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, i, ext))
		if !exists(cand) {
			return cand, nil
		}
	}
	return "", ErrExists
}

// exists, yolda bir şey olup olmadığını söyler. Belirsiz durumda (örneğin
// erişim hatası) üzerine yazmamak için true döner.
func exists(p string) bool {
	_, err := os.Lstat(p)
	return !errors.Is(err, os.ErrNotExist)
}

func checkFile(p string) error {
	if !filepath.IsAbs(p) {
		return fmt.Errorf("%w: %s", ErrNotAbsolute, p)
	}
	info, err := os.Lstat(p)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ErrNotRegular
	}
	return nil
}

// checkDir'in hataları ErrDestInvalid'i sarar; nedeni (ör. bulunamadı)
// açıklamada kalır, kod hep dest_invalid olur.
func checkDir(p string) error {
	if !filepath.IsAbs(p) {
		return fmt.Errorf("%w: not an absolute path: %s", ErrDestInvalid, p)
	}
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDestInvalid, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: not a folder", ErrDestInvalid)
	}
	return nil
}
