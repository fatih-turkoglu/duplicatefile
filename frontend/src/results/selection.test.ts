// Çalıştırma: node --test frontend/src/results/selection.test.ts
import { test } from 'node:test'
import assert from 'node:assert/strict'
import {
  applyChanges,
  filterGroups,
  folderCounts,
  folderOf,
  fullySelectedGroups,
  isUnder,
  keepLargest,
  keepNewest,
  wasted,
  keepOldest,
  selectInFolder,
  selectionStats,
  sortGroups,
} from './selection.ts'

const file = (path: string, modTime = 0, size = 100) => ({ id: 0, path, size, modTime, fullHash: 'h' })
const group = (hash: string, size: number, ...files: ReturnType<typeof file>[]) =>
  ({ hash, size, files: files.map((f) => ({ ...f, size })) }) as any

const A = group('a', 1000, file('D:\\foto\\a.jpg', 3), file('D:\\yedek\\a.jpg', 1), file('E:\\eski\\a.jpg', 2))
const B = group('b', 5000, file('D:\\foto\\b.mp4', 5), file('D:\\yedek\\b.mp4', 5))
const C = group('c', 10, file('D:\\yedek\\x\\c.txt', 1), file('D:\\yedek\\y\\c.txt', 2))

const sorted = (s: Set<string>) => [...s].sort()

test('folderOf ve isUnder', () => {
  assert.equal(folderOf('D:\\a\\b.txt'), 'D:\\a')
  assert.equal(folderOf('D:\\b.txt'), 'D:\\')
  assert.ok(isUnder('D:\\Foto\\a.jpg', 'd:\\foto'))
  assert.ok(isUnder('D:\\foto\\alt\\a.jpg', 'D:\\foto\\'))
  assert.ok(!isUnder('D:\\foto2\\a.jpg', 'D:\\foto'))
  assert.ok(isUnder('D:\\a.jpg', 'D:\\'))
})

test('sortGroups', () => {
  // kazanılabilir: A=2000, B=5000, C=10
  assert.deepEqual(sortGroups([A, B, C], 'wasted').map((g) => g.hash), ['b', 'a', 'c'])
  assert.deepEqual(sortGroups([A, B, C], 'count').map((g) => g.hash), ['a', 'b', 'c'])
  assert.deepEqual(sortGroups([C, A, B], 'size').map((g) => g.hash), ['b', 'a', 'c'])
})

test('filterGroups herhangi bir yolda arar', () => {
  assert.deepEqual(filterGroups([A, B, C], 'ESKI').map((g) => g.hash), ['a'])
  assert.deepEqual(filterGroups([A, B, C], '  ').length, 3)
})

test('keepNewest / keepOldest her grupta tam olarak birini bırakır', () => {
  const newest = keepNewest([A, B, C], new Set())
  assert.deepEqual(sorted(newest), ['D:\\yedek\\a.jpg', 'D:\\yedek\\b.mp4', 'D:\\yedek\\x\\c.txt', 'E:\\eski\\a.jpg'])
  // B'de tarihler eşit: kısa yol (D:\foto\b.mp4) tutulur.
  assert.ok(!newest.has('D:\\foto\\b.mp4'))

  const oldest = keepOldest([A, B, C], new Set())
  assert.ok(!oldest.has('D:\\yedek\\a.jpg'))
  assert.ok(!oldest.has('D:\\yedek\\x\\c.txt'))
  assert.equal(fullySelectedGroups([A, B, C], oldest).length, 0)
})

test('selectInFolder hiçbir grubu tamamen seçmez', () => {
  const { selection, skippedGroups } = selectInFolder([A, B, C], new Set(), 'D:\\yedek')
  // C'nin iki kopyası da D:\yedek altında: biri bırakılır.
  assert.equal(skippedGroups, 1)
  assert.ok(selection.has('D:\\yedek\\a.jpg'))
  assert.ok(selection.has('D:\\yedek\\b.mp4'))
  assert.equal(selectedOf(C, selection), 1)
  assert.equal(fullySelectedGroups([A, B, C], selection).length, 0)
})

test('selectInFolder önceki seçimi hesaba katar', () => {
  // B'nin D:\foto kopyası zaten seçiliyse D:\yedek kopyası da seçilirse grup biter: bırakılmalı.
  const { selection, skippedGroups } = selectInFolder([B], new Set(['D:\\foto\\b.mp4']), 'D:\\yedek')
  assert.equal(skippedGroups, 1)
  assert.equal(fullySelectedGroups([B], selection).length, 0)
})

test('selectionStats ve fullySelectedGroups', () => {
  const sel = new Set(['D:\\foto\\b.mp4', 'D:\\yedek\\b.mp4', 'E:\\eski\\a.jpg'])
  assert.deepEqual(selectionStats([A, B, C], sel), { files: 3, bytes: 11000 })
  assert.deepEqual(fullySelectedGroups([A, B, C], sel).map((g) => g.hash), ['b'])
})

test('folderCounts üst klasörleri de sayar', () => {
  const counts = Object.fromEntries(folderCounts([C]).map((c) => [c.folder, c.count]))
  assert.deepEqual(counts, { 'D:\\': 2, 'D:\\yedek': 2, 'D:\\yedek\\x': 1, 'D:\\yedek\\y': 1 })
})

test('wasted benzer görsellerde en büyüğü tutar', () => {
  assert.equal(wasted(A), 2000) // birebir: 1000 × 2
  const S = {
    hash: 's', size: 900, similar: true,
    files: [
      { ...file('D:\\b.jpg'), size: 900, width: 4000, height: 3000 },
      { ...file('D:\\k.jpg'), size: 200, width: 1000, height: 750 },
      { ...file('D:\\o.jpg'), size: 300, width: 2000, height: 1500 },
    ],
  } as any
  assert.equal(wasted(S), 500)

  // keepLargest en yüksek çözünürlüğü tutar.
  const sel = keepLargest([S], new Set())
  assert.deepEqual(sorted(sel), ['D:\\k.jpg', 'D:\\o.jpg'])
})

test('applyChanges silinenleri çıkarır, taşınanları günceller, tekli grupları düşürür', () => {
  const out = applyChanges(
    [A, B, C],
    new Set(['D:\\yedek\\b.mp4', 'E:\\eski\\a.jpg']),
    new Map([['D:\\yedek\\a.jpg', 'F:\\arsiv\\a.jpg']]),
  )
  assert.deepEqual(out.map((g) => g.hash), ['a', 'c'])
  assert.deepEqual(out[0].files.map((f: any) => f.path), ['D:\\foto\\a.jpg', 'F:\\arsiv\\a.jpg'])
  // Girdi değiştirilmemeli.
  assert.equal(A.files.length, 3)
})

function selectedOf(g: any, sel: Set<string>) {
  return g.files.filter((f: any) => sel.has(f.path)).length
}
