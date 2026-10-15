import { Button, IconButton, Tooltip, useMediaQuery } from '@mui/material'
import { useTheme } from '@mui/material/styles'
import DeleteIcon from '@mui/icons-material/DeleteOutlined'
import MoveIcon from '@mui/icons-material/DriveFileMoveOutlined'
import RenameIcon from '@mui/icons-material/DriveFileRenameOutline'
import { t } from '../i18n/core'

interface Props {
  disabled: boolean
  onDelete: () => void
  onMove: () => void
  onRename: () => void
}

// Seçili dosyalar için işlem düğmeleri; dar ekranda yalnızca simge.
export default function SelectionActions({ disabled, onDelete, onMove, onRename }: Props) {
  const theme = useTheme()
  const wide = useMediaQuery(theme.breakpoints.up('md'))
  const actions = [
    { label: t('actions.delete'), icon: <DeleteIcon />, onClick: onDelete, color: 'error' as const },
    { label: t('actions.move'), icon: <MoveIcon />, onClick: onMove, color: 'primary' as const },
    { label: t('actions.rename'), icon: <RenameIcon />, onClick: onRename, color: 'primary' as const },
  ]

  if (!wide) {
    return (
      <>
        {actions.map((a) => (
          <Tooltip key={a.label} title={a.label}>
            <span>
              <IconButton color={a.color} disabled={disabled} onClick={a.onClick} aria-label={a.label}>
                {a.icon}
              </IconButton>
            </span>
          </Tooltip>
        ))}
      </>
    )
  }
  return (
    <>
      {actions.map((a) => (
        <Button
          key={a.label}
          variant={a.color === 'error' ? 'contained' : 'outlined'}
          color={a.color}
          startIcon={a.icon}
          disabled={disabled}
          onClick={a.onClick}
        >
          {a.label}
        </Button>
      ))}
    </>
  )
}
