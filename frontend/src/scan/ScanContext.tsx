import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { api, events, type ScanMode, type ScanOptions, type ScanProgress, type ScanResult } from '../api'
import { wasted } from '../results/selection'

export type ScanState =
  | { status: 'idle' }
  | { status: 'running'; mode: ScanMode; startedAt: number; progress?: ScanProgress; cancelling: boolean }
  | { status: 'done'; result: ScanResult }
  | { status: 'error'; message: string }

interface ScanContextValue {
  state: ScanState
  // Son tamamlanan taramanın sonucu; yeni tarama bitene kadar korunur.
  result: ScanResult | null
  start: (opts: ScanOptions) => Promise<void>
  cancel: () => void
  reset: () => void
}

const ScanContext = createContext<ScanContextValue | null>(null)

// Tarama durumu uygulama genelinde tutulur: sayfa değişse de tarama ve
// ilerleme kaybolmaz. Go'dan gelen olaylara yalnızca burada abone olunur.
export function ScanProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<ScanState>({ status: 'idle' })
  const [result, setResult] = useState<ScanResult | null>(null)

  useEffect(() => {
    const offs = [
      events.onScanProgress((progress) =>
        setState((s) => (s.status === 'running' ? { ...s, progress } : s)),
      ),
      events.onScanDone((raw) => {
        // Go'da boş dilim JSON'da null olabilir; arayüz hep dizi bekler.
        const r = { ...raw, groups: raw.groups ?? [], errors: raw.errors ?? [], roots: raw.roots ?? [] }
        setResult(r)
        setState({ status: 'done', result: r })
      }),
      events.onScanError((message) => setState({ status: 'error', message })),
    ]
    // Arayüz yeniden yüklendiyse süren taramayı yakala.
    api.isScanning().then((running) => {
      // Mod bilinmiyor; yalnızca aşama etiketlerini etkiler.
      if (running) setState({ status: 'running', mode: 'exact', startedAt: Date.now(), cancelling: false })
    })
    return () => offs.forEach((off) => off())
  }, [])

  const start = useCallback(async (opts: ScanOptions) => {
    setState({ status: 'running', mode: opts.mode, startedAt: Date.now(), cancelling: false })
    try {
      await api.startScan(opts)
    } catch (e) {
      setState({ status: 'error', message: String(e) })
    }
  }, [])

  const cancel = useCallback(() => {
    setState((s) => (s.status === 'running' ? { ...s, cancelling: true } : s))
    api.cancelScan()
  }, [])

  const reset = useCallback(() => setState({ status: 'idle' }), [])

  const value = useMemo(() => ({ state, result, start, cancel, reset }), [state, result, start, cancel, reset])
  return <ScanContext.Provider value={value}>{children}</ScanContext.Provider>
}

export function useScan(): ScanContextValue {
  const ctx = useContext(ScanContext)
  if (!ctx) throw new Error('useScan, ScanProvider içinde kullanılmalı')
  return ctx
}

// Bir grubun gereksiz kopyalarının kapladığı alan (birini tutup gerisini silince kazanılan).
export function wastedBytes(r: ScanResult): number {
  return r.groups.reduce((sum, g) => sum + wasted(g), 0)
}
