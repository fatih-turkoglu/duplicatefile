// Package imaging, görsellerin algısal hash'ini (dHash) hesaplar ve önizleme
// için küçük resim üretir. Yalnızca saf Go çözücüleri kullanılır; HEIC gibi
// CGO gerektiren biçimler desteklenmez. EXIF yönü dikkate alınmaz.
package imaging

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif" // biçim kaydı
	"image/jpeg"
	_ "image/png"
	"math/bits"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Desteklenen uzantılar (küçük harf, noktalı).
var supported = map[string]bool{
	".jpg": true, ".jpeg": true, ".jfif": true, ".png": true, ".gif": true,
	".bmp": true, ".tif": true, ".tiff": true, ".webp": true,
}

// maxPixels, çözülecek en büyük görsel; daha büyükleri (ya da boyutunu yanlış
// bildiren bozuk dosyalar) belleği tüketmesin diye reddedilir.
const maxPixels = 150_000_000

var (
	ErrTooLarge   = errors.New("image is too large")
	ErrUnreadable = errors.New("could not read image")
)

// Supported, uzantının desteklenen bir görsel biçimi olup olmadığını söyler.
func Supported(path string) bool {
	return supported[strings.ToLower(filepath.Ext(path))]
}

// Extensions desteklenen uzantıları döner.
func Extensions() []string {
	out := make([]string, 0, len(supported))
	for e := range supported {
		out = append(out, e)
	}
	return out
}

// Info bir görselin algısal hash'i ve çözünürlüğüdür.
type Info struct {
	Hash          uint64
	Width, Height int
}

// Analyze görseli çözer ve dHash'ini hesaplar.
func Analyze(path string) (Info, error) {
	img, err := decode(path)
	if err != nil {
		return Info{}, err
	}
	b := img.Bounds()
	return Info{Hash: DHash(img), Width: b.Dx(), Height: b.Dy()}, nil
}

func decode(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnreadable, err)
	}
	if int64(cfg.Width)*int64(cfg.Height) > maxPixels {
		return nil, fmt.Errorf("%w (%d×%d)", ErrTooLarge, cfg.Width, cfg.Height)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnreadable, err)
	}
	return img, nil
}

// DHash, görselin 64 bitlik fark hash'idir: görsel 9×8 gri tona küçültülür
// ve her satırda yan yana iki pikselden soldakinin daha koyu olup olmadığı
// bir bit olarak yazılır. Boyutlandırma ve yeniden sıkıştırmaya dayanıklıdır.
func DHash(img image.Image) uint64 {
	const w, h = 9, 8
	g := shrinkGray(img, w, h)
	var hash uint64
	for y := 0; y < h; y++ {
		for x := 0; x < w-1; x++ {
			hash <<= 1
			if g[y*w+x] < g[y*w+x+1] {
				hash |= 1
			}
		}
	}
	return hash
}

// Distance iki hash arasında farklı olan bit sayısıdır (0 = aynı, 64 = tersi).
func Distance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

// FormatHash / ParseHash, hash'i veritabanında 16 karakterlik hex olarak tutar.
func FormatHash(h uint64) string {
	return fmt.Sprintf("%016x", h)
}

func ParseHash(s string) (uint64, bool) {
	h, err := strconv.ParseUint(s, 16, 64)
	return h, err == nil
}

// shrinkGray, görseli w×h hücreye bölüp her hücrenin ortalama parlaklığını
// döner. Hız için her hücreden en fazla 8×8 nokta örneklenir; dHash için bu
// yeterince kararlıdır.
func shrinkGray(img image.Image, w, h int) []float64 {
	const samples = 8
	b := img.Bounds()
	out := make([]float64, w*h)
	lum := luminance(img)
	for cy := 0; cy < h; cy++ {
		y0 := b.Min.Y + cy*b.Dy()/h
		y1 := b.Min.Y + (cy+1)*b.Dy()/h
		for cx := 0; cx < w; cx++ {
			x0 := b.Min.X + cx*b.Dx()/w
			x1 := b.Min.X + (cx+1)*b.Dx()/w
			var sum float64
			var n int
			for sy := 0; sy < samples; sy++ {
				y := y0 + (2*sy+1)*(y1-y0)/(2*samples)
				for sx := 0; sx < samples; sx++ {
					x := x0 + (2*sx+1)*(x1-x0)/(2*samples)
					sum += lum(x, y)
					n++
				}
			}
			out[cy*w+cx] = sum / float64(n)
		}
	}
	return out
}

// luminance, sık görülen görsel tiplerinde At() arayüz çağrısını atlayan
// hızlı bir parlaklık okuyucusu döner.
func luminance(img image.Image) func(x, y int) float64 {
	switch m := img.(type) {
	case *image.YCbCr: // JPEG
		return func(x, y int) float64 { return float64(m.Y[m.YOffset(x, y)]) }
	case *image.Gray:
		return func(x, y int) float64 { return float64(m.Pix[m.PixOffset(x, y)]) }
	case *image.NRGBA:
		return func(x, y int) float64 {
			i := m.PixOffset(x, y)
			return lumRGB(m.Pix[i], m.Pix[i+1], m.Pix[i+2])
		}
	case *image.RGBA:
		return func(x, y int) float64 {
			i := m.PixOffset(x, y)
			return lumRGB(m.Pix[i], m.Pix[i+1], m.Pix[i+2])
		}
	}
	return func(x, y int) float64 {
		return float64(color.GrayModel.Convert(img.At(x, y)).(color.Gray).Y)
	}
}

func lumRGB(r, g, b uint8) float64 {
	return 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
}

// Thumbnail, en uzun kenarı size pikseli geçmeyecek şekilde küçültülmüş JPEG
// döner. Zaten küçük görseller büyütülmez.
func Thumbnail(path string, size int) ([]byte, error) {
	img, err := decode(path)
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > size || h > size {
		if w >= h {
			w, h = size, max(1, h*size/w)
		} else {
			w, h = max(1, w*size/h), size
		}
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	// Saydam PNG'ler siyah değil beyaz zemin üzerinde görünsün.
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src)
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 82}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
