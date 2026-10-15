import { useState } from 'react'
import {
  Box,
  Button,
  Divider,
  InputAdornment,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import ChecklistIcon from '@mui/icons-material/Checklist'
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown'
import NewReleasesIcon from '@mui/icons-material/NewReleasesOutlined'
import HistoryIcon from '@mui/icons-material/History'
import FolderIcon from '@mui/icons-material/FolderOutlined'
import ClearIcon from '@mui/icons-material/Clear'
import HighQualityIcon from '@mui/icons-material/HighQualityOutlined'
import { formatBytes, formatNumber } from '../format'
import { t } from '../i18n/core'
import type { SortKey } from './selection'

interface Props {
  totalGroups: number
  totalFiles: number
  totalWasted: number
  visibleGroups: number
  filtered: boolean
  selectedFiles: number
  selectedBytes: number
  query: string
  onQueryChange: (q: string) => void
  sort: SortKey
  onSortChange: (s: SortKey) => void
  onKeepNewest: () => void
  onKeepOldest: () => void
  onKeepLargest?: () => void // yalnızca benzer görsel sonuçlarında
  onSelectFolder: () => void
  onClearSelection: () => void
  actions?: React.ReactNode // Faz 7: dosya işlemi düğmeleri
}

// Etiketler (çeviri anahtarları) çizimde çevrilir.
const sorts: { value: SortKey; label: string }[] = [
  { value: 'wasted', label: 'sort.wasted' },
  { value: 'size', label: 'sort.size' },
  { value: 'count', label: 'sort.count' },
]

// Sonuç ekranının üstte sabit kalan çubuğu: özet, arama, sıralama, seçim.
export default function ResultsToolbar(p: Props) {
  const [menuAnchor, setMenuAnchor] = useState<HTMLElement | null>(null)
  const closeMenu = () => setMenuAnchor(null)
  const run = (fn: () => void) => () => {
    closeMenu()
    fn()
  }

  return (
    <Box
      sx={{
        position: 'sticky',
        top: 0,
        zIndex: 2,
        bgcolor: 'background.default',
        pt: 1,
        pb: 1.5,
        borderBottom: 1,
        borderColor: 'divider',
      }}
    >
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        sx={{ justifyContent: 'space-between', alignItems: { sm: 'center' }, gap: 1, mb: 1.5 }}
      >
        <Typography variant="body2" color="text.secondary">
          {p.filtered && `${formatNumber(p.visibleGroups)} / `}
          {t('toolbar.groups', { count: p.totalGroups })} · {t('toolbar.files', { count: p.totalFiles })} ·{' '}
          <Box component="span" sx={{ color: 'text.primary', fontWeight: 500 }}>
            {t('toolbar.wasted', { bytes: formatBytes(p.totalWasted) })}
          </Box>
        </Typography>
        {p.selectedFiles > 0 && (
          <Typography variant="body2" sx={{ color: 'primary.main', fontWeight: 500 }}>
            {t('toolbar.selected', { count: p.selectedFiles, bytes: formatBytes(p.selectedBytes) })}
          </Typography>
        )}
      </Stack>

      <Stack direction="row" useFlexGap sx={{ flexWrap: 'wrap', gap: 1, alignItems: 'center' }}>
        <TextField
          size="small"
          placeholder={t('toolbar.search')}
          value={p.query}
          onChange={(e) => p.onQueryChange(e.target.value)}
          sx={{ flex: '1 1 220px' }}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" />
                </InputAdornment>
              ),
            },
          }}
        />
        <TextField
          select
          size="small"
          label={t('toolbar.sort')}
          value={p.sort}
          onChange={(e) => p.onSortChange(e.target.value as SortKey)}
          sx={{ flex: '0 1 200px', minWidth: 160 }}
        >
          {sorts.map((s) => (
            <MenuItem key={s.value} value={s.value}>
              {t(s.label)}
            </MenuItem>
          ))}
        </TextField>
        <Button
          variant="outlined"
          startIcon={<ChecklistIcon />}
          endIcon={<ArrowDropDownIcon />}
          onClick={(e) => setMenuAnchor(e.currentTarget)}
        >
          {t('select.button')}
        </Button>
        {p.actions}
      </Stack>

      <Menu anchorEl={menuAnchor} open={!!menuAnchor} onClose={closeMenu}>
        {p.filtered && (
          <Typography variant="caption" color="text.secondary" sx={{ px: 2, pb: 1, display: 'block' }}>
            {t('select.filteredNote')}
          </Typography>
        )}
        {p.onKeepLargest && (
          <MenuItem onClick={run(p.onKeepLargest)}>
            <ListItemIcon>
              <HighQualityIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText primary={t('select.keepLargest')} secondary={t('select.keepLargestHelp')} />
          </MenuItem>
        )}
        <MenuItem onClick={run(p.onKeepNewest)}>
          <ListItemIcon>
            <NewReleasesIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary={t('select.keepNewest')} secondary={t('select.othersHelp')} />
        </MenuItem>
        <MenuItem onClick={run(p.onKeepOldest)}>
          <ListItemIcon>
            <HistoryIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary={t('select.keepOldest')} secondary={t('select.othersHelp')} />
        </MenuItem>
        <MenuItem onClick={run(p.onSelectFolder)}>
          <ListItemIcon>
            <FolderIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary={t('select.inFolder')} />
        </MenuItem>
        <Divider />
        <MenuItem onClick={run(p.onClearSelection)} disabled={p.selectedFiles === 0}>
          <ListItemIcon>
            <ClearIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary={t('select.clear')} />
        </MenuItem>
      </Menu>
    </Box>
  )
}
