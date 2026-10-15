import { useEffect, useState } from 'react'
import {
  Box,
  Button,
  Card,
  CardContent,
  LinearProgress,
  Stack,
  Step,
  StepLabel,
  Stepper,
  Typography,
  useMediaQuery,
} from '@mui/material'
import { useTheme } from '@mui/material/styles'
import StopIcon from '@mui/icons-material/Stop'
import type { ScanMode, ScanProgress, ScanStage } from '../api'
import { formatBytes, formatElapsed, formatPercent, shortenPath } from '../format'
import { t } from '../i18n/core'

// Etiketler (çeviri anahtarları) çizimde çevrilir.
const stagesByMode: Record<ScanMode, { stage: ScanStage; label: string }[]> = {
  exact: [
    { stage: 'walk', label: 'scan.stage.walk' },
    { stage: 'partial', label: 'scan.stage.partial' },
    { stage: 'full', label: 'scan.stage.full' },
    { stage: 'cleanup', label: 'scan.stage.cleanup' },
    { stage: 'save', label: 'scan.stage.save' },
  ],
  similar: [
    { stage: 'walk', label: 'scan.stage.walkImages' },
    { stage: 'image', label: 'scan.stage.image' },
    { stage: 'cleanup', label: 'scan.stage.cleanup' },
    { stage: 'save', label: 'scan.stage.save' },
  ],
}

interface Props {
  mode: ScanMode
  startedAt: number
  progress?: ScanProgress
  cancelling: boolean
  onCancel: () => void
}

export default function ScanProgressCard({ mode, startedAt, progress, cancelling, onCancel }: Props) {
  const theme = useTheme()
  const wide = useMediaQuery(theme.breakpoints.up('sm'))
  const elapsed = useElapsed(startedAt)
  const stages = stagesByMode[mode]

  const stageIndex = Math.max(0, stages.findIndex((s) => s.stage === progress?.stage))
  const percent = progressPercent(progress)

  return (
    <Card>
      <CardContent>
        <Stack spacing={3}>
          {wide ? (
            <Stepper activeStep={stageIndex} alternativeLabel>
              {stages.map((s, i) => (
                <Step key={s.stage} active={i === stageIndex} completed={i < stageIndex}>
                  <StepLabel>{t(s.label)}</StepLabel>
                </Step>
              ))}
            </Stepper>
          ) : (
            <Typography variant="subtitle1" sx={{ fontWeight: 500 }}>
              {t('scan.progress.step', { n: stageIndex + 1, total: stages.length, label: t(stages[stageIndex].label) })}
            </Typography>
          )}

          <Box>
            {/* key: belirsiz ↔ belirli geçişinde çubuk eski animasyon konumunda kalmasın */}
            <LinearProgress
              key={`${progress?.stage}-${percent === null}`}
              variant={percent === null ? 'indeterminate' : 'determinate'}
              value={percent ?? 0}
              sx={{ height: 8, borderRadius: 4 }}
            />
            <Stack direction="row" sx={{ justifyContent: 'space-between', mt: 1, gap: 2 }}>
              <Typography variant="body2">{progressText(progress)}</Typography>
              <Typography variant="body2" color="text.secondary" sx={{ flexShrink: 0 }}>
                {percent !== null && `${formatPercent(Math.floor(percent))} · `}
                {formatElapsed(elapsed)}
              </Typography>
            </Stack>
          </Box>

          <Typography
            variant="caption"
            color="text.secondary"
            title={progress?.current}
            sx={{
              fontFamily: 'Consolas, monospace',
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              minHeight: '1.5em',
            }}
          >
            {progress?.current ? shortenPath(progress.current, wide ? 110 : 50) : ' '}
          </Typography>

          <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Button
              variant="outlined"
              color="error"
              startIcon={<StopIcon />}
              onClick={onCancel}
              disabled={cancelling}
              sx={{ width: { xs: '100%', sm: 'auto' } }}
            >
              {cancelling ? t('scan.stopping') : t('scan.stop')}
            </Button>
          </Box>
          <Typography variant="caption" color="text.secondary">
            {t('scan.stopHint')}
          </Typography>
        </Stack>
      </CardContent>
    </Card>
  )
}

// null: belirsiz ilerleme (dosya sayısı henüz bilinmiyor).
function progressPercent(p?: ScanProgress): number | null {
  if (!p || p.total === 0) return null
  if (p.totalBytes > 0) return (p.bytes / p.totalBytes) * 100
  return (p.done / p.total) * 100
}

function progressText(p?: ScanProgress): string {
  if (!p) return t('common.starting')
  if (p.stage === 'walk') return t('scan.progress.found', { count: p.done })
  let text = t('scan.progress.files', { done: p.done, total: p.total })
  if (p.totalBytes > 0) text += ` · ${formatBytes(p.bytes)} / ${formatBytes(p.totalBytes)}`
  return text
}

function useElapsed(since: number): number {
  const [now, setNow] = useState(Date.now())
  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [])
  return Math.max(0, now - since)
}
