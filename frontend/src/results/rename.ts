// Toplu yeniden adlandırma: şablon uygulama ve doğrulama. Kurallar Go'daki
// fileops.ValidateName ile aynıdır; son kontrolü yine de Go tarafı yapar.
// Hatalar kod olarak döner (rename.error.<kod> anahtarıyla çevrilir).
import { fileName, folderOf } from './selection.ts'

// Yer tutucular dilden bağımsızdır; açıklamaları çevrilir.
// Eski Türkçe yer tutucular ({ad}, {uzantı}) da kabul edilir.
export const TEMPLATE_TOKENS = [
  { token: '{name}', label: 'rename.token.name', help: 'rename.token.nameHelp' },
  { token: '{ext}', label: 'rename.token.ext', help: 'rename.token.extHelp' },
  { token: '{n}', label: 'rename.token.n', help: 'rename.token.nHelp' },
]

export type RenameErrorCode = 'empty' | 'trailing' | 'chars' | 'reserved' | 'unchanged' | 'duplicate'

export interface RenameError {
  code: RenameErrorCode
  name?: string // reserved: ayrılmış ad
}

export function splitName(name: string): { stem: string; ext: string } {
  const i = name.lastIndexOf('.')
  // ".gitignore" gibi nokta ile başlayan adların uzantısı yoktur.
  if (i <= 0) return { stem: name, ext: '' }
  return { stem: name.slice(0, i), ext: name.slice(i) }
}

export function applyTemplate(template: string, path: string, n: number): string {
  const { stem, ext } = splitName(fileName(path))
  return template
    .replaceAll('{name}', stem)
    .replaceAll('{ad}', stem)
    .replaceAll('{ext}', ext)
    .replaceAll('{uzantı}', ext)
    .replaceAll('{n}', String(n))
}

const RESERVED = new Set([
  'CON', 'PRN', 'AUX', 'NUL',
  ...[1, 2, 3, 4, 5, 6, 7, 8, 9].flatMap((i) => [`COM${i}`, `LPT${i}`]),
])

// Geçerliyse null, değilse hata kodu döner.
export function validateName(name: string): RenameError | null {
  if (name === '' || name === '.' || name === '..') return { code: 'empty' }
  if (/[ .]$/.test(name)) return { code: 'trailing' }
  // eslint-disable-next-line no-control-regex
  if (/[<>:"/\\|?*\x00-\x1f]/.test(name)) return { code: 'chars' }
  const stem = name.split('.')[0].toUpperCase()
  if (RESERVED.has(stem)) return { code: 'reserved', name: stem }
  return null
}

export interface RenamePlanItem {
  path: string
  newName: string
  error: RenameError | null
}

// Seçilen dosyalara şablonu uygular ve her birini doğrular. Aynı klasörde
// aynı ada düşen dosyalar hata olarak işaretlenir.
export function planRenames(paths: string[], template: string): RenamePlanItem[] {
  const items: RenamePlanItem[] = paths.map((path, i) => {
    const newName = applyTemplate(template, path, i + 1)
    let error = validateName(newName)
    if (!error && newName === fileName(path)) error = { code: 'unchanged' }
    return { path, newName, error }
  })

  const key = (it: RenamePlanItem) => `${folderOf(it.path)}\\${it.newName}`.toLowerCase()
  const seen = new Map<string, number>()
  for (const it of items) seen.set(key(it), (seen.get(key(it)) ?? 0) + 1)
  for (const it of items) {
    if (!it.error && seen.get(key(it))! > 1) it.error = { code: 'duplicate' }
  }
  return items
}
