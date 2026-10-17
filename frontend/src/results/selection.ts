// Sonuç ekranının saf mantığı: sıralama, filtreleme ve seçim yardımcıları.
// React'e ve Wails'e bağımlı değildir; node --test ile sınanır.
import type { DuplicateGroup, FileEntry } from '../api'
import { isWindowsPath, lastSeparator, pathKey } from '../paths.ts'

export type Selection = ReadonlySet<string> // seçili dosya yolları
export type SortKey = 'wasted' | 'size' | 'count'

// Grupta en büyük dosya tutulup gerisi silinince açılan alan. Birebir
// kopyalarda size × (n − 1); benzer görsellerde dosya boyutları farklıdır.
export function wasted(g: DuplicateGroup): number {
  let sum = 0
  let max = 0
  for (const f of g.files) {
    sum += f.size
    max = Math.max(max, f.size)
  }
  return sum - max
}

export function fileName(path: string): string {
  return path.slice(lastSeparator(path) + 1)
}

export function folderOf(path: string): string {
  const i = lastSeparator(path)
  if (i < 0) return ''
  // "D:\a.txt" → "D:\" (kök klasör ayraçla biter)
  if (i === 0) return path.slice(0, 1)
  return i === 2 && path[1] === ':' ? path.slice(0, 3) : path.slice(0, i)
}

// path, folder'ın kendisinde ya da altında mı (büyük/küçük harf duyarsız)?
export function isUnder(path: string, folder: string): boolean {
  const p = pathKey(path)
  let f = pathKey(folder)
  const sep = isWindowsPath(folder) ? '\\' : '/'
  if (!f.endsWith(sep)) f += sep
  return p.startsWith(f)
}

export function sortGroups(groups: DuplicateGroup[], key: SortKey): DuplicateGroup[] {
  const by: Record<SortKey, (g: DuplicateGroup) => number> = {
    wasted,
    size: (g) => g.size,
    count: (g) => g.files.length,
  }
  const f = by[key]
  // Eşitlikte boyut, sonra hash: sıra her seferinde aynı kalsın.
  return [...groups].sort((a, b) => f(b) - f(a) || b.size - a.size || a.hash.localeCompare(b.hash))
}

// Herhangi bir dosya yolu sorguyu içeren gruplar.
export function filterGroups(groups: DuplicateGroup[], query: string): DuplicateGroup[] {
  const q = query.trim().toLowerCase()
  if (!q) return groups
  return groups.filter((g) => g.files.some((f) => f.path.toLowerCase().includes(q)))
}

// Grupta kaç dosyanın seçili olduğu.
export function selectedCount(g: DuplicateGroup, sel: Selection): number {
  return g.files.reduce((n, f) => n + (sel.has(f.path) ? 1 : 0), 0)
}

// Seçimin tümü bir gruptaki bütün kopyaları kapsıyorsa o gruplar (silinirse
// dosyanın hiçbir kopyası kalmaz).
export function fullySelectedGroups(groups: DuplicateGroup[], sel: Selection): DuplicateGroup[] {
  return groups.filter((g) => g.files.every((f) => sel.has(f.path)))
}

export function selectionStats(groups: DuplicateGroup[], sel: Selection) {
  let files = 0
  let bytes = 0
  for (const g of groups) {
    for (const f of g.files) {
      if (sel.has(f.path)) {
        files++
        bytes += f.size
      }
    }
  }
  return { files, bytes }
}

// Her grupta keep() ile seçilen dosya dışındakileri seçer. Diğer gruplardaki
// mevcut seçimlere dokunmaz.
function selectAllBut(groups: DuplicateGroup[], sel: Selection, keep: (files: FileEntry[]) => FileEntry): Set<string> {
  const next = new Set(sel)
  for (const g of groups) {
    const kept = keep(g.files)
    for (const f of g.files) {
      if (f === kept) next.delete(f.path)
      else next.add(f.path)
    }
  }
  return next
}

// Eşit tarihte kısa yol (çoğunlukla "asıl" konum) tutulur.
const tieBreak = (a: FileEntry, b: FileEntry) => a.path.length - b.path.length || a.path.localeCompare(b.path)

export function keepNewest(groups: DuplicateGroup[], sel: Selection): Set<string> {
  return selectAllBut(groups, sel, (files) =>
    files.reduce((best, f) => (f.modTime > best.modTime || (f.modTime === best.modTime && tieBreak(f, best) < 0) ? f : best)),
  )
}

export function keepOldest(groups: DuplicateGroup[], sel: Selection): Set<string> {
  return selectAllBut(groups, sel, (files) =>
    files.reduce((best, f) => (f.modTime < best.modTime || (f.modTime === best.modTime && tieBreak(f, best) < 0) ? f : best)),
  )
}

// En yüksek çözünürlüklü görseli (eşitse büyük dosyayı) tutar.
export function keepLargest(groups: DuplicateGroup[], sel: Selection): Set<string> {
  const px = (f: FileEntry) => (f.width ?? 0) * (f.height ?? 0)
  return selectAllBut(groups, sel, (files) =>
    files.reduce((best, f) => {
      const d = px(f) - px(best) || f.size - best.size || -tieBreak(f, best)
      return d > 0 ? f : best
    }),
  )
}

// folder altındaki dosyaları seçer. Bir grubun bütün kopyaları o klasördeyse
// biri (kısa yollu olan) seçilmeden bırakılır. skippedGroups: bu nedenle
// tamamı seçilmeyen grup sayısı.
export function selectInFolder(groups: DuplicateGroup[], sel: Selection, folder: string) {
  const next = new Set(sel)
  let skippedGroups = 0
  for (const g of groups) {
    const inside = g.files.filter((f) => isUnder(f.path, folder))
    if (inside.length === 0) continue
    let keep: FileEntry | undefined
    const outsideKept = g.files.some((f) => !isUnder(f.path, folder) && !next.has(f.path))
    if (!outsideKept) {
      keep = [...inside].sort(tieBreak)[0]
      skippedGroups++
    }
    for (const f of inside) {
      if (f === keep) next.delete(f.path)
      else next.add(f.path)
    }
  }
  return { selection: next, skippedGroups }
}

// Sonuçlarda geçen klasörler ve her birinin altındaki kopya dosya sayısı;
// yalnızca dosyaların doğrudan bulunduğu klasörler değil, üst klasörleri de.
export function folderCounts(groups: DuplicateGroup[]): { folder: string; count: number }[] {
  const counts = new Map<string, { folder: string; count: number }>()
  for (const g of groups) {
    for (const f of g.files) {
      for (let dir = folderOf(f.path); dir; ) {
        const key = pathKey(dir)
        const c = counts.get(key)
        if (c) c.count++
        else counts.set(key, { folder: dir, count: 1 })
        const parent = folderOf(dir)
        if (parent === dir) break
        dir = parent
      }
    }
  }
  return [...counts.values()].sort((a, b) => a.folder.localeCompare(b.folder, 'tr'))
}

// Dosya işlemlerinden sonra grupları günceller: silinen/taşınan dosyalar
// çıkarılır ya da yeni yoluna aktarılır; tek kopyası kalan gruplar düşer.
export function applyChanges(
  groups: DuplicateGroup[],
  removed: ReadonlySet<string>,
  renamed: ReadonlyMap<string, string> = new Map(),
): DuplicateGroup[] {
  const out: DuplicateGroup[] = []
  for (const g of groups) {
    const files = g.files
      .filter((f) => !removed.has(f.path))
      .map((f) => (renamed.has(f.path) ? { ...f, path: renamed.get(f.path)! } : f))
    if (files.length > 1) out.push({ ...g, files } as DuplicateGroup)
  }
  return out
}
