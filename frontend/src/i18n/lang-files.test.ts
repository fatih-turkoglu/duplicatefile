// lang/*.lng dosyalarının tutarlılık testi: eksik/bilinmeyen anahtar, yer
// tutucu uyumu, dile uygun çoğul ekleri, @name/@locale ve en.lng'nin en.ts ile
// aynı olması. Çalıştırma: npm test
import { test } from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { en } from './en.ts'

const langDir = path.join(path.dirname(fileURLToPath(import.meta.url)), '..', '..', '..', 'lang')
const PLURALS = ['zero', 'one', 'two', 'few', 'many', 'other']

// Go'daki internal/i18n ayrıştırıcısıyla aynı kurallar (yorum, @ayar, kaçışlar).
function parse(text: string) {
  const meta: Record<string, string> = {}
  const entries: Record<string, string> = {}
  const problems: string[] = []
  text.replace(/^﻿/, '').split(/\r?\n/).forEach((raw, i) => {
    const line = raw.trim()
    if (!line || line.startsWith('#') || line.startsWith(';')) return
    const eq = line.indexOf('=')
    if (eq <= 0) return problems.push(`satır ${i + 1}: 'anahtar = metin' değil`)
    const key = line.slice(0, eq).trim()
    const value = line.slice(eq + 1).trim().replace(/\\(n|t|\\)/g, (_, c) => (c === 'n' ? '\n' : c === 't' ? '\t' : '\\'))
    if (key.startsWith('@')) meta[key.slice(1)] = value
    else {
      if (key in entries) problems.push(`satır ${i + 1}: ${key} tekrar ediyor`)
      entries[key] = value
    }
  })
  return { meta, entries, problems }
}

// İngilizcedeki anahtarları "temel anahtar" → { çoğul mu, yer tutucular } olarak çıkarır.
const placeholders = (s: string) => new Set([...s.matchAll(/\{(\w+)\}/g)].map((m) => m[1]))
const bases = new Map<string, { plural: boolean; vars: Set<string> }>()
for (const [key, value] of Object.entries(en)) {
  const last = key.split('.').pop()!
  const plural = PLURALS.includes(last)
  const base = plural ? key.slice(0, -last.length - 1) : key
  const cur = bases.get(base) ?? { plural, vars: new Set<string>() }
  placeholders(value).forEach((v) => cur.vars.add(v))
  bases.set(base, cur)
}

const files = fs.existsSync(langDir) ? fs.readdirSync(langDir).filter((f) => f.endsWith('.lng')) : []

test('lang klasöründe beklenen diller var', () => {
  for (const code of ['en', 'tr', 'de', 'fr', 'pt', 'ro', 'bg']) {
    assert.ok(files.includes(`${code}.lng`), `${code}.lng yok`)
  }
})

for (const file of files) {
  test(`${file}`, () => {
    const { meta, entries, problems } = parse(fs.readFileSync(path.join(langDir, file), 'utf8'))
    assert.deepEqual(problems, [], 'biçim sorunları')
    assert.ok(meta.name, '@name yok')
    assert.ok(meta.locale, '@locale yok')
    assert.doesNotThrow(() => new Intl.PluralRules(meta.locale), `@locale geçersiz: ${meta.locale}`)
    const categories = new Intl.PluralRules(meta.locale).resolvedOptions().pluralCategories as string[]

    const errors: string[] = []
    const seen = new Set<string>()
    for (const [key, value] of Object.entries(entries)) {
      const last = key.split('.').pop()!
      const isForm = PLURALS.includes(last)
      const base = isForm ? key.slice(0, -last.length - 1) : key
      const info = bases.get(base)
      if (!info) {
        errors.push(`bilinmeyen anahtar: ${key}`)
        continue
      }
      seen.add(base)
      if (info.plural && !isForm) errors.push(`${key}: çoğul anahtar, ek gerekli (.other …)`)
      if (!info.plural && isForm) errors.push(`${key}: çoğul değil, ek kullanılmamalı`)
      if (isForm && !categories.includes(last)) errors.push(`${key}: ${meta.locale} dilinde '${last}' çoğul biçimi yok`)
      // Yer tutucular İngilizcede olmalı; çoğul olmayan metinlerde hepsi bulunmalı.
      const vars = placeholders(value)
      for (const v of vars) if (!info.vars.has(v)) errors.push(`${key}: bilinmeyen yer tutucu {${v}}`)
      if (!info.plural) for (const v of info.vars) if (!vars.has(v)) errors.push(`${key}: {${v}} eksik`)
      if (!value) errors.push(`${key}: boş`)
    }
    for (const [base, info] of bases) {
      if (!seen.has(base)) errors.push(`eksik anahtar: ${base}`)
      else if (info.plural && !(`${base}.other` in entries)) errors.push(`${base}.other eksik`)
    }
    assert.deepEqual(errors, [])
  })
}

test('en.lng, en.ts ile aynı (npm run lang:en)', () => {
  const { entries } = parse(fs.readFileSync(path.join(langDir, 'en.lng'), 'utf8'))
  assert.deepEqual(entries, en)
})
