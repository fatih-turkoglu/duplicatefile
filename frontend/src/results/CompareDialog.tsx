import {
  Box,
  Button,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
  useMediaQuery,
} from '@mui/material'
import { useTheme } from '@mui/material/styles'
import type { DuplicateGroup } from '../api'
import Thumb from '../components/Thumb'
import { formatBytes, formatDate } from '../format'
import { t } from '../i18n/core'
import { fileName, folderOf, type Selection } from './selection'

interface Props {
  group: DuplicateGroup | null
  selection: Selection
  onToggle: (path: string) => void
  onClose: () => void
}

// Benzer görsel grubunu büyük önizlemelerle yan yana gösterir; hangisinin
// tutulacağına karar vermek için seçim burada da yapılabilir.
export default function CompareDialog({ group, selection, onToggle, onClose }: Props) {
  const theme = useTheme()
  const fullScreen = useMediaQuery(theme.breakpoints.down('sm'))
  if (!group) return null
  const selected = group.files.filter((f) => selection.has(f.path)).length
  const all = selected === group.files.length

  return (
    <Dialog open onClose={onClose} fullWidth maxWidth="lg" fullScreen={fullScreen}>
      <DialogTitle>
        {t('compare.title', { count: group.files.length })}
        <Typography variant="body2" color={all ? 'error' : 'text.secondary'}>
          {all
            ? t('compare.allSelected')
            : t('compare.selectedInfo', { count: selected })}
        </Typography>
      </DialogTitle>
      <DialogContent dividers>
        <Box sx={{ display: 'grid', gap: 2, gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))' }}>
          {group.files.map((f, i) => {
            const checked = selection.has(f.path)
            return (
              <Box
                key={f.path}
                onClick={() => onToggle(f.path)}
                sx={{
                  border: 2,
                  borderColor: checked ? 'primary.main' : 'divider',
                  borderRadius: 3,
                  overflow: 'hidden',
                  cursor: 'pointer',
                  bgcolor: checked ? 'action.selected' : 'background.paper',
                }}
              >
                <Box sx={{ position: 'relative', bgcolor: 'action.hover', display: 'grid', placeItems: 'center', p: 1 }}>
                  <Thumb path={f.path} box={220} fit="contain" />
                  {i === 0 && (
                    <Chip
                      size="small"
                      color="secondary"
                      label={t('compare.highestRes')}
                      sx={{ position: 'absolute', top: 8, left: 8 }}
                    />
                  )}
                </Box>
                <Stack direction="row" sx={{ alignItems: 'flex-start', p: 1, gap: 0.5 }}>
                  <Checkbox
                    checked={checked}
                    onClick={(e) => e.stopPropagation()}
                    onChange={() => onToggle(f.path)}
                    slotProps={{ input: { 'aria-label': t('compare.select', { path: f.path }) } }}
                    sx={{ mt: -0.5 }}
                  />
                  <Box sx={{ minWidth: 0, flex: 1 }} title={f.path}>
                    <Typography variant="body2" noWrap sx={{ fontWeight: 500 }}>
                      {fileName(f.path)}
                    </Typography>
                    <Typography variant="caption" color="text.secondary" noWrap component="div">
                      {folderOf(f.path)}
                    </Typography>
                    <Typography variant="caption" component="div">
                      {f.width && f.height ? `${f.width}×${f.height} · ` : ''}
                      {formatBytes(f.size)}
                      {f.similarity ? ' · ' + t('file.similarity', { percent: f.similarity }) : ''}
                    </Typography>
                    <Typography variant="caption" color="text.secondary" component="div">
                      {formatDate(f.modTime)}
                    </Typography>
                  </Box>
                </Stack>
              </Box>
            )
          })}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button variant="contained" onClick={onClose}>
          {t('common.close')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
