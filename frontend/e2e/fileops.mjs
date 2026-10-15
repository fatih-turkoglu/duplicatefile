// Dosya işlemleri uçtan uca testi: geçici bir klasörde kopyalar oluşturur,
// `wails dev` üzerinden tarar; silme (son kopya korumasıyla), taşıma ve toplu
// yeniden adlandırmayı arayüzden yapar ve sonucu diskte doğrular.
// Silmede kalıcı silme kullanılır; Geri Dönüşüm Kutusu'na dosya bırakılmaz.
//
// Kullanım (proje kökünde, iki ayrı terminalde):
//   . .\env.ps1; wails dev
//   . .\env.ps1; node frontend\e2e\fileops.mjs [çıktı-klasörü]
import { chromium } from 'playwright-core'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'

const outDir = process.argv[2] ?? 'e2e-out'
fs.mkdirSync(outDir, { recursive: true })

// ---- Test klasörü
const base = fs.mkdtempSync(path.join(os.tmpdir(), 'duplicatefile-e2e-'))
const src = path.join(base, 'kaynak')
const dest = path.join(base, 'hedef')
fs.mkdirSync(dest)
const write = (rel, content, ageDays) => {
  const p = path.join(src, rel)
  fs.mkdirSync(path.dirname(p), { recursive: true })
  fs.writeFileSync(p, content)
  const t = new Date(Date.now() - ageDays * 86400000)
  fs.utimesSync(p, t, t)
  return p
}
const big = (ch) => Buffer.alloc(20000, ch)
const files = {
  foto: [write('a\\foto.jpg', big('x'), 1), write('b\\foto.jpg', big('x'), 5), write('c\\foto.jpg', big('x'), 9)],
  belge: [write('a\\belge.pdf', big('y'), 2), write('b\\belge.pdf', big('y'), 3)],
  not: [write('a\\not.txt', 'aynı not', 2), write('b\\not.txt', 'aynı not', 3)],
}
const exists = (p) => fs.existsSync(p)

// ---- Tarayıcı
const edge = [
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  'C:/Program Files/Microsoft/Edge/Application/msedge.exe',
].find((p) => fs.existsSync(p))
const browser = await chromium.launch({ executablePath: edge, headless: true })
const page = await browser.newPage({ viewport: { width: 1100, height: 800 } })
const errors = []
// Test dosyalarının içeriği gerçek görsel değil ("foto.jpg" yalnızca 'x' bayt):
// küçük resim isteği 422 alır ve arayüz kırık görsel simgesi gösterir; beklenen davranış.
const ignored = [/favicon\.ico/, /wails\/ipc\.js/, /status of 422 .*\/thumb\?/]
page.on('pageerror', (e) => !ignored.some((r) => r.test(e.stack)) && errors.push(e.stack))
page.on('console', (m) => {
  const t = `${m.text()} @ ${m.location().url}`
  if (m.type() === 'error' && !ignored.some((r) => r.test(t))) errors.push(t)
})

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const shot = async (name) => {
  const f = path.join(outDir, `ops-${name}.png`)
  await page.screenshot({ path: f })
  console.log('ekran:', f)
}
const search = async (q) => {
  await page.fill('input[placeholder="Dosya ya da klasör ara"]', q)
  await sleep(300)
}
// CSS seçicisinde '\' kaçış karakteri olduğu için yollar etiketle aranır.
const checkbox = (p) => page.getByLabel(`${p} seç`, { exact: true })
const check = (p) => checkbox(p).click()
const selectMenu = async (item) => {
  await page.click('button:has-text("Seç")')
  await page.click(`text=${item}`)
}
const snackbar = async (text) => {
  await page.waitForSelector(`.MuiSnackbar-root >> text=${text}`)
  await page.click('.MuiSnackbar-root', { trial: true }).catch(() => {})
}

try {
  await page.goto('http://127.0.0.1:34115')
  await page.evaluate((root) => {
    localStorage.setItem('language', 'tr') // testler Türkçe metinlerle yazıldı
    localStorage.setItem('scanOptions', JSON.stringify({ roots: [root], minSize: 0, skipHidden: true }))
  }, src)
  await page.reload()
  await page.click('text=Taramayı başlat')
  await page.waitForSelector('text=Tarama tamamlandı', { timeout: 60000 })
  await page.click('text=Kopyaları göster')
  await page.waitForSelector('text=3 grup')

  // Klasör diyaloğu otomatikleştirilemez: SelectFolder hedef klasörü döndürsün.
  await page.evaluate((d) => {
    window.go.main.App.SelectFolder = async () => d
  }, dest)

  // 1) Son kopya koruması: bir grubun bütün kopyaları elle seçilip silinmek istenir.
  await search('foto')
  for (const p of files.foto) await check(p)
  await page.waitForSelector('text=Tümü seçili')
  await page.waitForSelector('text=bütün dosyalar seçili')
  await page.click('button:has-text("Sil")')
  await page.waitForSelector('text=Güvenlik nedeniyle')
  await page.check('text=kalıcı olarak sil')
  await shot('1-delete-dialog-guard')
  await page.click('[role="dialog"] button:has-text("Kalıcı olarak sil")')
  await page.waitForSelector('text=0 başarılı, 3 başarısız')
  await shot('2-report-last-copy')
  for (const p of files.foto) assert.ok(exists(p), `son kopya koruması ${p} dosyasını sildi`)
  await page.click('[role="dialog"] button:has-text("Tamam")')
  console.log('✔ son kopya koruması: hiçbir kopya silinmedi')

  // 2) En yeniyi tut + kalıcı sil: en yeni (a\foto.jpg) kalmalı.
  await selectMenu('Seçimi temizle')
  await selectMenu('Her grupta en yeniyi tut')
  await page.click('button:has-text("Sil")')
  await page.check('text=kalıcı olarak sil')
  await page.click('[role="dialog"] button:has-text("Kalıcı olarak sil")')
  await snackbar('2 dosya kalıcı olarak silindi')
  assert.ok(exists(files.foto[0]), 'en yeni kopya silindi')
  assert.ok(!exists(files.foto[1]) && !exists(files.foto[2]), 'eski kopyalar silinmedi')
  await search('')
  await page.waitForSelector('text=2 grup')
  console.log('✔ en yeniyi tut + sil: yalnızca en yeni kopya kaldı, grup listeden düştü')

  // 3) Taşıma: b\belge.pdf hedef klasöre.
  await search('belge')
  await check(files.belge[1])
  await page.click('button:has-text("Taşı")')
  await page.click('[role="dialog"] button:has-text("Klasör seç")')
  await page.waitForSelector(`[role="dialog"] >> text=${dest}`)
  await shot('3-move-dialog')
  await page.click('[role="dialog"] button:has-text("Taşı")')
  await snackbar('1 dosya taşındı')
  const moved = path.join(dest, 'belge.pdf')
  assert.ok(exists(moved) && !exists(files.belge[1]), 'dosya taşınmadı')
  await checkbox(moved).waitFor()
  console.log('✔ taşıma: dosya diskte taşındı, listede yeni yoluyla görünüyor')

  // 4) Toplu yeniden adlandırma: not grubunun iki kopyası.
  await search('not')
  for (const p of files.not) await check(p)
  await page.click('button:has-text("Yeniden adlandır")')
  await page.fill('[role="dialog"] input', '{ad}_eski_{n}{uzantı}')
  await sleep(200)
  await shot('4-rename-dialog')
  await page.click('[role="dialog"] button:has-text("Yeniden adlandır")')
  await snackbar('2 dosya yeniden adlandırıldı')
  const renamed = [path.join(src, 'a', 'not_eski_1.txt'), path.join(src, 'b', 'not_eski_2.txt')]
  for (const p of renamed) assert.ok(exists(p), `${p} yok`)
  for (const p of files.not) assert.ok(!exists(p), `${p} hâlâ var`)
  await search('not_eski')
  await checkbox(renamed[0]).waitFor()
  await shot('5-after-rename')
  console.log('✔ toplu yeniden adlandırma: diskte ve listede yeni adlar')

  // 5) İşlemlerden sonra yeniden tarama: liste diskteki son durumla aynı olmalı.
  // foto'nun tek kopyası kaldı, belge'nin biri tarama klasörü dışına taşındı;
  // yalnızca yeniden adlandırılan not_eski_1/2 grubu kalır.
  await page.click('nav >> text=Tara')
  await page.click('text=Taramayı başlat')
  await page.waitForSelector('text=Tarama tamamlandı')
  await page.click('text=Kopyaları göster')
  await page.waitForSelector('text=1 grup')
  await checkbox(path.join(src, 'a', 'not_eski_1.txt')).waitFor()
  console.log('✔ işlemlerden sonra yeniden tarama: yalnızca not_eski grubu kaldı')

  // 6) Dar ekranda işlem düğmeleri
  await search('')
  await page.setViewportSize({ width: 400, height: 800 })
  await sleep(300)
  await shot('6-narrow')
} finally {
  await page.evaluate(() => localStorage.removeItem('scanOptions')).catch(() => {})
  await browser.close()
  fs.rmSync(base, { recursive: true, force: true })
}

if (errors.length) {
  console.error('konsol hataları:', errors)
  process.exit(1)
}
console.log('konsol hatası yok')
