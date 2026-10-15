import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  List,
  ListItem,
  ListItemText,
  Stack,
  Typography,
} from '@mui/material'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import PauseCircleIcon from '@mui/icons-material/PauseCircle'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import FileCopyIcon from '@mui/icons-material/FileCopyOutlined'
import type { ScanResult } from '../api'
import { formatBytes, formatDuration, formatNumber } from '../format'
import { has, t } from '../i18n/core'
import { wastedBytes } from './ScanContext'

interface Props {
  result: ScanResult
  onShowResults: () => void
}

export default function ScanSummaryCard({ result: r, onShowResults }: Props) {
  const completed = r.status === 'completed'
  const duplicateFiles = r.groups.reduce((n, g) => n + g.files.length, 0)

  const similar = r.mode === 'similar'
  const stats: { label: string; value: string; highlight?: boolean }[] = [
    {
      label: t(similar ? 'summary.groupsSimilar' : 'summary.groups'),
      value: formatNumber(r.groups.length),
      highlight: true,
    },
    { label: t('summary.wasted'), value: formatBytes(wastedBytes(r)), highlight: true },
    { label: t(similar ? 'summary.dupFilesSimilar' : 'summary.dupFiles'), value: formatNumber(duplicateFiles) },
    { label: t(similar ? 'summary.scannedSimilar' : 'summary.scanned'), value: formatNumber(r.filesSeen) },
    { label: t('summary.cached'), value: formatNumber(r.cacheHits) },
    { label: t('summary.read'), value: t('summary.readValue', { count: r.hashed, bytes: formatBytes(r.bytesRead) }) },
    { label: t('summary.duration'), value: formatDuration(r.durationMs) },
    { label: t('summary.removed'), value: formatNumber(r.removed) },
  ]

  return (
    <Card>
      <CardContent>
        <Stack spacing={2.5}>
          <Stack direction="row" spacing={1.5} sx={{ alignItems: 'center' }}>
            {completed ? <CheckCircleIcon color="success" /> : <PauseCircleIcon color="warning" />}
            <Typography variant="h6">{t(completed ? 'summary.completed' : 'summary.stopped')}</Typography>
          </Stack>

          {!completed && <Alert severity="info">{t('summary.stoppedInfo')}</Alert>}

          {completed && (
            <Box
              sx={{
                display: 'grid',
                gap: 1.5,
                gridTemplateColumns: { xs: 'repeat(2, 1fr)', sm: 'repeat(3, 1fr)', md: 'repeat(4, 1fr)' },
              }}
            >
              {stats.map((s) => (
                <Box
                  key={s.label}
                  sx={{
                    p: 1.5,
                    borderRadius: 2,
                    bgcolor: s.highlight ? 'action.selected' : 'action.hover',
                  }}
                >
                  <Typography variant="caption" color="text.secondary">
                    {s.label}
                  </Typography>
                  <Typography variant={s.highlight ? 'h6' : 'body1'} sx={{ fontWeight: 500 }}>
                    {s.value}
                  </Typography>
                </Box>
              ))}
            </Box>
          )}

          {r.errorCount > 0 && (
            <Accordion disableGutters variant="outlined" sx={{ borderRadius: 2, '&::before': { display: 'none' } }}>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Typography variant="body2" color="warning.main">
                  {t('summary.errors', { count: r.errorCount })}
                </Typography>
              </AccordionSummary>
              <AccordionDetails sx={{ maxHeight: 280, overflow: 'auto' }}>
                <List dense disablePadding>
                  {r.errors.map((e, i) => (
                    <ListItem key={i} disableGutters>
                      <ListItemText
                        primary={e.path}
                        secondary={e.code && has(`scanError.${e.code}`) ? t(`scanError.${e.code}`) : e.message}
                        slotProps={{ primary: { sx: { wordBreak: 'break-all' } } }}
                      />
                    </ListItem>
                  ))}
                </List>
                {r.errorCount > r.errors.length && (
                  <Typography variant="caption" color="text.secondary">
                    {t('summary.errorsTruncated', { count: r.errors.length })}
                  </Typography>
                )}
              </AccordionDetails>
            </Accordion>
          )}

          {completed && (
            <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
              <Button
                variant="contained"
                startIcon={<FileCopyIcon />}
                disabled={r.groups.length === 0}
                onClick={onShowResults}
                sx={{ width: { xs: '100%', sm: 'auto' } }}
              >
                {r.groups.length === 0
                  ? t(similar ? 'summary.noneSimilar' : 'summary.none')
                  : t(similar ? 'summary.showSimilar' : 'summary.show')}
              </Button>
            </Box>
          )}
        </Stack>
      </CardContent>
    </Card>
  )
}
