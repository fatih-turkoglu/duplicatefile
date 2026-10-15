// Çeviri çekirdeği: React'e bağlı değildir (node --test ile sınanır).
// Seçili dil modül düzeyinde tutulur; dil değişince I18nProvider uygulamayı
// yeniden çizer, böylece her bileşen t()'yi doğrudan çağırabilir.
import { en } from './en.ts'

export type Params = Record<string, string | number>

export interface LanguageData {
  code: string
  name: string
  locale: string // Intl için: tr-TR, de-DE …
  entries: Record<string, string>
}

const english: LanguageData = { code: 'en', name: 'English', locale: 'en-US', entries: {} }

let current = english
let plural = new Intl.PluralRules(current.locale)
const englishPlural = new Intl.PluralRules('en-US')
let numberFmt = new Intl.NumberFormat(current.locale)

// setLanguage seçili dili değiştirir; null İngilizceye döner.
export function setLanguage(lang: LanguageData | null): void {
  current = lang ?? english
  plural = safePluralRules(current.locale)
  numberFmt = safeNumberFormat(current.locale)
}

export function language(): LanguageData {
  return current
}

// locale, Intl biçimlendirmesi için geçerli dil etiketi.
export function locale(): string {
  return current.locale
}

// has, anahtarın seçili dilde ya da İngilizcede bulunup bulunmadığını söyler.
export function has(key: string): boolean {
  return key in current.entries || key in en
}

// t, anahtarın seçili dildeki metnini döner. Sırası: seçili dil → İngilizce
// → anahtarın kendisi. params.count bir sayıysa dilin çoğul kuralına göre
// .one/.few/.other … eki seçilir. {ad} yer tutucuları params'tan doldurulur;
// sayılar dile göre biçimlenir; params'ta olmayan yer tutucu olduğu gibi kalır.
export function t(key: string, params?: Params): string {
  let text: string | undefined
  if (params && typeof params.count === 'number') {
    const rule = plural.select(params.count)
    text =
      current.entries[`${key}.${rule}`] ??
      current.entries[`${key}.other`] ??
      en[`${key}.${englishPlural.select(params.count)}`] ??
      en[`${key}.other`]
  }
  text ??= current.entries[key] ?? en[key] ?? key
  if (!params) return text
  return text.replace(/\{(\w+)\}/g, (whole, name: string) => {
    if (!(name in params)) return whole
    const v = params[name]
    return typeof v === 'number' ? numberFmt.format(v) : v
  })
}

function safePluralRules(loc: string): Intl.PluralRules {
  try {
    return new Intl.PluralRules(loc)
  } catch {
    return new Intl.PluralRules('en-US') // dosyada geçersiz @locale
  }
}

function safeNumberFormat(loc: string): Intl.NumberFormat {
  try {
    return new Intl.NumberFormat(loc)
  } catch {
    return new Intl.NumberFormat('en-US')
  }
}

// errorText, Go'dan gelen "[kod] açıklama" biçimindeki hatayı seçili dile
// çevirir. Kod tanınmıyorsa açıklama (İngilizce) olduğu gibi döner.
// Açıklamada tırnak içinde bir yol varsa çevirinin sonuna eklenir.
export function errorText(err: unknown): string {
  const raw = String(err instanceof Error ? err.message : err)
  const m = /^\[([a-z_]+)\] ([\s\S]*)$/.exec(raw)
  if (!m) return raw
  const [, code, detail] = m
  const key = `error.${code}`
  if (!has(key)) return detail
  const path = /"([^"]+)"/.exec(detail)?.[1]
  return path ? `${t(key)} (${path})` : t(key)
}

// opErrorText, dosya işlemi sonucundaki kodu çevirir; bilinmiyorsa Go'nun metni.
export function opErrorText(code: string | undefined, message: string | undefined): string {
  if (code && has(`opError.${code}`)) return t(`opError.${code}`)
  return message ?? ''
}
