package main

import (
	"container/list"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"DuplicateFile/internal/imaging"

	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// thumbPath, arayüzün küçük resim istediği adres: /thumb?path=<tam yol>&s=<piksel>
const thumbPath = "/thumb"

// thumbMiddleware, /thumb isteklerini Vite'a ya da gömülü dosyalara gitmeden
// karşılar; diğer bütün istekleri olduğu gibi geçirir.
func thumbMiddleware(cache *thumbCache) assetserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != thumbPath {
				next.ServeHTTP(w, r)
				return
			}
			serveThumb(w, r, cache)
		})
	}
}

func serveThumb(w http.ResponseWriter, r *http.Request, cache *thumbCache) {
	path := r.URL.Query().Get("path")
	size, _ := strconv.Atoi(r.URL.Query().Get("s"))
	size = min(max(size, 32), 512)

	// Yalnızca desteklenen görsel uzantılı tam yollar; başka dosyalar okunmaz.
	if !filepath.IsAbs(path) || !imaging.Supported(path) {
		http.Error(w, "desteklenmeyen dosya", http.StatusBadRequest)
		return
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}

	// Dosya değişirse anahtar değişir; eski küçük resim kullanılmaz.
	key := fmt.Sprintf("%s|%d|%d|%d", path, info.Size(), info.ModTime().UnixNano(), size)
	data, ok := cache.get(key)
	if !ok {
		data, err = imaging.Thumbnail(path, size)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		cache.put(key, data)
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.Write(data)
}

// thumbCache, son üretilen küçük resimleri toplam bayt sınırıyla bellekte
// tutar (en uzun süredir kullanılmayan önce atılır).
type thumbCache struct {
	mu       sync.Mutex
	maxBytes int
	bytes    int
	order    *list.List               // ön: en yeni
	items    map[string]*list.Element // anahtar -> order öğesi
}

type thumbItem struct {
	key  string
	data []byte
}

func newThumbCache(maxBytes int) *thumbCache {
	return &thumbCache{maxBytes: maxBytes, order: list.New(), items: make(map[string]*list.Element)}
}

func (c *thumbCache) get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		return el.Value.(*thumbItem).data, true
	}
	return nil, false
}

func (c *thumbCache) put(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.items[key]; ok {
		return
	}
	c.items[key] = c.order.PushFront(&thumbItem{key, data})
	c.bytes += len(data)
	for c.bytes > c.maxBytes && c.order.Len() > 1 {
		el := c.order.Back()
		it := el.Value.(*thumbItem)
		c.order.Remove(el)
		delete(c.items, it.key)
		c.bytes -= len(it.data)
	}
}
