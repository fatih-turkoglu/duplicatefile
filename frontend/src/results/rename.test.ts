import { test } from 'node:test'
import assert from 'node:assert/strict'
import { applyTemplate, planRenames, splitName, validateName } from './rename.ts'

test('splitName', () => {
  assert.deepEqual(splitName('a.tar.gz'), { stem: 'a.tar', ext: '.gz' })
  assert.deepEqual(splitName('.gitignore'), { stem: '.gitignore', ext: '' })
  assert.deepEqual(splitName('README'), { stem: 'README', ext: '' })
})

test('applyTemplate', () => {
  assert.equal(applyTemplate('{name} (copy){ext}', 'D:\\f\\tatil.jpg', 1), 'tatil (copy).jpg')
  assert.equal(applyTemplate('{name}_{n}{ext}', 'D:\\f\\a.jpg', 7), 'a_7.jpg')
  assert.equal(applyTemplate('ESKİ-{name}{ext}', 'D:\\f\\a.jpg', 1), 'ESKİ-a.jpg')
  // eski Türkçe yer tutucular
  assert.equal(applyTemplate('{ad} (kopya){uzantı}', 'D:\\f\\tatil.jpg', 1), 'tatil (kopya).jpg')
})

test('validateName Go kurallarıyla aynı', () => {
  for (const ok of ['a.txt', 'çok güzel (2).jpg', '.gitignore', 'console.txt', 'COM10.txt']) {
    assert.equal(validateName(ok), null, ok)
  }
  const cases: [string, string][] = [
    ['', 'empty'], ['.', 'empty'], ['..', 'empty'],
    ['son nokta.', 'trailing'], ['son boşluk ', 'trailing'],
    ['a/b', 'chars'], ['a\\b', 'chars'], ['a:b', 'chars'], ['a?', 'chars'], ['a*', 'chars'],
    ['a"', 'chars'], ['a<', 'chars'], ['a>', 'chars'], ['a|', 'chars'], ['a\tb', 'chars'],
    ['CON', 'reserved'], ['con.txt', 'reserved'], ['Lpt1.log', 'reserved'], ['nul', 'reserved'],
  ]
  for (const [name, code] of cases) {
    assert.equal(validateName(name)?.code, code, JSON.stringify(name))
  }
  assert.deepEqual(validateName('con.txt'), { code: 'reserved', name: 'CON' })
})

test('planRenames çakışmaları ve değişmeyen adları yakalar', () => {
  const plan = planRenames(['D:\\a\\x.jpg', 'D:\\b\\x.jpg', 'D:\\a\\X.png', 'D:\\a\\y.jpg'], 'foto{ext}')
  // D:\a\x.jpg ve D:\a\y.jpg ikisi de D:\a\foto.jpg olur: çakışma.
  assert.equal(plan[0].error?.code, 'duplicate')
  assert.equal(plan[1].error, null) // farklı klasör
  assert.equal(plan[2].error, null) // foto.png
  assert.equal(plan[3].error?.code, 'duplicate')

  const withN = planRenames(['D:\\a\\x.jpg', 'D:\\a\\y.jpg'], 'foto{n}{ext}')
  assert.deepEqual(withN.map((p) => [p.newName, p.error]), [['foto1.jpg', null], ['foto2.jpg', null]])

  assert.equal(planRenames(['D:\\a\\x.jpg'], '{name}{ext}')[0].error?.code, 'unchanged')
  assert.equal(planRenames(['D:\\a\\x.jpg'], '{name}?{ext}')[0].error?.code, 'chars')
})
