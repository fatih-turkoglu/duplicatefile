package scanner

import (
	"context"
	"math/bits"
	"sort"
	"sync"

	"DuplicateFile/internal/db"
	"DuplicateFile/internal/imaging"
)

// hashImages, algısal hash'i önbellekte olmayan görselleri çözüp dHash'lerini
// ve çözünürlüklerini hesaplar.
func (r *run) hashImages(entries []*entry) {
	var todo []*entry
	for _, e := range entries {
		if e.phash == "" {
			todo = append(todo, e)
		}
	}
	r.hashStage(StageImage, todo,
		func(e *entry) int64 { return e.size },
		func(_ context.Context, e *entry, add func(int64)) error {
			info, err := imaging.Analyze(e.path)
			if err != nil {
				return err
			}
			e.phash = imaging.FormatHash(info.Hash)
			e.width, e.height = info.Width, info.Height
			add(e.size)
			return nil
		})
}

// buildSimilarGroups, dHash'leri arasında en fazla threshold bit fark olan
// görselleri birleşim-bulma (union-find) ile gruplar. Benzerlik geçişlidir:
// A~B ve B~C ise A, B, C aynı gruptadır.
//
// Karşılaştırma tüm çiftler üzerinden yapılır (n²/2); 64 bitlik XOR + popcount
// olduğundan 100.000 görselde bile işçilere bölünmüş hâlde birkaç saniyedir.
func buildSimilarGroups(entries []*entry, threshold, workers int) []db.DuplicateGroup {
	var imgs []*entry
	var hashes []uint64
	for _, e := range entries {
		if e.failed || e.phash == "" {
			continue
		}
		if h, ok := imaging.ParseHash(e.phash); ok {
			imgs = append(imgs, e)
			hashes = append(hashes, h)
		}
	}
	n := len(imgs)
	uf := newUnionFind(n)

	var mu sync.Mutex
	var wg sync.WaitGroup
	for w := range max(1, workers) {
		wg.Go(func() {
			var pairs [][2]int32
			for i := w; i < n; i += max(1, workers) {
				hi := hashes[i]
				for j := i + 1; j < n; j++ {
					if bits.OnesCount64(hi^hashes[j]) <= threshold {
						pairs = append(pairs, [2]int32{int32(i), int32(j)})
					}
				}
			}
			mu.Lock()
			for _, p := range pairs {
				uf.union(int(p[0]), int(p[1]))
			}
			mu.Unlock()
		})
	}
	wg.Wait()

	members := make(map[int][]int)
	for i := range n {
		root := uf.find(i)
		members[root] = append(members[root], i)
	}

	groups := []db.DuplicateGroup{} // nil değil: JSON'da null yerine [] olsun
	for _, idx := range members {
		if len(idx) < 2 {
			continue
		}
		// Referans: en yüksek çözünürlük, sonra en büyük dosya, sonra yol.
		sort.Slice(idx, func(a, b int) bool {
			ea, eb := imgs[idx[a]], imgs[idx[b]]
			pa, pb := ea.width*ea.height, eb.width*eb.height
			if pa != pb {
				return pa > pb
			}
			if ea.size != eb.size {
				return ea.size > eb.size
			}
			return ea.path < eb.path
		})
		ref := hashes[idx[0]]
		g := db.DuplicateGroup{Hash: imgs[idx[0]].phash, Similar: true}
		for _, i := range idx {
			e := imgs[i]
			g.Files = append(g.Files, db.File{
				Path: e.path, Size: e.size, ModTime: e.mtime, FullHash: e.full,
				PHash: e.phash, Width: e.width, Height: e.height,
				Similarity: 100 - imaging.Distance(ref, hashes[i])*100/64,
			})
			g.Size = max(g.Size, e.size)
		}
		groups = append(groups, g)
	}

	// Kazanılabilir alana göre (en büyüğü tutup gerisini silmek) büyükten küçüğe.
	saved := func(g db.DuplicateGroup) int64 {
		var sum int64
		for _, f := range g.Files {
			sum += f.Size
		}
		return sum - g.Size
	}
	sort.Slice(groups, func(i, j int) bool {
		si, sj := saved(groups[i]), saved(groups[j])
		if si != sj {
			return si > sj
		}
		return groups[i].Hash < groups[j].Hash
	})
	return groups
}

type unionFind struct{ parent []int }

func newUnionFind(n int) *unionFind {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &unionFind{p}
}

func (u *unionFind) find(x int) int {
	for u.parent[x] != x {
		u.parent[x] = u.parent[u.parent[x]] // yol yarılama
		x = u.parent[x]
	}
	return x
}

func (u *unionFind) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra != rb {
		u.parent[ra] = rb
	}
}
