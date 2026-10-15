import { createContext, Fragment, useCallback, useContext, useEffect, useState } from 'react'
import { api } from '../api'
import { language, setLanguage, type LanguageData } from './core'
import { AUTO, selectLanguage } from './select'

const PREF_KEY = 'language' // localStorage: 'auto' ya da dil kodu

export interface LanguageOption {
  code: string
  name: string
}

interface I18nContextValue {
  code: string // kullanılan dil ('en' = gömülü İngilizce)
  preference: string // 'auto' ya da dil kodu
  autoCode: string // 'auto' seçilirse kullanılacak dil
  options: LanguageOption[] // seçilebilir diller (İngilizce dahil)
  dir: string // lang klasörünün yolu
  setPreference: (pref: string) => void
}

const I18nContext = createContext<I18nContextValue | null>(null)

function readPref(): string {
  try {
    return localStorage.getItem(PREF_KEY) || AUTO
  } catch {
    return AUTO
  }
}

// Açılışta lang klasörünü ve Windows dil sırasını Go'dan alır, dili seçip
// yükler. Dil yüklenene kadar hiçbir şey çizilmez (birkaç milisaniye; İngilizce
// metinlerin bir anlık görünmesini önler). Dil değişince bütün uygulama
// yeniden kurulur, böylece t() çağıran her bileşen yeni dili kullanır.
export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(false)
  const [preference, setPref] = useState(readPref)
  const [state, setState] = useState({ code: 'en', autoCode: 'en', options: [] as LanguageOption[], dir: '' })

  const apply = useCallback(async (pref: string) => {
    let setup = { dir: '', available: [] as LanguageOption[], system: [] as string[] }
    try {
      const s = await api.languageSetup()
      setup = { dir: s.dir, available: s.available ?? [], system: s.system ?? [] }
    } catch {
      // Go'ya ulaşılamazsa İngilizce
    }
    const codes = setup.available.map((a) => a.code)
    const autoCode = selectLanguage(AUTO, setup.system, codes)
    const code = selectLanguage(pref, setup.system, codes)

    let loaded: LanguageData | null = null
    if (codes.includes(code)) {
      try {
        const f = await api.loadLanguage(code)
        loaded = { code: f.code, name: f.name, locale: f.locale, entries: f.entries ?? {} }
        if (f.warnings?.length) console.warn(`${code}.lng:`, f.warnings)
      } catch (e) {
        console.error(`${code}.lng okunamadı:`, e)
      }
    }
    setLanguage(loaded)
    document.documentElement.lang = language().locale

    const options = [{ code: 'en', name: 'English' }, ...setup.available.filter((a) => a.code !== 'en')]
    setState({ code: language().code, autoCode, options, dir: setup.dir })
    setReady(true)
  }, [])

  useEffect(() => {
    apply(preference)
  }, [apply, preference])

  const setPreference = useCallback((pref: string) => {
    try {
      localStorage.setItem(PREF_KEY, pref)
    } catch {
      // yalnızca bu oturum için geçerli olur
    }
    setPref(pref)
  }, [])

  if (!ready) return null
  return (
    <I18nContext.Provider value={{ ...state, preference, setPreference }}>
      <Fragment key={state.code}>{children}</Fragment>
    </I18nContext.Provider>
  )
}

export function useI18n(): I18nContextValue {
  const ctx = useContext(I18nContext)
  if (!ctx) throw new Error('useI18n, I18nProvider içinde kullanılmalı')
  return ctx
}
