import { useMemo, useState } from 'react'
import {
  Autocomplete,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
  Typography,
} from '@mui/material'
import FolderOpenIcon from '@mui/icons-material/FolderOpenOutlined'
import { api, type DuplicateGroup } from '../api'
import { t } from '../i18n/core'
import { folderCounts } from './selection'

interface Props {
  open: boolean
  groups: DuplicateGroup[]
  onClose: () => void
  onSelect: (folder: string) => void
}

// Seçim için klasör sorar: sonuçlarda geçen klasörlerden aranarak ya da
// Windows diyaloğuyla seçilir.
export default function FolderSelectDialog({ open, groups, onClose, onSelect }: Props) {
  const folders = useMemo(() => (open ? folderCounts(groups) : []), [open, groups])
  const [folder, setFolder] = useState<string | null>(null)

  const browse = async () => {
    const path = await api.selectFolder(t('folderSelect.dialog'))
    if (path) setFolder(path)
  }

  const close = () => {
    setFolder(null)
    onClose()
  }

  return (
    <Dialog open={open} onClose={close} fullWidth maxWidth="sm">
      <DialogTitle>{t('folderSelect.title')}</DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          {t('folderSelect.text')}
        </Typography>
        <Autocomplete
          options={folders}
          value={folders.find((f) => f.folder === folder) ?? (folder ? { folder, count: 0 } : null)}
          onChange={(_, v) => setFolder(v?.folder ?? null)}
          getOptionLabel={(o) => o.folder}
          isOptionEqualToValue={(a, b) => a.folder === b.folder}
          renderOption={({ key, ...props }, o) => (
            <Box component="li" key={key} {...props} sx={{ gap: 2, justifyContent: 'space-between' }}>
              <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
                {o.folder}
              </Typography>
              <Typography variant="caption" color="text.secondary" sx={{ flexShrink: 0 }}>
                {t('folderSelect.files', { count: o.count })}
              </Typography>
            </Box>
          )}
          renderInput={(params) => (
            <TextField {...params} autoFocus label={t('folderSelect.label')} placeholder={t('folderSelect.placeholder')} />
          )}
        />
        <Button startIcon={<FolderOpenIcon />} onClick={browse} sx={{ mt: 1 }}>
          {t('folderSelect.browse')}
        </Button>
      </DialogContent>
      <DialogActions>
        <Button onClick={close}>{t('common.cancel')}</Button>
        <Button
          variant="contained"
          disabled={!folder}
          onClick={() => {
            onSelect(folder!)
            close()
          }}
        >
          {t('folderSelect.confirm')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
