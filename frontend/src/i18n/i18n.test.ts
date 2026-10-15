import { test } from 'node:test'
import assert from 'node:assert/strict'
import { errorText, opErrorText, setLanguage, t } from './core.ts'
import { selectLanguage } from './select.ts'

test('t: İngilizce, yer tutucu ve çoğul', () => {
  setLanguage(null)
  assert.equal(t('nav.scan'), 'Scan')
  assert.equal(t('toolbar.files', { count: 1 }), '1 file')
  assert.equal(t('toolbar.files', { count: 1234 }), '1,234 files')
  assert.equal(t('scan.threshold.label', { value: 8 }), 'Similarity threshold: 8')
  assert.equal(t('olmayan.anahtar'), 'olmayan.anahtar')
  // params'ta olmayan yer tutucu olduğu gibi kalır
  assert.equal(t('rename.defaultTemplate'), '{name} (copy){ext}')
  assert.equal(t('rename.error.duplicate', {}), 'Another file in the same folder gets this name; add {n} to the template.')
})

test('t: seçili dil, eksik anahtarda İngilizce, dile özgü çoğul', () => {
  setLanguage({
    code: 'ro',
    name: 'Română',
    locale: 'ro-RO',
    entries: {
      'nav.scan': 'Scanare',
      'toolbar.files.one': '{count} fișier',
      'toolbar.files.few': '{count} fișiere',
      'toolbar.files.other': '{count} de fișiere',
    },
  })
  assert.equal(t('nav.scan'), 'Scanare')
  assert.equal(t('nav.settings'), 'Settings') // eksik → İngilizce
  assert.equal(t('toolbar.files', { count: 1 }), '1 fișier')
  assert.equal(t('toolbar.files', { count: 5 }), '5 fișiere') // Romence "few"
  assert.equal(t('toolbar.files', { count: 25 }), '25 de fișiere')
  assert.equal(t('toolbar.files', { count: 1500 }), '1.500 de fișiere') // ro-RO binlik ayıracı

  // Türkçede yalnızca .other: sayı ne olursa olsun o kullanılır
  setLanguage({ code: 'tr', name: 'Türkçe', locale: 'tr-TR', entries: { 'toolbar.files.other': '{count} dosya' } })
  assert.equal(t('toolbar.files', { count: 1 }), '1 dosya')
  // dilde hiç çoğul yoksa İngilizcenin kuralı
  assert.equal(t('toolbar.groups', { count: 1 }), '1 group')

  // geçersiz @locale uygulamayı çökertmemeli
  setLanguage({ code: 'xx', name: 'X', locale: 'geçersiz yerel', entries: {} })
  assert.equal(t('toolbar.files', { count: 2 }), '2 files')
  setLanguage(null)
})

test('errorText ve opErrorText', () => {
  setLanguage(null)
  assert.equal(errorText('[busy] another scan is running'), t('error.busy'))
  assert.equal(errorText('[root_not_found] folder not found "D:\\Fotoğraf": x'), 'Folder not found (D:\\Fotoğraf)')
  assert.equal(errorText('[bilinmeyen_kod] english detail'), 'english detail')
  assert.equal(errorText('kodsuz hata'), 'kodsuz hata')
  assert.equal(errorText(new Error('[busy] x')), t('error.busy'))
  assert.equal(opErrorText('last_copy', 'go metni'), t('opError.last_copy'))
  assert.equal(opErrorText('error', 'go metni'), 'go metni')
  assert.equal(opErrorText(undefined, undefined), '')
})

test('selectLanguage', () => {
  const avail = ['tr', 'de', 'pt-br', 'bg']
  // otomatik: Windows sırası
  assert.equal(selectLanguage('auto', ['tr-TR', 'en-US'], avail), 'tr')
  assert.equal(selectLanguage('auto', ['en-US', 'tr-TR'], avail), 'en') // İngilizce önce
  assert.equal(selectLanguage('auto', ['fr-FR', 'de-DE'], avail), 'de') // fr yok, sıradaki
  assert.equal(selectLanguage('auto', ['pt-PT'], avail), 'pt-br') // aynı dilin varyantı
  assert.equal(selectLanguage('auto', ['ja-JP'], avail), 'en') // hiçbiri yok
  assert.equal(selectLanguage('auto', [], []), 'en') // lang klasörü yok
  // elle seçim
  assert.equal(selectLanguage('bg', ['tr-TR'], avail), 'bg')
  assert.equal(selectLanguage('en', ['tr-TR'], avail), 'en')
  assert.equal(selectLanguage('fr', ['tr-TR'], avail), 'tr') // dosyası silinmiş tercih → otomatik
})
