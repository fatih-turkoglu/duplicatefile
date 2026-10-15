// genicon, uygulama simgesini (build/appicon.png, 1024×1024) çizer: tema
// rengindeki yuvarlatılmış kare üzerinde üst üste iki belge ve öndekinde "=".
// Wails, build/windows/icon.ico yoksa bu PNG'den üretir.
//
//	go run ./cmd/genicon
//	Remove-Item build\windows\icon.ico; wails build
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/vector"
)

const size = 1024

var (
	primary = color.NRGBA{0x39, 0x49, 0xab, 0xff} // tema: primary.main (açık)
	back    = color.NRGBA{0xff, 0xff, 0xff, 0x8c} // arkadaki belge, yarı saydam
	front   = color.NRGBA{0xff, 0xff, 0xff, 0xff}
)

func main() {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	// Zemin 40..984; belgeler bunun içinde, 120 px kaydırılmış.
	fill(img, roundRect(40, 40, 984, 984, 220), primary)
	fill(img, document(210, 150, 500, 640, 48, 140), back)  // 210..710 × 150..790
	fill(img, document(330, 270, 500, 640, 48, 140), front) // 330..830 × 270..910
	// "=" işareti, öndeki belgenin ortasında (merkez ≈ 580, 620)
	fill(img, roundRect(440, 560, 720, 620, 30), primary)
	fill(img, roundRect(440, 670, 720, 730, 30), primary)

	f, err := os.Create("build/appicon.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Println("build/appicon.png yazıldı")
}

type path func(r *vector.Rasterizer)

func fill(dst *image.NRGBA, p path, c color.NRGBA) {
	r := vector.NewRasterizer(size, size)
	p(r)
	r.ClosePath()
	r.Draw(dst, dst.Bounds(), image.NewUniform(c), image.Point{})
}

// roundRect, köşeleri rad yarıçaplı dikdörtgen.
func roundRect(x0, y0, x1, y1, rad float32) path {
	const k = 0.5523 // çeyrek çemberin kübik Bézier yaklaşımı
	return func(r *vector.Rasterizer) {
		r.MoveTo(x0+rad, y0)
		r.LineTo(x1-rad, y0)
		r.CubeTo(x1-rad+rad*k, y0, x1, y0+rad-rad*k, x1, y0+rad)
		r.LineTo(x1, y1-rad)
		r.CubeTo(x1, y1-rad+rad*k, x1-rad+rad*k, y1, x1-rad, y1)
		r.LineTo(x0+rad, y1)
		r.CubeTo(x0+rad-rad*k, y1, x0, y1-rad+rad*k, x0, y1-rad)
		r.LineTo(x0, y0+rad)
		r.CubeTo(x0, y0+rad-rad*k, x0+rad-rad*k, y0, x0+rad, y0)
	}
}

// document, sağ üst köşesi kıvrık (fold) yuvarlatılmış belge.
func document(x, y, w, h, rad, fold float32) path {
	const k = 0.5523
	x1, y1 := x+w, y+h
	return func(r *vector.Rasterizer) {
		r.MoveTo(x+rad, y)
		r.LineTo(x1-fold, y)
		r.LineTo(x1, y+fold)
		r.LineTo(x1, y1-rad)
		r.CubeTo(x1, y1-rad+rad*k, x1-rad+rad*k, y1, x1-rad, y1)
		r.LineTo(x+rad, y1)
		r.CubeTo(x+rad-rad*k, y1, x, y1-rad+rad*k, x, y1-rad)
		r.LineTo(x, y+rad)
		r.CubeTo(x, y+rad-rad*k, x+rad-rad*k, y, x+rad, y)
	}
}
