import { Button, IconButton, List, ListItem, ListItemIcon, ListItemText, Tooltip, Typography } from '@mui/material'
import FolderIcon from '@mui/icons-material/FolderOutlined'
import CloseIcon from '@mui/icons-material/Close'
import AddIcon from '@mui/icons-material/Add'
import { api } from '../api'
import { t } from '../i18n/core'
import { pathKey } from '../paths'

interface Props {
  folders: string[]
  onChange: (folders: string[]) => void
  addLabel: string
  dialogTitle: string
  emptyText: string
  disabled?: boolean
}

// Klasör listesi: diyalogla ekleme, satır sonundaki düğmeyle çıkarma.
export default function FolderList({ folders, onChange, addLabel, dialogTitle, emptyText, disabled }: Props) {
  const add = async () => {
    const path = await api.selectFolder(dialogTitle)
    if (!path) return
    // Windows yolları büyük/küçük harf duyarsız
    if (folders.some((f) => pathKey(f) === pathKey(path))) return
    onChange([...folders, path])
  }

  return (
    <>
      {folders.length === 0 ? (
        <Typography variant="body2" color="text.secondary" sx={{ py: 1 }}>
          {emptyText}
        </Typography>
      ) : (
        <List dense disablePadding sx={{ mb: 1 }}>
          {folders.map((f) => (
            <ListItem
              key={f}
              disableGutters
              secondaryAction={
                <Tooltip title={t('folder.remove')}>
                  <IconButton
                    edge="end"
                    size="small"
                    disabled={disabled}
                    onClick={() => onChange(folders.filter((x) => x !== f))}
                    aria-label={t('folder.removeAria', { path: f })}
                  >
                    <CloseIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              }
            >
              <ListItemIcon sx={{ minWidth: 36 }}>
                <FolderIcon color="primary" fontSize="small" />
              </ListItemIcon>
              <ListItemText
                primary={f}
                slotProps={{ primary: { sx: { wordBreak: 'break-all' } } }}
              />
            </ListItem>
          ))}
        </List>
      )}
      <Button startIcon={<AddIcon />} onClick={add} disabled={disabled}>
        {addLabel}
      </Button>
    </>
  )
}
