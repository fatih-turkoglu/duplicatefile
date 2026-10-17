// Dil desteği uçtan uca testi.
//
//   1) Dil dosyası yokken İngilizce:
//        $env:DUPLICATEFILE_LANG_DIR = "<boş klasör>"; wails dev
//        node frontend\e2e\language.mjs no-lang [çıktı-klasörü]
//   2) Dil dosyalarıyla (otomatik seçim, dil değiştirme, çoğul, sonuçların korunması):
//        $env:DUPLICATEFILE_LANG_DIR = "<proje>\lang"; wails dev
//        node frontend\e2e\language.mjs with-lang [çıktı-klasörü]
import { launchBrowser } from './browser.mjs'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'

const [mode, outDir = 'e2e-out'] = process.argv.slice(2)
if (mode !== 'no-lang' && mode !== 'with-lang') {
  console.error('kullanım: node frontend/e2e/language.mjs no-lang|with-lang [çıktı-klasörü]')
  process.exit(2)
}
fs.mkdirSync(outDir, { recursive: true })

const browser = await launchBrowser()
const page = await browser.newPage({ viewport: { width: 1100, height: 900 } })
const errors = []
const ignored = [/favicon\.ico/, /wails\/ipc\.js/]
page.on('pageerror', (e) => !ignored.some((r) => r.test(e.stack)) && errors.push(e.stack))
page.on('console', (m) => {
  const t = `${m.text()} @ ${m.location().url}`
  if (m.type() === 'error' && !ignored.some((r) => r.test(t))) errors.push(t)
})
const shot = async (name) => {
  const f = path.join(outDir, `lang-${mode}-${name}.png`)
  await page.screenshot({ path: f, fullPage: true })
  console.log('ekran:', f)
}
const chooseLanguage = async (label) => {
  await page.getByRole('combobox', { name: /^(Language|Dil|Sprache|Langue|Idioma|Limbă|Език)$/ }).click()
  await page.getByRole('option', { name: label }).click()
}

const base = fs.mkdtempSync(path.join(os.tmpdir(), 'duplicatefile-lang-'))
try {
  await page.goto('http://127.0.0.1:34115')
  await page.evaluate(() => {
    localStorage.removeItem('language') // otomatik
    sessionStorage.clear()
  })
  await page.reload()
  const setup = await page.evaluate(() => window.go.main.App.LanguageSetup())
  console.log('lang klasörü:', setup.dir, '| diller:', (setup.available ?? []).map((a) => a.code).join(','), '| sistem:', setup.system.join(','))

  if (mode === 'no-lang') {
    assert.equal((setup.available ?? []).length, 0, 'bu modda lang klasörü boş olmalı')
    await page.waitForSelector('text=Start scan')
    await page.waitForSelector('nav >> text=Settings')
    await shot('1-scan-english')
    await page.click('nav >> text=Settings')
    await page.getByRole('combobox', { name: 'Language' }).click()
    const options = await page.getByRole('option').allInnerTexts()
    assert.deepEqual(options, ['Automatic (English)', 'English'], `seçenekler: ${options}`)
    await page.keyboard.press('Escape')
    await shot('2-settings-english')
    console.log('✔ dil dosyası yokken arayüz İngilizce, seçenek yalnızca English')
  } else {
    // Otomatik: Windows dil sırasında dosyası olan ilk dil (yoksa İngilizce).
    const settingsLabel = { tr: 'Ayarlar', de: 'Einstellungen', fr: 'Paramètres', pt: 'Configurações', ro: 'Setări', bg: 'Настройки', en: 'Settings' }
    const avail = (setup.available ?? []).map((a) => a.code)
    let autoCode = 'en'
    for (const s of setup.system) {
      const primary = s.toLowerCase().split('-')[0]
      if (avail.includes(primary) || primary === 'en') {
        autoCode = primary
        break
      }
    }
    await page.waitForSelector(`nav >> text=${settingsLabel[autoCode]}`)
    console.log(`✔ otomatik seçim: sistem ${setup.system[0]} → ${autoCode}`)

    // Bir tarama yap: dil değişince sonuçlar kaybolmamalı.
    fs.mkdirSync(path.join(base, 'a'))
    fs.mkdirSync(path.join(base, 'b'))
    fs.writeFileSync(path.join(base, 'a', 'x.txt'), 'aynı')
    fs.writeFileSync(path.join(base, 'b', 'x.txt'), 'aynı')
    await page.evaluate((root) => {
      localStorage.setItem('language', 'tr')
      localStorage.setItem('scanOptions', JSON.stringify({ roots: [root], minSize: 0, skipHidden: true }))
    }, base)
    await page.reload()
    await page.click('text=Taramayı başlat')
    await page.waitForSelector('text=Tarama tamamlandı')

    // Ayarlar → Deutsch: arayüz Almanca, sayfa Ayarlar'da kalır.
    await page.click('nav >> text=Ayarlar')
    await chooseLanguage('Deutsch')
    await page.waitForSelector('nav >> text=Einstellungen')
    await page.waitForSelector('h5:has-text("Einstellungen")') // hâlâ ayarlar sayfasında
    await page.waitForSelector('text=Sprache')
    assert.equal(await page.evaluate(() => document.documentElement.lang), 'de-DE')
    await shot('1-settings-de')
    console.log('✔ Deutsch seçildi, sayfa korundu, <html lang> = de-DE')

    // Sonuçlar korunmuş olmalı (Almanca çoğul: "1 Gruppe").
    await page.click('nav >> text=Duplikate')
    await page.waitForSelector('text=1 Gruppe')
    await page.waitForSelector('text=2 Kopien')
    await shot('2-results-de')
    console.log('✔ dil değişince tarama sonuçları korundu ("1 Gruppe", "2 Kopien")')

    // Română: üç çoğul biçimi; 2 → "few"
    await page.click('nav >> text=Einstellungen')
    await chooseLanguage('Română')
    await page.waitForSelector('nav >> text=Setări')
    await page.click('nav >> text=Duplicate')
    await page.waitForSelector('text=2 fișiere') // few (de olmadan)
    await page.waitForSelector('text=1 grup')
    console.log('✔ Română: "1 grup", "2 fișiere" (few)')

    // Français: birimler "o/Ko"
    await page.click('nav >> text=Setări')
    await chooseLanguage('Français')
    await page.waitForSelector('nav >> text=Paramètres')
    await page.click('nav >> text=Doublons')
    await page.waitForSelector('text=/\\d+ o récupérables/')
    await shot('3-results-fr')
    console.log('✔ Français: bayt birimi "o"')

    // Български: Kiril metin
    await page.click('nav >> text=Paramètres')
    await chooseLanguage('Български')
    await page.waitForSelector('nav >> text=Настройки')
    await shot('4-settings-bg')
    console.log('✔ Български')

    // Português
    await chooseLanguage('Português')
    await page.waitForSelector('nav >> text=Configurações')
    console.log('✔ Português')

    // Otomatiğe dön
    await page.getByRole('combobox', { name: 'Idioma' }).click()
    await page.getByRole('option', { name: /^Automático/ }).click()
    await page.waitForSelector(`nav >> text=${settingsLabel[autoCode]}`)
    console.log(`✔ otomatiğe dönüldü (${autoCode})`)
  }
} finally {
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  }).catch(() => {})
  await browser.close()
  fs.rmSync(base, { recursive: true, force: true })
}

if (errors.length) {
  console.error('konsol hataları:', errors)
  process.exit(1)
}
console.log('konsol hatası yok')
