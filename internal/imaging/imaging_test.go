package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/draw"
)

// scene, tohuma göre rastgele dolu daireler ve bir gradyan içeren görsel çizer.
func scene(seed int64, w, h int) *image.RGBA {
	r := rand.New(rand.NewSource(seed))
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 255 / w), uint8(y * 255 / h), 128, 255})
		}
	}
	for i := 0; i < 25; i++ {
		cx, cy, rad := r.Intn(w), r.Intn(h), 20+r.Intn(w/5)
		c := color.RGBA{uint8(r.Intn(256)), uint8(r.Intn(256)), uint8(r.Intn(256)), 255}
		for y := max(0, cy-rad); y < min(h, cy+rad); y++ {
			for x := max(0, cx-rad); x < min(w, cx+rad); x++ {
				if (x-cx)*(x-cx)+(y-cy)*(y-cy) < rad*rad {
					img.Set(x, y, c)
				}
			}
		}
	}
	return img
}

func resize(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
	return dst
}

func brighten(src *image.RGBA, d int) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	for i, v := range src.Pix {
		if i%4 == 3 {
			dst.Pix[i] = v
			continue
		}
		dst.Pix[i] = uint8(min(255, int(v)+d))
	}
	return dst
}

func save(t *testing.T, dir, name string, img image.Image) string {
	t.Helper()
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if filepath.Ext(name) == ".png" {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: 55})
	}
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func analyze(t *testing.T, p string) Info {
	t.Helper()
	info, err := Analyze(p)
	if err != nil {
		t.Fatalf("Analyze(%s): %v", p, err)
	}
	return info
}

func TestDHashSimilarAndDifferent(t *testing.T) {
	dir := t.TempDir()
	orig := scene(1, 1200, 900)
	base := analyze(t, save(t, dir, "orig.png", orig))
	if base.Width != 1200 || base.Height != 900 {
		t.Errorf("çözünürlük = %dx%d", base.Width, base.Height)
	}

	similar := map[string]string{
		"küçültülmüş JPEG": save(t, dir, "kucuk.jpg", resize(orig, 400, 300)),
		"aynı boyut JPEG":  save(t, dir, "ayni.jpg", orig),
		"parlak":           save(t, dir, "parlak.png", brighten(orig, 25)),
	}
	for name, p := range similar {
		d := Distance(base.Hash, analyze(t, p).Hash)
		t.Logf("%s: mesafe %d", name, d)
		if d > 6 {
			t.Errorf("%s: mesafe = %d, benzer sayılmalıydı (≤ 6)", name, d)
		}
	}

	for seed := int64(2); seed <= 4; seed++ {
		other := analyze(t, save(t, dir, "baska.png", scene(seed, 1200, 900)))
		d := Distance(base.Hash, other.Hash)
		t.Logf("farklı görsel (tohum %d): mesafe %d", seed, d)
		if d < 16 {
			t.Errorf("tohum %d: farklı görselin mesafesi = %d, çok yakın", seed, d)
		}
	}
}

func TestThumbnail(t *testing.T) {
	dir := t.TempDir()
	p := save(t, dir, "buyuk.png", scene(1, 1000, 500))
	data, err := Thumbnail(p, 160)
	if err != nil {
		t.Fatal(err)
	}
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("küçük resim JPEG değil: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 160 || b.Dy() != 80 {
		t.Errorf("küçük resim = %dx%d, beklenen 160x80", b.Dx(), b.Dy())
	}

	// Küçük görsel büyütülmez.
	small := save(t, dir, "kucuk.png", scene(1, 50, 30))
	data, _ = Thumbnail(small, 160)
	img, _ = jpeg.Decode(bytes.NewReader(data))
	if b := img.Bounds(); b.Dx() != 50 || b.Dy() != 30 {
		t.Errorf("küçük görsel büyütülmüş: %dx%d", b.Dx(), b.Dy())
	}
}

func TestErrorsAndFormats(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bozuk.jpg")
	os.WriteFile(bad, []byte("bu bir görsel değil"), 0o644)
	if _, err := Analyze(bad); err == nil {
		t.Error("bozuk dosya hata vermeliydi")
	}

	for p, want := range map[string]bool{
		"a.JPG": true, "a.jpeg": true, "a.png": true, "a.webp": true, "a.tiff": true, "a.bmp": true,
		"a.heic": false, "a.txt": false, "jpg": false,
	} {
		if Supported(p) != want {
			t.Errorf("Supported(%q) = %v", p, !want)
		}
	}

	h := uint64(0xa1b2c3d4e5f60789)
	if got, ok := ParseHash(FormatHash(h)); !ok || got != h {
		t.Errorf("hash gidiş-dönüş: %x", got)
	}
	if Distance(0, ^uint64(0)) != 64 || Distance(h, h) != 0 {
		t.Error("Distance yanlış")
	}
}
