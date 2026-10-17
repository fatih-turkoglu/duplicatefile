import { useEffect, useMemo, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Tooltip,
  Typography,
} from '@mui/material'
import ArrowForwardIcon from '@mui/icons-material/ArrowForward'
import type { RenameItem } from '../api'
import { t } from '../i18n/core'
import { planRenames, splitName, TEMPLATE_TOKENS, validateName, type RenameError } from './rename'
import { fileName } from './selection'
import { isWindowsPath } from '../paths'

interface Props {
  open: boolean
  paths: string[]
  onClose: () => void
  onConfirm: (items: RenameItem[]) => void
}

const PREVIEW_LIMIT = 50

// Tek dosyada doğrudan yeni ad, birden çok dosyada şablon ve önizleme.
export default function RenameDialog({ open, paths, onClose, onConfirm }: Props) {
  const single = paths.length === 1
  const [name, setName] = useState('')
  const [template, setTemplate] = useState(() => t('rename.defaultTemplate'))

  useEffect(() => {
    if (open && single) setName(fileName(paths[0]))
  }, [open, single, paths])

  const plan = useMemo(() => (open && !single ? planRenames(paths, template) : []), [open, single, paths, template])
  const planErrors = plan.filter((p) => p.error).length

  const singleError = single
    ? (validateName(name, isWindowsPath(paths[0])) ?? (name === fileName(paths[0]) ? ({ code: 'unchanged' } as RenameError) : null))
    : null
  const canConfirm = single ? !singleError : planErrors === 0 && plan.length > 0

  const confirm = () => {
    if (single) onConfirm([{ path: paths[0], newName: name }])
    else onConfirm(plan.map((p) => ({ path: p.path, newName: p.newName })))
  }

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>{single ? t('rename.titleOne') : t('rename.titleMany', { count: paths.length })}</DialogTitle>
      <DialogContent>
        {single ? (
          <TextField
            autoFocus
            fullWidth
            margin="dense"
            label={t('rename.newName')}
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && canConfirm && confirm()}
            error={!!singleError && name !== fileName(paths[0])}
            helperText={name !== fileName(paths[0]) && singleError ? errorMessage(singleError) : ' '}
            // Açılışta yalnızca uzantısız kısım seçili olsun (Gezgin gibi).
            onFocus={(e) => e.target.setSelectionRange(0, splitName(e.target.value).stem.length)}
          />
        ) : (
          <>
            <TextField
              autoFocus
              fullWidth
              margin="dense"
              label={t('rename.template')}
              value={template}
              onChange={(e) => setTemplate(e.target.value)}
            />
            <Stack direction="row" useFlexGap spacing={1} sx={{ flexWrap: 'wrap', my: 1 }}>
              {TEMPLATE_TOKENS.map((tok) => (
                <Tooltip key={tok.token} title={t(tok.help)}>
                  <Chip
                    size="small"
                    variant="outlined"
                    label={`${tok.token} ${t(tok.label)}`}
                    onClick={() => setTemplate((s) => s + tok.token)}
                  />
                </Tooltip>
              ))}
            </Stack>
            {planErrors > 0 && (
              <Alert severity="error" sx={{ my: 1 }}>
                {t('rename.invalidCount', { count: planErrors })}
              </Alert>
            )}
            <Typography variant="body2" sx={{ fontWeight: 500, mt: 2, mb: 1 }}>
              {t('rename.preview')}
            </Typography>
            <Box sx={{ maxHeight: 280, overflow: 'auto', border: 1, borderColor: 'divider', borderRadius: 2 }}>
              {plan.slice(0, PREVIEW_LIMIT).map((p) => (
                <Box key={p.path} sx={{ px: 1.5, py: 1, borderBottom: 1, borderColor: 'divider' }}>
                  <Stack direction="row" spacing={1} sx={{ alignItems: 'center', minWidth: 0 }}>
                    <Typography variant="body2" noWrap sx={{ flex: 1, minWidth: 0 }} title={p.path}>
                      {fileName(p.path)}
                    </Typography>
                    <ArrowForwardIcon fontSize="small" color="action" />
                    <Typography
                      variant="body2"
                      noWrap
                      sx={{ flex: 1, minWidth: 0, fontWeight: 500, color: p.error ? 'error.main' : 'text.primary' }}
                    >
                      {p.newName}
                    </Typography>
                  </Stack>
                  {p.error && (
                    <Typography variant="caption" color="error">
                      {errorMessage(p.error)}
                    </Typography>
                  )}
                </Box>
              ))}
              {plan.length > PREVIEW_LIMIT && (
                <Typography variant="caption" color="text.secondary" sx={{ display: 'block', p: 1.5 }}>
                  {t('rename.more', { count: plan.length - PREVIEW_LIMIT })}
                </Typography>
              )}
            </Box>
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('common.cancel')}</Button>
        <Button variant="contained" disabled={!canConfirm} onClick={confirm}>
          {t('rename.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

function errorMessage(e: RenameError): string {
  return t(`rename.error.${e.code}`, e.name ? { name: e.name } : undefined)
}
