// Sayı, boyut, süre ve tarihleri seçili dilin biçimiyle gösterir.
import { locale, t } from './i18n/core.ts'

// Intl biçimlendiricileri pahalıdır; dil değişene kadar saklanır.
let cachedLocale = ''
let numberFmt: Intl.NumberFormat
let decimalFmt: Intl.NumberFormat
let dateFmt: Intl.DateTimeFormat
let percentFmt: Intl.NumberFormat

function build(loc: string) {
  numberFmt = new Intl.NumberFormat(loc)
  decimalFmt = new Intl.NumberFormat(loc, { maximumFractionDigits: 1 })
  dateFmt = new Intl.DateTimeFormat(loc, { dateStyle: 'medium', timeStyle: 'short' })
  percentFmt = new Intl.NumberFormat(loc, { style: 'percent', maximumFractionDigits: 0 })
}

function fmt() {
  const loc = locale()
  if (loc !== cachedLocale) {
    cachedLocale = loc
    try {
      build(loc)
    } catch {
      build('en-US') // dil dosyasında geçersiz @locale
    }
  }
  return { numberFmt, decimalFmt, dateFmt, percentFmt }
}

// formatPercent(38) → "38%" / "%38" / "38 %" (dile göre).
export function formatPercent(percent: number): string {
  return fmt().percentFmt.format(percent / 100)
}

export function formatNumber(n: number): string {
  return fmt().numberFmt.format(n)
}

export function formatBytes(bytes: number): string {
  const { numberFmt, decimalFmt } = fmt()
  if (bytes < 1024) return `${numberFmt.format(bytes)} ${t('unit.B')}`
  const units = ['unit.KB', 'unit.MB', 'unit.GB', 'unit.TB']
  let value = bytes / 1024
  let i = 0
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024
    i++
  }
  return `${decimalFmt.format(value)} ${t(units[i])}`
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return t('duration.ms', { n: ms })
  const s = Math.round(ms / 1000)
  if (s < 60) return t('duration.s', { n: s })
  const m = Math.floor(s / 60)
  if (m < 60) return t('duration.minSec', { m, s: s % 60 })
  return t('duration.hourMin', { h: Math.floor(m / 60), m: m % 60 })
}

// Go'dan gelen Unix nanosaniye zamanını tarih olarak gösterir.
export function formatDate(unixNano: number): string {
  return fmt().dateFmt.format(new Date(unixNano / 1e6))
}

// Süren bir işlemin geçen süresi; saniye altını göstermez ("0 s", "1 min 5 s").
export function formatElapsed(ms: number): string {
  return ms < 1000 ? t('duration.s', { n: 0 }) : formatDuration(Math.floor(ms / 1000) * 1000)
}

// Uzun yolu ortadan kısaltır; dosya adı görünür kalsın diye sonu korunur.
export function shortenPath(path: string, max = 80): string {
  if (path.length <= max) return path
  const tail = Math.floor(max * 0.6)
  return path.slice(0, max - tail - 1) + '…' + path.slice(-tail)
}

// Dosya adlarını seçili dilin alfabetik sırasıyla karşılaştırır.
export function compareText(a: string, b: string): number {
  return a.localeCompare(b, locale())
}
