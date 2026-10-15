import { useCallback, useDeferredValue, useMemo, useState } from 'react'
import { Alert, Button, Card, CardContent, LinearProgress, Snackbar, Typography } from '@mui/material'
import FileCopyIcon from '@mui/icons-material/FileCopyOutlined'
import { api, type DuplicateGroup } from '../api'
import CompareDialog from '../results/CompareDialog'
import EmptyState from '../components/EmptyState'
import { compareText } from '../format'
import { errorText, t } from '../i18n/core'
import type { Page } from '../layout/AppLayout'
import FolderSelectDialog from '../results/FolderSelectDialog'
import GroupList from '../results/GroupList'
import { useResults } from '../results/ResultsContext'
import ResultsToolbar from '../results/ResultsToolbar'
import RenameDialog from '../results/RenameDialog'
import SelectionActions from '../results/SelectionActions'
import { DeleteDialog, MoveDialog, OpsProgressDialog, OpsReportDialog, outcomeMessage } from '../results/OpsDialogs'
import { useFileOps, type OpOutcome } from '../results/useFileOps'
import {
  filterGroups,
  fullySelectedGroups,
  keepLargest,
  keepNewest,
  keepOldest,
  selectInFolder,
  selectionStats,
  sortGroups,
  wasted,
  type SortKey,
} from '../results/selection'

export default function ResultsPage({ onNavigate }: { onNavigate: (p: Page) => void }) {
  const { groups, source, mode, loading, selection, setSelection, toggle, clearSelection } = useResults()
  const [query, setQuery] = useState('')
  // Karşılaştırılan grup anahtarla tutulur: işlemlerden sonra güncel hâli
  // gösterilir, grup kaybolursa pencere kapanır.
  const [compareKey, setCompareKey] = useState<string | null>(null)
  const onCompare = useCallback((g: DuplicateGroup) => setCompareKey(g.hash), [])
  const [sort, setSort] = useState<SortKey>('wasted')
  const [folderDialog, setFolderDialog] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const deferredQuery = useDeferredValue(query)

  const all = groups ?? []
  const visible = useMemo(() => sortGroups(filterGroups(all, deferredQuery), sort), [all, deferredQuery, sort])
  const totals = useMemo(
    () => ({
      files: all.reduce((n, g) => n + g.files.length, 0),
      wasted: all.reduce((n, g) => n + wasted(g), 0),
    }),
    [all],
  )
  const selected = useMemo(() => selectionStats(all, selection), [all, selection])
  const dangerous = useMemo(() => fullySelectedGroups(all, selection).length, [all, selection])
  const compareGroup = compareKey ? (all.find((g) => g.similar && g.hash === compareKey) ?? null) : null

  const showInExplorer = useCallback((path: string) => {
    api.showInExplorer(path).catch((e) => setMessage(t('results.explorerFailed', { error: errorText(e) })))
  }, [])

  // ---- Dosya işlemleri
  const ops = useFileOps()
  const [dialog, setDialog] = useState<'delete' | 'move' | 'rename' | null>(null)
  const [outcome, setOutcome] = useState<OpOutcome | null>(null)
  // Sıralı: toplu yeniden adlandırmadaki {n} her seferinde aynı dosyaya düşsün.
  const selectedPaths = useMemo(() => [...selection].sort(compareText), [selection])

  const runOp = async (op: () => Promise<OpOutcome>) => {
    setDialog(null)
    try {
      const o = await op()
      if (o.report.failed > 0) setOutcome(o)
      else setMessage(outcomeMessage(o) + '.')
    } catch (e) {
      setMessage(t('results.opFailed', { error: errorText(e) }))
    }
  }

  const selectFolder = (folder: string) => {
    const { selection: next, skippedGroups } = selectInFolder(visible, selection, folder)
    const added = next.size - selection.size
    setSelection(next)
    let text = t('results.folderSelected', { count: Math.max(0, added) })
    if (skippedGroups > 0) {
      text += ' ' + t('results.folderSkipped', { count: skippedGroups })
    }
    setMessage(text)
  }

  // Her durumda çizilir: son grubu kaldıran bir işlemden sonra liste boşalsa da
  // sonuç bildirimi görünsün.
  const feedback = (
    <>
      <OpsProgressDialog kind={ops.running} progress={ops.progress} />
      <OpsReportDialog
        outcome={outcome}
        onClose={() => setOutcome(null)}
        onRetryPermanent={(paths) => {
          setOutcome(null)
          runOp(() => ops.deleteFiles(paths, true))
        }}
      />
      <Snackbar open={!!message} autoHideDuration={6000} onClose={() => setMessage(null)} message={message} />
    </>
  )

  if (loading && !groups) {
    return (
      <>
        <Typography variant="h5" gutterBottom>
          {t('results.title')}
        </Typography>
        <LinearProgress />
      </>
    )
  }

  if (all.length === 0) {
    return (
      <>
        <Typography variant="h5" gutterBottom>
          {t('results.title')}
        </Typography>
        <Card>
          <CardContent>
            <EmptyState
              icon={<FileCopyIcon />}
              title={t(groups ? 'results.none' : 'results.empty')}
              description={
                groups
                  ? t(mode === 'similar' ? 'results.noneSimilarText' : 'results.noneText')
                  : t('results.emptyText')
              }
              action={
                <Button variant="contained" onClick={() => onNavigate('scan')}>
                  {t('results.goScan')}
                </Button>
              }
            />
          </CardContent>
        </Card>
        {feedback}
      </>
    )
  }

  return (
    <>
      <Typography variant="h5" gutterBottom>
        {t('results.title')}
      </Typography>

      {source === 'saved' && (
        <Alert
          severity="info"
          sx={{ mb: 2 }}
          action={
            <Button color="inherit" size="small" onClick={() => onNavigate('scan')}>
              {t('results.rescan')}
            </Button>
          }
        >
          {t('results.saved')}
        </Alert>
      )}

      <ResultsToolbar
        totalGroups={all.length}
        totalFiles={totals.files}
        totalWasted={totals.wasted}
        visibleGroups={visible.length}
        filtered={deferredQuery.trim() !== ''}
        selectedFiles={selected.files}
        selectedBytes={selected.bytes}
        query={query}
        onQueryChange={setQuery}
        sort={sort}
        onSortChange={setSort}
        onKeepNewest={() => setSelection(keepNewest(visible, selection))}
        onKeepOldest={() => setSelection(keepOldest(visible, selection))}
        onKeepLargest={mode === 'similar' ? () => setSelection(keepLargest(visible, selection)) : undefined}
        onSelectFolder={() => setFolderDialog(true)}
        onClearSelection={clearSelection}
        actions={
          <SelectionActions
            disabled={selected.files === 0}
            onDelete={() => setDialog('delete')}
            onMove={() => setDialog('move')}
            onRename={() => setDialog('rename')}
          />
        }
      />

      {dangerous > 0 && (
        <Alert severity="warning" sx={{ mt: 2 }}>
          {t('results.dangerous', { count: dangerous })}
        </Alert>
      )}

      {visible.length === 0 ? (
        <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>
          {t('results.noMatch', { query: deferredQuery })}
        </Typography>
      ) : (
        <GroupList
          groups={visible}
          selection={selection}
          onToggle={toggle}
          onShowInExplorer={showInExplorer}
          onCompare={onCompare}
        />
      )}

      <CompareDialog
        group={compareGroup}
        selection={selection}
        onToggle={toggle}
        onClose={() => setCompareKey(null)}
      />

      <FolderSelectDialog
        open={folderDialog}
        groups={visible}
        onClose={() => setFolderDialog(false)}
        onSelect={selectFolder}
      />
      <DeleteDialog
        open={dialog === 'delete'}
        files={selected.files}
        bytes={selected.bytes}
        fullyGroups={dangerous}
        onClose={() => setDialog(null)}
        onConfirm={(permanent) => runOp(() => ops.deleteFiles(selectedPaths, permanent))}
      />
      <MoveDialog
        open={dialog === 'move'}
        files={selected.files}
        onClose={() => setDialog(null)}
        onConfirm={(dest, conflict) => runOp(() => ops.moveFiles(selectedPaths, dest, conflict))}
      />
      <RenameDialog
        open={dialog === 'rename'}
        paths={selectedPaths}
        onClose={() => setDialog(null)}
        onConfirm={(items) => runOp(() => ops.renameFiles(items))}
      />
      {feedback}
    </>
  )
}
