// Arayüz duman testi: `wails dev` sunucusuna (gerçek Go bağlamalarıyla) kurulu
// Edge üzerinden bağlanır, bir tarama yapar ve her adımın ekran görüntüsünü alır.
//
// Kullanım (proje kökünde, iki ayrı terminalde):
//   . .\env.ps1; wails dev
//   . .\env.ps1; node frontend\e2e\smoke.mjs <taranacak-klasör> [çıktı-klasörü]
//
// Klasör seçme diyaloğu otomatikleştirilemediği için seçenekler localStorage'a
// önceden yazılır; test sonunda temizlenir.
import { chromium } from 'playwright-core'
import path from 'node:path'
import fs from 'node:fs'

const [scanRoot, outDir = 'e2e-out'] = process.argv.slice(2)
if (!scanRoot) {
  console.error('kullanım: node frontend/e2e/smoke.mjs <taranacak-klasör> [çıktı-klasörü]')
  process.exit(2)
}
const edge = [
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  'C:/Program Files/Microsoft/Edge/Application/msedge.exe',
].find((p) => fs.existsSync(p))
fs.mkdirSync(outDir, { recursive: true })

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const browser = await chromium.launch({ executablePath: edge, headless: true })
const page = await browser.newPage({ colorScheme: 'dark', viewport: { width: 1100, height: 760 } })

// Yalnızca tarayıcıdan `wails dev` kullanırken çıkan, uygulamaya ait olmayan hatalar.
const ignored = [/favicon\.ico/, /wails\/ipc\.js/]
const errors = []
page.on('console', (m) => {
  const text = `${m.text()} @ ${m.location().url}`
  if (m.type() === 'error' && !ignored.some((r) => r.test(text))) errors.push(text)
})
page.on('pageerror', (e) => {
  if (!ignored.some((r) => r.test(e.stack))) errors.push(e.stack)
})

const shot = async (name) => {
  const f = path.join(outDir, `${name}.png`)
  await page.screenshot({ path: f, fullPage: true })
  console.log('ekran:', f)
}

try {
  await page.goto('http://127.0.0.1:34115')
  await page.evaluate((root) => {
    localStorage.setItem('language', 'tr') // testler Türkçe metinlerle yazıldı
    localStorage.setItem('scanOptions', JSON.stringify({ roots: [root], minSize: 0, skipHidden: true }))
    localStorage.setItem('mui-mode', 'system')
  }, scanRoot)
  await page.reload()
  await page.waitForSelector('text=Taramayı başlat')
  await shot('1-form')

  await page.click('text=Taramayı başlat')
  for (let i = 0; i < 4; i++) {
    await sleep(300)
    if (await page.isVisible('text=Durdur')) await shot(`2-progress-${i}`)
  }
  await page.waitForSelector('text=Tarama tamamlandı', { timeout: 300000 })
  await shot('3-summary')

  await page.setViewportSize({ width: 400, height: 800 })
  await sleep(300)
  await shot('4-summary-narrow')

  await page.setViewportSize({ width: 1100, height: 760 })
  await page.click('text=Kopyaları göster')
  await page.waitForSelector('text=kazanılabilir')
  await sleep(300)
  await shot('5-results')

  // Sanal liste: DOM'daki dosya satırı sayısı toplamdan çok az olmalı.
  const rendered = await page.locator('[aria-label$=" seç"]').count()
  const total = await page.evaluate(() => {
    const m = document.body.innerText.match(/([\d.]+) dosya ·/)
    return m ? Number(m[1].replace(/\./g, '')) : 0
  })
  console.log(`sanal liste: ${rendered} satır çizildi / ${total} dosya`)
  if (total > 200 && rendered > 150) throw new Error('liste sanallaştırılmamış görünüyor')

  // Aşağı kaydırınca yeni satırlar çizilmeli, araç çubuğu üstte kalmalı.
  await page.evaluate(() => document.querySelector('main').scrollTo({ top: 20000 }))
  await sleep(300)
  await shot('6-results-scrolled')

  await page.evaluate(() => document.querySelector('main').scrollTo({ top: 0 }))
  await page.fill('input[placeholder="Dosya ya da klasör ara"]', 'zerrors_windows')
  await sleep(400)
  await shot('7-results-search')

  await page.click('button:has-text("Seç")')
  await page.click('text=Her grupta en yeniyi tut')
  await page.waitForSelector('text=dosya seçili')
  await sleep(300)
  await shot('8-results-keep-newest')
  if (await page.isVisible('text=Tümü seçili')) throw new Error('en yeniyi tut bir grubu tamamen seçti')

  await page.fill('input[placeholder="Dosya ya da klasör ara"]', '')
  await page.click('button:has-text("Seç")')
  await page.click('text=Klasördeki kopyaları seç…')
  await page.fill('input[placeholder="Aramak için yazın"]', 'gopath\\pkg\\mod\\golang.org')
  await sleep(300)
  await page.keyboard.press('ArrowDown')
  await page.keyboard.press('Enter')
  await page.click('[role="dialog"] button:has-text("Seç")')
  await page.waitForSelector('text=dosya seçildi')
  await sleep(300)
  await shot('9-results-folder-select')

  await page.setViewportSize({ width: 400, height: 800 })
  await sleep(400)
  await shot('10-results-narrow')

  await page.setViewportSize({ width: 1100, height: 760 })
  await page.click('[aria-label="Temayı değiştir"]') // sistem -> açık
  await page.click('nav >> text=Ayarlar')
  await sleep(500)
  await shot('11-settings-light')
} finally {
  await page.evaluate(() => {
    localStorage.removeItem('scanOptions')
    localStorage.removeItem('mui-mode')
  }).catch(() => {})
  await browser.close()
}

if (errors.length) {
  console.error('konsol hataları:', errors)
  process.exit(1)
}
console.log('konsol hatası yok')
