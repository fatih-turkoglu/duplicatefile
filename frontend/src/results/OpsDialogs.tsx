import { useEffect, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  FormControlLabel,
  LinearProgress,
  List,
  ListItem,
  ListItemText,
  Radio,
  RadioGroup,
  Typography,
} from '@mui/material'
import FolderOpenIcon from '@mui/icons-material/FolderOpenOutlined'
import { api, type Conflict, type OpErrorCode, type OpsProgress } from '../api'
import { formatBytes } from '../format'
import { opErrorText, t } from '../i18n/core'
import type { OpOutcome } from './useFileOps'

// ---------------------------------------------------------------- Silme

interface DeleteProps {
  open: boolean
  files: number
  bytes: number
  fullyGroups: number // bütün kopyaları seçili grup sayısı
  onClose: () => void
  onConfirm: (permanent: boolean) => void
}

export function DeleteDialog({ open, files, bytes, fullyGroups, onClose, onConfirm }: DeleteProps) {
  const [permanent, setPermanent] = useState(false)
  useEffect(() => {
    if (open) setPermanent(false) // her açılışta güvenli varsayılan
  }, [open])

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="xs">
      <DialogTitle>{t('delete.title', { count: files })}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{ mb: 2 }}>
          {t(permanent ? 'delete.permanentText' : 'delete.recycleText', { bytes: formatBytes(bytes) })}
        </DialogContentText>
        {fullyGroups > 0 && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            {t('delete.fullyWarning', { count: fullyGroups })}
          </Alert>
        )}
        <FormControlLabel
          control={<Checkbox checked={permanent} onChange={(e) => setPermanent(e.target.checked)} color="error" />}
          label={t('delete.permanentCheck')}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" color="error" onClick={() => onConfirm(permanent)}>
          {t(permanent ? 'delete.confirmPermanent' : 'delete.confirmRecycle')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

// ---------------------------------------------------------------- Taşıma

interface MoveProps {
  open: boolean
  files: number
  onClose: () => void
  onConfirm: (dest: string, conflict: Conflict) => void
}

export function MoveDialog({ open, files, onClose, onConfirm }: MoveProps) {
  const [dest, setDest] = useState('')
  const [conflict, setConflict] = useState<Conflict>('rename')

  const browse = async () => {
    const path = await api.selectFolder(t('move.dialog'))
    if (path) setDest(path)
  }

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>{t('move.title', { count: files })}</DialogTitle>
      <DialogContent>
        <Typography variant="body2" sx={{ fontWeight: 500, mb: 0.5 }}>
          {t('move.target')}
        </Typography>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2, flexWrap: 'wrap' }}>
          <Typography
            variant="body2"
            color={dest ? 'text.primary' : 'text.secondary'}
            sx={{ flex: '1 1 200px', wordBreak: 'break-all', fontFamily: dest ? 'Consolas, monospace' : undefined }}
          >
            {dest || t('move.notSelected')}
          </Typography>
          <Button startIcon={<FolderOpenIcon />} onClick={browse}>
            {t(dest ? 'move.change' : 'move.choose')}
          </Button>
        </Box>
        <Typography variant="body2" sx={{ fontWeight: 500 }}>
          {t('move.conflict')}
        </Typography>
        <RadioGroup value={conflict} onChange={(e) => setConflict(e.target.value as Conflict)}>
          <FormControlLabel value="rename" control={<Radio />} label={t('move.conflictRename')} />
          <FormControlLabel value="skip" control={<Radio />} label={t('move.conflictSkip')} />
        </RadioGroup>
        <Typography variant="caption" color="text.secondary">
          {t('move.note')}
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" disabled={!dest} onClick={() => onConfirm(dest, conflict)}>
          {t('move.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

// ---------------------------------------------------------------- İlerleme

const runningKeys = { delete: 'progress.delete', move: 'progress.move', rename: 'progress.rename' }

export function OpsProgressDialog({ kind, progress }: { kind: keyof typeof runningKeys | null; progress: OpsProgress | null }) {
  const percent = progress && progress.total > 0 ? (progress.done / progress.total) * 100 : null
  return (
    <Dialog open={kind !== null} fullWidth maxWidth="xs">
      <DialogTitle>{kind ? t(runningKeys[kind]) : ''}</DialogTitle>
      <DialogContent>
        <LinearProgress
          variant={percent === null ? 'indeterminate' : 'determinate'}
          value={percent ?? 0}
          sx={{ height: 8, borderRadius: 4, mb: 1 }}
        />
        <Typography variant="body2" color="text.secondary">
          {progress ? t('progress.files', { done: progress.done, total: progress.total }) : t('common.starting')}
        </Typography>
      </DialogContent>
    </Dialog>
  )
}

// ---------------------------------------------------------------- Rapor

export function outcomeMessage(o: OpOutcome): string {
  const count = o.report.succeeded
  switch (o.kind) {
    case 'delete':
      return t(o.permanent ? 'outcome.deletedPermanent' : 'outcome.recycled', { count })
    case 'move':
      return t('outcome.moved', { count })
    case 'rename':
      return t('outcome.renamed', { count })
  }
}

interface ReportProps {
  outcome: OpOutcome | null
  onClose: () => void
  onRetryPermanent: (paths: string[]) => void
}

// Başarısız dosya varsa gösterilir. Geri Dönüşüm Kutusu olmayan sürücülerdeki
// dosyalar için kalıcı silme seçeneği sunar.
export function OpsReportDialog({ outcome, onClose, onRetryPermanent }: ReportProps) {
  if (!outcome) return null
  const { report } = outcome
  const failed = report.results.filter((r) => r.error)
  const noRecycle = failed.filter((r) => (r.code as OpErrorCode) === 'no_recycle_bin').map((r) => r.path)
  const warnings = report.results.filter((r) => !r.error && r.warning)

  return (
    <Dialog open onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>{t('report.title', { ok: report.succeeded, failed: report.failed })}</DialogTitle>
      <DialogContent>
        {report.succeeded > 0 && <DialogContentText sx={{ mb: 2 }}>{outcomeMessage(outcome)}.</DialogContentText>}
        {noRecycle.length > 0 && (
          <Alert severity="info" sx={{ mb: 2 }}>
            {t('report.noRecycle', { count: noRecycle.length })}
          </Alert>
        )}
        <List dense disablePadding sx={{ maxHeight: 320, overflow: 'auto' }}>
          {failed.map((r) => (
            <ListItem key={r.path} disableGutters>
              <ListItemText
                primary={r.path}
                secondary={opErrorText(r.code, r.error)}
                slotProps={{
                  primary: { sx: { wordBreak: 'break-all' } },
                  secondary: { sx: { color: 'error.main' } },
                }}
              />
            </ListItem>
          ))}
        </List>
        {warnings.length > 0 && (
          <Typography variant="caption" color="text.secondary">
            {t('report.warnings', { count: warnings.length })}
          </Typography>
        )}
      </DialogContent>
      <DialogActions>
        {noRecycle.length > 0 && (
          <Button color="error" onClick={() => onRetryPermanent(noRecycle)}>
            {t('report.retryPermanent')}
          </Button>
        )}
        <Button variant="contained" onClick={onClose}>
          {t('common.ok')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
