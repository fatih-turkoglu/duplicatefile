import { test } from 'node:test'
import assert from 'node:assert/strict'
import { fileName, folderOf, folderCounts, isUnder, selectInFolder } from './results/selection.ts'
import { planRenames, validateName } from './results/rename.ts'

test('macOS paths preserve Unicode, case, roots and literal backslashes', () => {
  assert.equal(fileName('/Users/ayşe/Fotoğraflar/a.jpg'), 'a.jpg')
  assert.equal(fileName('/Volumes/Arşiv/a\\b.txt'), 'a\\b.txt')
  assert.equal(folderOf('/Users/ayşe/a.jpg'), '/Users/ayşe')
  assert.equal(folderOf('/a.jpg'), '/')
  assert.equal(folderOf('/'), '/')
  assert.ok(isUnder('/Volumes/Arşiv/alt/a.jpg', '/Volumes/Arşiv'))
  assert.ok(isUnder('/a.jpg', '/'))
  assert.ok(!isUnder('/Volumes/Arşiv2/a.jpg', '/Volumes/Arşiv'))
  assert.ok(!isUnder('/Volumes/arşiv/a.jpg', '/Volumes/Arşiv'))
})

test('Mac folder selection keeps one copy and lists all parents', () => {
  const files = ['/Users/test/Foto/a.jpg', '/Users/test/Foto/alt/b.jpg'].map((path) =>
    ({ id: 0, path, size: 10, modTime: 0, fullHash: 'h' }))
  const groups = [{ hash: 'h', size: 10, files }]
  const counts = Object.fromEntries(folderCounts(groups).map(({ folder, count }) => [folder, count]))
  assert.deepEqual(counts, { '/': 2, '/Users': 2, '/Users/test': 2, '/Users/test/Foto': 2, '/Users/test/Foto/alt': 1 })
  const { selection } = selectInFolder(groups, new Set(), '/Users/test/Foto')
  assert.equal(selection.size, 1)
})

test('Mac rename templates preserve POSIX names and distinct folders', () => {
  const plan = planRenames(['/Volumes/disk/Foto/a.jpg', '/Volumes/disk/foto/b.jpg'], 'yeni{ext}')
  assert.ok(plan.every((item) => item.error === null))
  assert.equal(planRenames(['/Users/test/a.jpg'], '{name}?{ext}')[0].newName, 'a?.jpg')
  assert.equal(validateName('CON.txt', false), null)
  assert.equal(validateName('a\\b?.txt', false), null)
  for (const name of ['', '..', 'a/b', 'a\x00b', 'a:b']) assert.ok(validateName(name, false))
  const collision = planRenames(['/Users/test/a.jpg', '/Users/test/b.jpg'], 'same{ext}')
  assert.ok(collision.every((item) => item.error?.code === 'duplicate'))
})
