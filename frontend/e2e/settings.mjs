// Ayarlar ve veritabanı bakımı uçtan uca testi. Önbelleği temizlediği ve
// veritabanını taşıdığı için YALNIZCA geçici bir veri klasörüyle çalışır:
//
//   $env:DUPLICATEFILE_DATA_DIR = "<boş geçici klasör>"; . .\env.ps1; wails dev
//   . .\env.ps1; node frontend\e2e\settings.mjs "<aynı klasör>" [çıktı-klasörü]
//
// Uygulamanın gösterdiği veritabanı yolu verilen klasörde değilse test
// hiçbir şey yapmadan durur (gerçek kullanıcı verisini korumak için).
import { chromium } from 'playwright-core'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'

const [dataDir, outDir = 'e2e-out'] = process.argv.slice(2)
if (!dataDir) {
  console.error('kullanım: node frontend/e2e/settings.mjs <DUPLICATEFILE_DATA_DIR> [çıktı-klasörü]')
  process.exit(2)
}
fs.mkdirSync(outDir, { recursive: true })
const base = fs.mkdtempSync(path.join(os.tmpdir(), 'duplicatefile-settings-'))
const moved = path.join(base, 'tasinan-db')

const edge = [
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  'C:/Program Files/Microsoft/Edge/Application/msedge.exe',
].find((p) => fs.existsSync(p))
const browser = await chromium.launch({ executablePath: edge, headless: true })
const page = await browser.newPage({ viewport: { width: 1100, height: 900 } })
const errors = []
const ignored = [/favicon\.ico/, /wails\/ipc\.js/]
page.on('pageerror', (e) => !ignored.some((r) => r.test(e.stack)) && errors.push(e.stack))
page.on('console', (m) => {
  const t = `${m.text()} @ ${m.location().url}`
  if (m.type() === 'error' && !ignored.some((r) => r.test(t))) errors.push(t)
})
const shot = async (name) => {
  const f = path.join(outDir, `settings-${name}.png`)
  await page.screenshot({ path: f, fullPage: true })
  console.log('ekran:', f)
}
const snackbar = (text) => page.waitForSelector(`.MuiSnackbar-root >> text=${text}`)
const dbPathText = () => page.locator('text=index.db').first().innerText()
const under = (p, dir) => path.resolve(p).toLowerCase().startsWith(path.resolve(dir).toLowerCase() + path.sep)

try {
  await page.goto('http://127.0.0.1:34115')
  await page.evaluate(() => localStorage.setItem('language', 'tr')) // testler Türkçe metinlerle yazıldı
  await page.reload()
  await page.click('nav >> text=Ayarlar')
  await page.waitForSelector('text=index.db')
  const startPath = await dbPathText()
  if (!under(startPath, dataDir)) {
    throw new Error(`Uygulamanın veritabanı (${startPath}) verilen geçici klasörde değil; test durduruldu.`)
  }
  if (!(await page.isVisible('text=0 dosya kaydı'))) {
    throw new Error('Veri klasörü boş değil (önceki çalıştırmadan kayıt kalmış); wails dev\'i boş bir klasörle yeniden başlatın.')
  }

  // Taranacak dosyalar: iki kopya + bir tekil.
  const src = path.join(base, 'kaynak')
  fs.mkdirSync(path.join(src, 'a'), { recursive: true })
  fs.mkdirSync(path.join(src, 'b'), { recursive: true })
  fs.writeFileSync(path.join(src, 'a', 'x.txt'), 'aynı içerik')
  fs.writeFileSync(path.join(src, 'b', 'x.txt'), 'aynı içerik')
  fs.writeFileSync(path.join(src, 'tek.txt'), 'tekil')
  await page.evaluate((root) => {
    localStorage.setItem('scanOptions', JSON.stringify({ roots: [root], minSize: 0, skipHidden: true }))
  }, src)
  await page.reload()
  await page.click('nav >> text=Tara') // açık sayfa oturumda saklanır; yeniden yükleme Ayarlar'da kalır
  await page.click('text=Taramayı başlat')
  await page.waitForSelector('text=Tarama tamamlandı')

  await page.click('nav >> text=Ayarlar')
  await page.waitForSelector('text=3 dosya kaydı')
  await shot('1-settings')
  console.log('✔ istatistikler: 3 dosya kaydı')

  // Sıkıştır
  await page.click('button:has-text("Sıkıştır")')
  await snackbar('Sıkıştırıldı')
  console.log('✔ sıkıştırma')

  // Silinmiş dosyaların kayıtları
  fs.rmSync(path.join(src, 'tek.txt'))
  await page.click('button:has-text("Kayıtları temizle")')
  await snackbar('1 kayıt temizlendi')
  await page.waitForSelector('text=2 dosya kaydı')
  console.log('✔ silinmiş dosya kaydı temizlendi')

  // Paralel okuma ayarı tarama seçeneklerine yazılmalı.
  await page.getByRole('combobox', { name: 'Aynı anda okunan dosya' }).click()
  await page.getByRole('option', { name: /^1 \(klasik sabit disk/ }).click()
  const workers = await page.evaluate(() => JSON.parse(localStorage.getItem('scanOptions')).workers)
  assert.equal(workers, 1)
  console.log('✔ paralel okuma ayarı kaydedildi')

  // Başka klasöre taşı (klasör diyaloğu yerine sabit yol)
  await page.evaluate((d) => {
    window.go.main.App.SelectFolder = async () => d
  }, moved)
  await page.click('button:has-text("Başka klasöre taşı")')
  await snackbar('Veritabanı taşındı')
  await page.waitForSelector('text=Özel konum')
  const movedPath = await dbPathText()
  assert.ok(under(movedPath, moved), `yeni yol: ${movedPath}`)
  assert.ok(fs.existsSync(path.join(moved, 'index.db')), 'yeni konumda dosya yok')
  assert.ok(!fs.existsSync(startPath), 'eski veritabanı silinmemiş')
  const cfg = JSON.parse(fs.readFileSync(path.join(dataDir, 'config.json'), 'utf8'))
  assert.ok(under(cfg.dbPath, moved), `config.json: ${JSON.stringify(cfg)}`)
  await page.waitForSelector('text=2 dosya kaydı') // kayıtlar taşındı
  await shot('2-moved')
  console.log('✔ veritabanı kayıtlarıyla taşındı, config.json güncellendi')

  // Varsayılana geri
  await page.click('button:has-text("Varsayılan konuma taşı")')
  await snackbar('Veritabanı taşındı')
  await page.waitForSelector('text=Varsayılan konum')
  assert.ok(fs.existsSync(startPath), 'varsayılan konuma dönülmedi')
  assert.ok(!fs.existsSync(path.join(moved, 'index.db')), 'özel konumdaki dosya kalmış')
  console.log('✔ varsayılan konuma geri taşındı')

  // Önbelleği temizle: sonuçlar da kalkmalı.
  await page.click('button:has-text("Önbelleği temizle")')
  await page.waitForSelector('text=Önbellek temizlensin mi?')
  await shot('3-confirm-clear')
  await page.click('[role="dialog"] button:has-text("Temizle")')
  await snackbar('Önbellek temizlendi')
  await page.waitForSelector('text=0 dosya kaydı')
  await page.click('nav >> text=Kopyalar')
  await page.waitForSelector('text=Henüz sonuç yok')
  console.log('✔ önbellek temizlendi, sonuçlar kaldırıldı')

  await page.click('nav >> text=Ayarlar')
  await page.setViewportSize({ width: 400, height: 900 })
  await page.waitForTimeout(300)
  await shot('4-narrow')
} finally {
  await page.evaluate(() => localStorage.clear()).catch(() => {})
  await browser.close()
  fs.rmSync(base, { recursive: true, force: true })
}

if (errors.length) {
  console.error('konsol hataları:', errors)
  process.exit(1)
}
console.log('konsol hatası yok')
