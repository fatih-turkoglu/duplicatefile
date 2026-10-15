// Benzer görseller uçtan uca testi: görselleri tarayıcıdaki canvas ile üretip
// geçici klasöre yazar, benzer görsel modunda tarar; küçük resimleri,
// karşılaştırma penceresini, silme korumasını ve "en yüksek çözünürlüklüyü
// tut" ile silmeyi dener; sonucu diskte doğrular. Kalıcı silme kullanılır.
//
// Kullanım (proje kökünde, iki ayrı terminalde):
//   . .\env.ps1; wails dev
//   . .\env.ps1; node frontend\e2e\similar.mjs [çıktı-klasörü]
import { chromium } from 'playwright-core'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'

const outDir = process.argv[2] ?? 'e2e-out'
fs.mkdirSync(outDir, { recursive: true })
const base = fs.mkdtempSync(path.join(os.tmpdir(), 'duplicatefile-similar-'))

const edge = [
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  'C:/Program Files/Microsoft/Edge/Application/msedge.exe',
].find((p) => fs.existsSync(p))
const browser = await chromium.launch({ executablePath: edge, headless: true })
const page = await browser.newPage({ viewport: { width: 1100, height: 800 } })
const errors = []
const ignored = [/favicon\.ico/, /wails\/ipc\.js/]
page.on('pageerror', (e) => !ignored.some((r) => r.test(e.stack)) && errors.push(e.stack))
page.on('console', (m) => {
  const t = `${m.text()} @ ${m.location().url}`
  if (m.type() === 'error' && !ignored.some((r) => r.test(t))) errors.push(t)
})
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const shot = async (name) => {
  const f = path.join(outDir, `similar-${name}.png`)
  await page.screenshot({ path: f })
  console.log('ekran:', f)
}
const checkbox = (p) => page.getByLabel(`${p} seç`, { exact: true })

// Tohuma göre rastgele dikdörtgenler; aynı tohum farklı boyut/biçimde benzer görsel verir.
async function drawImage(seed, w, h, type, quality) {
  const dataUrl = await page.evaluate(
    ({ seed, w, h, type, quality }) => {
      let s = seed
      const rnd = () => ((s = (s * 1103515245 + 12345) % 2147483648) / 2147483648)
      const c = document.createElement('canvas')
      c.width = w
      c.height = h
      const g = c.getContext('2d')
      g.fillStyle = '#808080'
      g.fillRect(0, 0, w, h)
      for (let i = 0; i < 30; i++) {
        g.fillStyle = `rgb(${rnd() * 255},${rnd() * 255},${rnd() * 255})`
        g.fillRect(rnd() * w, rnd() * h, w / 6 + rnd() * (w / 4), h / 6 + rnd() * (h / 4))
      }
      return c.toDataURL(type, quality)
    },
    { seed, w, h, type, quality },
  )
  return Buffer.from(dataUrl.split(',')[1], 'base64')
}

const write = (rel, buf) => {
  const p = path.join(base, rel)
  fs.mkdirSync(path.dirname(p), { recursive: true })
  fs.writeFileSync(p, buf)
  return p
}

try {
  await page.goto('http://127.0.0.1:34115')
  const f = {
    tatil: write('foto\\tatil.png', await drawImage(1, 1200, 800, 'image/png')),
    tatilKucuk: write('yedek\\tatil_kucuk.jpg', await drawImage(1, 600, 400, 'image/jpeg', 0.6)),
    manzara: write('foto\\manzara.png', await drawImage(7, 1600, 1000, 'image/png')),
    manzaraKopya: write('yedek\\manzara_kopya.jpg', await drawImage(7, 800, 500, 'image/jpeg', 0.7)),
    tek: write('foto\\tek.png', await drawImage(42, 900, 900, 'image/png')),
  }
  write('notlar.txt', Buffer.from('görsel değil'))

  await page.evaluate((root) => {
    localStorage.setItem('language', 'tr') // testler Türkçe metinlerle yazıldı
    localStorage.setItem(
      'scanOptions',
      JSON.stringify({ roots: [root], minSize: 0, skipHidden: true, mode: 'similar', threshold: 8 }),
    )
  }, base)
  await page.reload()
  await page.waitForSelector('text=Benzer görseller')
  await shot('1-form')

  await page.click('text=Taramayı başlat')
  await page.waitForSelector('text=Tarama tamamlandı', { timeout: 60000 })
  await page.waitForSelector('text=Benzer görsel grubu')
  await shot('2-summary')

  await page.click('text=Benzerleri göster')
  await page.waitForSelector('text=2 grup')
  // Küçük resimler Go'nun /thumb adresinden gerçekten yüklenmeli.
  await page.waitForFunction(() => {
    const imgs = [...document.querySelectorAll('main img')]
    return imgs.length >= 4 && imgs.every((i) => i.complete && i.naturalWidth > 0)
  }, null, { timeout: 15000 })
  assert.ok(!(await page.isVisible(`text=tek.png`)), 'benzeri olmayan görsel listede')
  await shot('3-results')
  console.log('✔ 2 benzer grup bulundu, küçük resimler yüklendi, tekil görsel listede yok')

  // Karşılaştırma penceresi
  await page.click('button:has-text("Karşılaştır") >> nth=0')
  await page.waitForSelector('text=En yüksek çözünürlük')
  await page.waitForFunction(() => {
    const imgs = [...document.querySelectorAll('[role="dialog"] img')]
    return imgs.length === 2 && imgs.every((i) => i.complete && i.naturalWidth > 0)
  })
  await sleep(500) // açılış animasyonu bitsin
  await shot('4-compare')
  await page.click('[role="dialog"] button:has-text("Kapat")')
  console.log('✔ karşılaştırma penceresi önizlemelerle açıldı')

  // Koruma: bir grubun iki görseli birden silinemez.
  await checkbox(f.tatil).click()
  await checkbox(f.tatilKucuk).click()
  await page.click('button:has-text("Sil")')
  await page.check('text=kalıcı olarak sil')
  await page.click('[role="dialog"] button:has-text("Kalıcı olarak sil")')
  await page.waitForSelector('text=0 başarılı, 2 başarısız')
  assert.ok(fs.existsSync(f.tatil) && fs.existsSync(f.tatilKucuk), 'koruma görsel sildi')
  await page.click('[role="dialog"] button:has-text("Tamam")')
  console.log('✔ benzer görsel koruması: grubun tamamı silinmedi')

  // En yüksek çözünürlüklüyü tut + kalıcı sil
  await page.click('button:has-text("Seç")')
  await page.click('text=Seçimi temizle')
  await page.click('button:has-text("Seç")')
  await page.click('text=Her grupta en yüksek çözünürlüklüyü tut')
  await page.waitForSelector('text=2 dosya seçili')
  await shot('5-keep-largest')
  await page.click('button:has-text("Sil")')
  await page.check('text=kalıcı olarak sil')
  await page.click('[role="dialog"] button:has-text("Kalıcı olarak sil")')
  await page.waitForSelector('text=2 dosya kalıcı olarak silindi')
  assert.ok(fs.existsSync(f.tatil) && fs.existsSync(f.manzara), 'yüksek çözünürlüklü görsel silindi')
  assert.ok(!fs.existsSync(f.tatilKucuk) && !fs.existsSync(f.manzaraKopya), 'küçük kopyalar silinmedi')
  await page.waitForSelector('text=Kopya bulunamadı')
  console.log('✔ en yüksek çözünürlüklüyü tut: küçük kopyalar silindi, büyükler kaldı')

  // Gerileme: hiç grup kalmamışken yeniden tarama boş ekranda kalıyordu
  // (Go boş grup listesini null gönderiyordu).
  await page.click('nav >> text=Tara')
  await page.click('text=Taramayı başlat')
  await page.waitForSelector('text=Tarama tamamlandı')
  await page.waitForSelector('text=Benzer görsel bulunamadı')
  console.log('✔ silme sonrası yeniden tarama: "Benzer görsel bulunamadı"')
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
