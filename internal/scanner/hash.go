package scanner

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/zeebo/xxh3"
)

const (
	// partialChunk, kısmi hash için dosyanın başından ve sonundan okunan bayt.
	// Boyutu 2*partialChunk veya daha küçük dosyaların tamamı okunur; bu
	// durumda kısmi hash aynı zamanda tam hash'tir.
	partialChunk = 4 << 10
	fullBufSize  = 1 << 20
)

var bufPool = sync.Pool{New: func() any { b := make([]byte, fullBufSize); return &b }}

func sumHex(h *xxh3.Hasher128) string {
	b := h.Sum128().Bytes()
	return hex.EncodeToString(b[:])
}

// partialHash, dosyanın ilk ve son partialChunk baytının hash'ini hesaplar.
func partialHash(e *entry) error {
	f, err := os.Open(e.path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := xxh3.New128()
	if e.size <= 2*partialChunk {
		buf := make([]byte, e.size)
		if _, err := io.ReadFull(f, buf); err != nil {
			return errChanged(err)
		}
		h.Write(buf)
		e.partial = sumHex(h)
		e.full = e.partial
		return nil
	}

	buf := make([]byte, partialChunk)
	if _, err := f.ReadAt(buf, 0); err != nil {
		return errChanged(err)
	}
	h.Write(buf)
	if _, err := f.ReadAt(buf, e.size-partialChunk); err != nil {
		return errChanged(err)
	}
	h.Write(buf)
	e.partial = sumHex(h)
	return nil
}

// fullHash, dosyanın tamamının hash'ini hesaplar. Okunan her parçada iptali
// kontrol eder ve ilerlemeyi progress ile bildirir.
func fullHash(ctx context.Context, e *entry, progress func(n int64)) error {
	if e.size <= 2*partialChunk && e.full != "" {
		return nil // küçük dosyada kısmi hash zaten tam hash
	}
	f, err := os.Open(e.path)
	if err != nil {
		return err
	}
	defer f.Close()

	bp := bufPool.Get().(*[]byte)
	defer bufPool.Put(bp)
	buf := *bp

	h := xxh3.New128()
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := f.Read(buf)
		if n > 0 {
			h.Write(buf[:n])
			total += int64(n)
			progress(int64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	if total != e.size {
		return errChanged(nil)
	}
	e.full = sumHex(h)
	return nil
}

// ErrChanged, dosyanın okunurken boyutunun değiştiğini bildirir.
var ErrChanged = errors.New("file changed during scan")

func errChanged(err error) error {
	if err == nil || err == io.EOF || err == io.ErrUnexpectedEOF {
		return ErrChanged
	}
	return err
}
