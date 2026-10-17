package scanner

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/image/draw"
)

// pattern, tohuma göre rastgele dikdörtgenlerden oluşan bir görsel çizer.
func pattern(seed int64, w, h int) *image.RGBA {
	r := rand.New(rand.NewSource(seed))
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{128, 128, 128, 255}), image.Point{}, draw.Src)
	for i := 0; i < 30; i++ {
		x, y := r.Intn(w), r.Intn(h)
		rect := image.Rect(x, y, x+w/6+r.Intn(w/4), y+h/6+r.Intn(h/4))
		c := color.RGBA{uint8(r.Intn(256)), uint8(r.Intn(256)), uint8(r.Intn(256)), 255}
		draw.Draw(img, rect, image.NewUniform(c), image.Point{}, draw.Src)
	}
	return img
}

func saveImage(t *testing.T, root, rel string, img image.Image) string {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(strings.ReplaceAll(rel, `\`, "/")))
	os.MkdirAll(filepath.Dir(p), 0o755)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if filepath.Ext(p) == ".png" {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: 60})
	}
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func scaled(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
	return dst
}

func TestScanSimilarImages(t *testing.T) {
	root := t.TempDir()
	a := pattern(1, 1200, 800)
	big := saveImage(t, root, `foto\tatil.png`, a)
	small := saveImage(t, root, `yedek\tatil_kucuk.jpg`, scaled(a, 600, 400))
	other := saveImage(t, root, `foto\baska.png`, pattern(2, 1200, 800))
	write(t, root, `belge.txt`, []byte("görsel değil"))
	write(t, root, `bozuk.jpg`, []byte("bozuk görsel"))

	s := New(openDB(t))
	res := scan(t, s, Options{Roots: []string{root}, Mode: ModeSimilar})
	if res.Mode != ModeSimilar || res.Threshold != DefaultThreshold {
		t.Errorf("mod/eşik = %s/%d", res.Mode, res.Threshold)
	}
	// belge.txt hiç taranmamalı; bozuk.jpg hata olarak bildirilmeli.
	if res.FilesSeen != 4 || res.ErrorCount != 1 {
		t.Errorf("FilesSeen = %d, ErrorCount = %d; beklenen 4, 1 (%+v)", res.FilesSeen, res.ErrorCount, res.Errors)
	}
	if len(res.Groups) != 1 {
		t.Fatalf("grup sayısı = %d, beklenen 1: %+v", len(res.Groups), res.Groups)
	}
	g := res.Groups[0]
	if !g.Similar || len(g.Files) != 2 {
		t.Fatalf("grup: %+v", g)
	}
	// Referans (ilk) en yüksek çözünürlüklü görsel olmalı.
	if g.Files[0].Path != big || g.Files[0].Width != 1200 || g.Files[0].Similarity != 100 {
		t.Errorf("referans: %+v", g.Files[0])
	}
	if g.Files[1].Path != small || g.Files[1].Width != 600 || g.Files[1].Similarity < 90 {
		t.Errorf("küçük kopya: %+v", g.Files[1])
	}
	for _, f := range g.Files {
		if f.Path == other {
			t.Error("farklı görsel gruba girmiş")
		}
	}

	// İkinci tarama: görseller yeniden çözülmemeli (bozuk olan yine denenir).
	res = scan(t, s, Options{Roots: []string{root}, Mode: ModeSimilar})
	if res.CacheHits != 3 || res.Hashed != 0 {
		t.Errorf("ikinci tarama: CacheHits = %d, Hashed = %d; beklenen 3, 0", res.CacheHits, res.Hashed)
	}

	// Çok sıkı eşikte küçük JPEG artık benzer sayılmayabilir ama 64'te her şey tek grup.
	res = scan(t, s, Options{Roots: []string{root}, Mode: ModeSimilar, Threshold: 64})
	if len(res.Groups) != 1 || len(res.Groups[0].Files) != 3 {
		t.Errorf("eşik 64: %+v", res.Groups)
	}
}

// Birebir kopya taraması, benzer görsel taramasının önbelleğe aldığı algısal
// hash'i silmemeli (ve tersi).
func TestModesKeepEachOthersCache(t *testing.T) {
	root := t.TempDir()
	img := pattern(1, 400, 300)
	p1 := saveImage(t, root, `a\x.png`, img)
	saveImage(t, root, `b\x.png`, img) // birebir kopya: exact taramada hash'lenir

	d := openDB(t)
	s := New(d)
	ctx := context.Background()

	scan(t, s, Options{Roots: []string{root}, Mode: ModeSimilar})
	scan(t, s, Options{Roots: []string{root}}) // exact: dosyayı okuyup yazar
	f, err := d.FileByPath(ctx, p1)
	if err != nil || f.PHash == "" || f.FullHash == "" || f.Width != 400 {
		t.Fatalf("exact taramadan sonra kayıt: %+v %v", f, err)
	}

	res := scan(t, s, Options{Roots: []string{root}, Mode: ModeSimilar})
	if res.Hashed != 0 {
		t.Errorf("algısal hash önbelleği kaybolmuş: Hashed = %d", res.Hashed)
	}
	f, _ = d.FileByPath(ctx, p1)
	if f.FullHash == "" {
		t.Error("benzer görsel taraması tam hash'i silmiş")
	}
}

func TestUnionFindTransitive(t *testing.T) {
	// a~b (1 bit), b~c (1 bit), a~c (2 bit): eşik 1'de üçü tek grup olmalı.
	es := []*entry{
		{path: "a", phash: "0000000000000000", width: 10, height: 10, size: 1},
		{path: "b", phash: "0000000000000001", width: 10, height: 10, size: 1},
		{path: "c", phash: "0000000000000003", width: 10, height: 10, size: 1},
		{path: "d", phash: "ffffffffffffffff", width: 10, height: 10, size: 1},
	}
	groups := buildSimilarGroups(es, 1, 3)
	if len(groups) != 1 || len(groups[0].Files) != 3 {
		t.Fatalf("gruplar: %+v", groups)
	}
}
