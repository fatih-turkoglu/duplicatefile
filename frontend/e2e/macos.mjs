// Mac UI smoke test: system fonts, theme, POSIX paths, Finder and Trash wording.
// Uses disposable scan fixtures and never confirms deletion.
import { launchBrowser } from './browser.mjs'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'

if (process.platform !== 'darwin') throw new Error('Run this test on macOS.')
const base = fs.mkdtempSync(path.join(os.tmpdir(), 'duplicatefile-mac-'))
const out = process.argv[2] ?? 'e2e-out'
fs.mkdirSync(out, { recursive: true })
const copies = ['Fotoğraflar', 'Yedek'].map((folder) => {
  const dir = path.join(base, folder)
  fs.mkdirSync(dir)
  const file = path.join(dir, 'örnek.txt')
  fs.writeFileSync(file, 'DuplicateFile macOS test')
  return file
})
const browser = await launchBrowser()
const page = await browser.newPage({ viewport: { width: 1100, height: 800 }, colorScheme: 'dark' })
try {
  await page.goto('http://127.0.0.1:34115')
  await page.evaluate((root) => {
    localStorage.clear()
    localStorage.setItem('language', 'tr')
    localStorage.setItem('scanOptions', JSON.stringify({ roots: [root], minSize: 0, skipHidden: true }))
    sessionStorage.clear()
  }, base)
  await page.reload()
  await page.getByRole('button', { name: 'Taramayı başlat', exact: true }).click()
  await page.getByText('Tarama tamamlandı', { exact: true }).waitFor()
  await page.getByRole('button', { name: 'Kopyaları göster', exact: true }).click()
  await page.getByRole('button', { name: `${copies[0]} dosyasını Finder’da göster`, exact: true }).waitFor()
  await page.getByPlaceholder('Dosya ya da klasör ara').fill('örnek')
  await page.getByPlaceholder('Dosya ya da klasör ara').press('Meta+A')
  assert.equal(await page.evaluate(() => document.activeElement.selectionEnd - document.activeElement.selectionStart), 5)
  assert.match(await page.evaluate(() => getComputedStyle(document.body).fontFamily), /-apple-system/)
  assert.ok(await page.evaluate(() => document.documentElement.classList.contains('dark')))
  await page.screenshot({ path: path.join(out, 'macos-results-dark.png') })

  await page.getByLabel(`${copies[0]} seç`, { exact: true }).check()
  await page.getByRole('button', { name: 'Sil', exact: true }).click()
  await page.getByRole('dialog').getByRole('button', { name: 'Çöp Sepeti’ne taşı', exact: true }).waitFor()
  await page.screenshot({ path: path.join(out, 'macos-trash-dialog.png') })
  await page.getByRole('dialog').getByRole('button', { name: 'Vazgeç', exact: true }).click()
  for (const file of copies) assert.ok(fs.existsSync(file))

  await page.getByRole('button', { name: 'Temayı değiştir' }).click()
  await page.locator('html.light').waitFor()
  await page.setViewportSize({ width: 400, height: 800 })
  await page.getByRole('button', { name: 'Tara', exact: true }).waitFor()
  await page.screenshot({ path: path.join(out, 'macos-results-light-narrow.png') })
  console.log('✔ macOS: dark/light, system font, Cmd+A, Unicode paths, Finder and Trash labels')
} finally {
  await browser.close()
  fs.rmSync(base, { recursive: true, force: true })
}
