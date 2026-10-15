import { Alert, Stack, Typography } from '@mui/material'
import { useLocalStorage } from '../hooks/useLocalStorage'
import { errorText, t } from '../i18n/core'
import type { Page } from '../layout/AppLayout'
import { useScan } from '../scan/ScanContext'
import ScanOptionsForm, { defaultScanOptions, SCAN_OPTIONS_KEY } from '../scan/ScanOptionsForm'
import ScanProgressCard from '../scan/ScanProgressCard'
import ScanSummaryCard from '../scan/ScanSummaryCard'

export default function ScanPage({ onNavigate }: { onNavigate: (p: Page) => void }) {
  const { state, start, cancel, reset } = useScan()
  // Seçenekler (klasörler dahil) bir sonraki açılışta hatırlanır.
  const [options, setOptions] = useLocalStorage(SCAN_OPTIONS_KEY, defaultScanOptions)

  return (
    <>
      <Typography variant="h5" gutterBottom>
        {t('scan.title')}
      </Typography>

      {state.status === 'running' ? (
        <ScanProgressCard
          mode={state.mode}
          startedAt={state.startedAt}
          progress={state.progress}
          cancelling={state.cancelling}
          onCancel={cancel}
        />
      ) : (
        <Stack spacing={2}>
          {state.status === 'done' && (
            <ScanSummaryCard result={state.result} onShowResults={() => onNavigate('results')} />
          )}
          {state.status === 'error' && (
            <Alert severity="error" onClose={reset}>
              {t('scan.failed', { error: errorText(state.message) })}
            </Alert>
          )}
          <ScanOptionsForm options={options} onChange={setOptions} onStart={() => start(options)} />
        </Stack>
      )}
    </>
  )
}
