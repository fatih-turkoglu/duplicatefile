// scancli, tarama motorunu arayüz olmadan denemek ve ölçmek için geliştirici
// aracıdır.
//
//	go run ./cmd/scancli -db test.db D:\Klasor1 D:\Klasor2
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"DuplicateFile/internal/db"
	"DuplicateFile/internal/scanner"
)

func main() {
	dbPath := flag.String("db", "scancli.db", "veritabanı dosyası")
	minSize := flag.Int64("min", 0, "en küçük dosya boyutu (bayt)")
	workers := flag.Int("workers", 0, "paralel hash sayısı (0 = otomatik)")
	exts := flag.String("ext", "", "virgülle ayrılmış uzantılar, örn. jpg,png")
	hidden := flag.Bool("skip-hidden", false, "gizli/sistem dosyalarını atla")
	top := flag.Int("top", 10, "gösterilecek en büyük grup sayısı")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "kullanım: scancli [seçenekler] klasör...")
		flag.PrintDefaults()
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	d, err := db.Open(ctx, *dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer d.Close()

	opts := scanner.Options{Roots: flag.Args(), MinSize: *minSize, Workers: *workers, SkipHidden: *hidden}
	if *exts != "" {
		opts.Extensions = strings.Split(*exts, ",")
	}

	res, err := scanner.New(d).Scan(ctx, opts, func(p scanner.Progress) {
		line := fmt.Sprintf("[%-7s] %d", p.Stage, p.Done)
		if p.Total > 0 {
			line += fmt.Sprintf("/%d", p.Total)
		}
		if p.TotalBytes > 0 {
			line += fmt.Sprintf("  %s/%s", human(p.Bytes), human(p.TotalBytes))
		}
		fmt.Fprintf(os.Stderr, "\r%-70s", line)
	})
	fmt.Fprintln(os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var wasted int64
	for _, g := range res.Groups {
		wasted += g.Size * int64(len(g.Files)-1)
	}
	fmt.Printf("Durum        : %s\n", res.Status)
	fmt.Printf("Süre         : %d ms\n", res.DurationMs)
	fmt.Printf("Dosya        : %d (önbellekten %d, okunan %d, %s)\n", res.FilesSeen, res.CacheHits, res.Hashed, human(res.BytesRead))
	fmt.Printf("Temizlenen   : %d kayıt\n", res.Removed)
	fmt.Printf("Kopya grubu  : %d (kazanılabilir alan %s)\n", len(res.Groups), human(wasted))
	fmt.Printf("Hata         : %d\n", res.ErrorCount)
	for i, g := range res.Groups {
		if i == *top {
			break
		}
		fmt.Printf("\n%s x%d\n", human(g.Size), len(g.Files))
		for _, f := range g.Files {
			fmt.Printf("  %s\n", f.Path)
		}
	}
}

func human(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
