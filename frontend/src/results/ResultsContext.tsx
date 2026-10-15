import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { api, type DuplicateGroup, type ScanMode, type ScanOptions } from '../api'
import { useScan } from '../scan/ScanContext'
import { SCAN_OPTIONS_KEY } from '../scan/ScanOptionsForm'
import { applyChanges, type Selection } from './selection'

// scan: bu oturumdaki son taramanın sonucu
// saved: uygulama açılışında veritabanından yüklenen, yeniden doğrulanmamış sonuç
export type ResultsSource = 'scan' | 'saved'

interface ResultsContextValue {
  groups: DuplicateGroup[] | null // null: henüz sonuç yok
  source: ResultsSource | null
  mode: ScanMode // sonuçlar hangi tarama türünden
  threshold: number // benzer görsel eşiği (silme koruması bununla kontrol eder)
  loading: boolean
  selection: Selection
  setSelection: (s: Selection) => void
  toggle: (path: string) => void
  clearSelection: () => void
  // Dosya işlemlerinden sonra listeyi günceller (Faz 7).
  applyFileChanges: (removed: ReadonlySet<string>, renamed?: ReadonlyMap<string, string>) => void
  // Önbellek temizlenince sonuçlar kaldırılır: silme koruması onları artık
  // doğrulayamaz.
  resetResults: () => void
}

const ResultsContext = createContext<ResultsContextValue | null>(null)

// Sonuçlar ve seçim uygulama genelinde tutulur; sayfa değişince kaybolmaz.
export function ResultsProvider({ children }: { children: React.ReactNode }) {
  const { result } = useScan()
  const [groups, setGroups] = useState<DuplicateGroup[] | null>(null)
  const [source, setSource] = useState<ResultsSource | null>(null)
  const [mode, setMode] = useState<ScanMode>('exact')
  const [threshold, setThreshold] = useState(0)
  const [loading, setLoading] = useState(false)
  const [selection, setSelection] = useState<Selection>(new Set())

  // Yeni tamamlanan tarama önceki sonucun ve seçimin yerini alır.
  useEffect(() => {
    if (result?.status !== 'completed') return
    setGroups(result.groups)
    setSource('scan')
    setMode(result.mode)
    setThreshold(result.threshold)
    setSelection(new Set())
  }, [result])

  // Açılışta: son taranan klasörlerin kayıtlı sonuçlarını yükle. Veritabanı
  // yalnızca birebir kopya gruplarını verebildiği için son tarama benzer
  // görsel taramasıysa yüklenmez.
  useEffect(() => {
    const opts = readSavedOptions()
    if (!opts?.roots?.length || opts.mode === 'similar') return
    let cancelled = false
    setLoading(true)
    api
      .duplicateGroups(opts.roots, opts.minSize ?? 0)
      .then((g) => {
        // Bu arada bir tarama bittiyse onun sonucu geçerlidir.
        if (cancelled) return
        setGroups((cur) => cur ?? g)
        setSource((cur) => cur ?? 'saved')
      })
      .catch(() => {}) // kayıtlı sonuç yüklenemezse boş ekran yeterli
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [])

  const toggle = useCallback((path: string) => {
    setSelection((s) => {
      const next = new Set(s)
      if (next.has(path)) next.delete(path)
      else next.add(path)
      return next
    })
  }, [])

  const clearSelection = useCallback(() => setSelection(new Set()), [])

  const applyFileChanges = useCallback((removed: ReadonlySet<string>, renamed?: ReadonlyMap<string, string>) => {
    setGroups((g) => (g ? applyChanges(g, removed, renamed) : g))
    setSelection((s) => new Set([...s].filter((p) => !removed.has(p) && !renamed?.has(p))))
  }, [])

  const resetResults = useCallback(() => {
    setGroups(null)
    setSource(null)
    setSelection(new Set())
  }, [])

  const value = useMemo(
    () => ({
      groups, source, mode, threshold, loading, selection, setSelection, toggle, clearSelection, applyFileChanges,
      resetResults,
    }),
    [groups, source, mode, threshold, loading, selection, toggle, clearSelection, applyFileChanges, resetResults],
  )
  return <ResultsContext.Provider value={value}>{children}</ResultsContext.Provider>
}

export function useResults(): ResultsContextValue {
  const ctx = useContext(ResultsContext)
  if (!ctx) throw new Error('useResults, ResultsProvider içinde kullanılmalı')
  return ctx
}

function readSavedOptions(): Partial<ScanOptions> | null {
  try {
    return JSON.parse(localStorage.getItem(SCAN_OPTIONS_KEY) ?? 'null')
  } catch {
    return null
  }
}
