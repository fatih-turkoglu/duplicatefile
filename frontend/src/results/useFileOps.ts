import { useCallback, useState } from 'react'
import { api, events, type OpsProgress, type OpsReport, type RenameItem, type Conflict } from '../api'
import { useResults } from './ResultsContext'

export type OpKind = 'delete' | 'move' | 'rename'

export interface OpOutcome {
  kind: OpKind
  permanent?: boolean
  report: OpsReport
}

// Dosya işlemlerini çalıştırır, ilerlemeyi izler ve başarılı olanları sonuç
// listesine yansıtır (silinenler çıkar, taşınan/yeniden adlandırılanların yolu
// güncellenir).
export function useFileOps() {
  const { applyFileChanges, mode, threshold } = useResults()
  const [running, setRunning] = useState<OpKind | null>(null)
  const [progress, setProgress] = useState<OpsProgress | null>(null)

  const run = useCallback(
    async (kind: OpKind, call: () => Promise<OpsReport>, permanent?: boolean): Promise<OpOutcome> => {
      setRunning(kind)
      setProgress(null)
      const off = events.onOpsProgress(setProgress)
      try {
        const report = await call()
        const ok = report.results.filter((r) => !r.error)
        if (kind === 'delete') {
          applyFileChanges(new Set(ok.map((r) => r.path)))
        } else {
          applyFileChanges(new Set(), new Map(ok.map((r) => [r.path, r.newPath!])))
        }
        return { kind, permanent, report }
      } finally {
        off()
        setRunning(null)
        setProgress(null)
      }
    },
    [applyFileChanges],
  )

  return {
    running,
    progress,
    // Benzer görsel sonuçlarında koruma, birebir kopya yerine taramanın
    // eşiğine göre benzer bir görselin kalmasını arar.
    deleteFiles: (paths: string[], permanent: boolean) =>
      run(
        'delete',
        () => api.deleteFiles(paths, { permanent, similar: mode === 'similar', threshold }),
        permanent,
      ),
    moveFiles: (paths: string[], dest: string, conflict: Conflict) =>
      run('move', () => api.moveFiles(paths, dest, conflict)),
    renameFiles: (items: RenameItem[]) => run('rename', () => api.renameFiles(items)),
  }
}
