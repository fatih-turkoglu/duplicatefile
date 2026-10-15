import { memo, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { Box, Button, Checkbox, Chip, IconButton, Stack, Tooltip, Typography, useMediaQuery } from '@mui/material'
import { useTheme } from '@mui/material/styles'
import FolderOpenIcon from '@mui/icons-material/FolderOpenOutlined'
import CompareIcon from '@mui/icons-material/CompareOutlined'
import { useVirtualizer } from '@tanstack/react-virtual'
import type { DuplicateGroup, FileEntry } from '../api'
import Thumb from '../components/Thumb'
import { formatBytes, formatDate } from '../format'
import { t } from '../i18n/core'
import { useScrollContainer } from '../layout/scrollContainer'
import { fileName, folderOf, selectedCount, wasted, type Selection } from './selection'

type Row =
  | { kind: 'header'; key: string; group: DuplicateGroup }
  | { kind: 'file'; key: string; group: DuplicateGroup; file: FileEntry; last: boolean }

const GAP = 12 // gruplar arası boşluk (başlık satırının üstünde)
const HEADER_H = 56 + GAP

// Satır yükseklikleri: benzer görsellerde büyük küçük resim ve fazladan bilgi satırı.
function fileRowHeight(similar: boolean, wide: boolean): number {
  if (similar) return wide ? 76 : 92
  return wide ? 56 : 76
}

interface Props {
  groups: DuplicateGroup[]
  selection: Selection
  onToggle: (path: string) => void
  onShowInExplorer: (path: string) => void
  onCompare: (group: DuplicateGroup) => void
}

// Grupları "başlık + dosya satırları" olarak düzleştirip yalnızca görünen
// satırları çizer. Liste kendi kaydırma kutusunu açmaz; sayfanın ana kaydırma
// alanını kullanır.
export default function GroupList({ groups, selection, onToggle, onShowInExplorer, onCompare }: Props) {
  const theme = useTheme()
  const wide = useMediaQuery(theme.breakpoints.up('sm'))

  const rows = useMemo(() => {
    const out: Row[] = []
    for (const g of groups) {
      out.push({ kind: 'header', key: `h:${g.size}:${g.hash}`, group: g })
      g.files.forEach((f, i) =>
        out.push({ kind: 'file', key: `f:${f.path}`, group: g, file: f, last: i === g.files.length - 1 }),
      )
    }
    return out
  }, [groups])

  const scrollRef = useScrollContainer()
  const listRef = useRef<HTMLDivElement>(null)
  const [scrollMargin, setScrollMargin] = useState(0)
  // Listenin kaydırma alanı içindeki konumu (üstündeki araç çubuğu vb. kadar).
  useLayoutEffect(() => {
    const top = listRef.current?.offsetTop ?? 0
    if (top !== scrollMargin) setScrollMargin(top)
  })

  const virtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: (i) => {
      const r = rows[i]
      return r.kind === 'header' ? HEADER_H : fileRowHeight(!!r.group.similar, wide)
    },
    getItemKey: (i) => rows[i].key,
    overscan: 12,
    scrollMargin,
  })

  // Satır yükseklikleri satırlar ya da ekran genişliği değişince yeniden hesaplanır.
  useEffect(() => {
    virtualizer.measure()
  }, [rows, wide, virtualizer])

  return (
    <Box ref={listRef} sx={{ position: 'relative', height: virtualizer.getTotalSize() }}>
      {virtualizer.getVirtualItems().map((item) => {
        const row = rows[item.index]
        return (
          <Box
            key={item.key}
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: item.size,
              transform: `translateY(${item.start - scrollMargin}px)`,
            }}
          >
            {row.kind === 'header' ? (
              <GroupHeader group={row.group} selected={selectedCount(row.group, selection)} onCompare={onCompare} />
            ) : (
              <FileRow
                file={row.file}
                similar={!!row.group.similar}
                last={row.last}
                wide={wide}
                checked={selection.has(row.file.path)}
                onToggle={onToggle}
                onShowInExplorer={onShowInExplorer}
              />
            )}
          </Box>
        )
      })}
    </Box>
  )
}

interface HeaderProps {
  group: DuplicateGroup
  selected: number
  onCompare: (group: DuplicateGroup) => void
}

const GroupHeader = memo(function GroupHeader({ group: g, selected, onCompare }: HeaderProps) {
  const all = selected === g.files.length
  const total = g.files.reduce((n, f) => n + f.size, 0)
  return (
    <Stack
      direction="row"
      sx={{
        mt: `${GAP}px`,
        height: HEADER_H - GAP,
        px: 2,
        alignItems: 'center',
        gap: 1.5,
        bgcolor: 'action.hover',
        border: 1,
        borderColor: 'divider',
        borderTopLeftRadius: 12,
        borderTopRightRadius: 12,
      }}
    >
      <Box sx={{ flex: 1, minWidth: 0 }}>
        <Typography variant="subtitle2" noWrap>
          {g.similar
            ? t('group.similar', { count: g.files.length, size: formatBytes(total) })
            : t('group.exact', { count: g.files.length, size: formatBytes(g.size) })}
        </Typography>
        <Typography variant="caption" color="text.secondary" noWrap component="div">
          {t('group.wasted', { bytes: formatBytes(wasted(g)) })}
        </Typography>
      </Box>
      {selected > 0 && (
        <Tooltip title={all ? t('group.allSelectedTip') : ''}>
          <Chip
            size="small"
            color={all ? 'error' : 'primary'}
            variant={all ? 'filled' : 'outlined'}
            label={all ? t('group.allSelected') : t('group.selected', { selected, total: g.files.length })}
          />
        </Tooltip>
      )}
      {g.similar && (
        <Button size="small" startIcon={<CompareIcon />} onClick={() => onCompare(g)} sx={{ flexShrink: 0 }}>
          {t('group.compare')}
        </Button>
      )}
    </Stack>
  )
})

interface FileRowProps {
  file: FileEntry
  similar: boolean
  last: boolean
  wide: boolean
  checked: boolean
  onToggle: (path: string) => void
  onShowInExplorer: (path: string) => void
}

const FileRow = memo(function FileRow({ file, similar, last, wide, checked, onToggle, onShowInExplorer }: FileRowProps) {
  const date = formatDate(file.modTime)
  const resolution = file.width && file.height ? `${file.width}×${file.height}` : ''
  // Benzer görsellerde karar için gereken bilgi: çözünürlük, boyut, benzerlik.
  const meta = similar
    ? [resolution, formatBytes(file.size), file.similarity ? t('file.similarity', { percent: file.similarity }) : ''].filter(Boolean).join(' · ')
    : ''

  return (
    <Stack
      direction="row"
      onClick={() => onToggle(file.path)}
      sx={{
        height: '100%',
        pl: 0.5,
        pr: 1,
        alignItems: 'center',
        gap: 1,
        cursor: 'pointer',
        bgcolor: checked ? 'action.selected' : 'background.paper',
        borderLeft: 1,
        borderRight: 1,
        borderBottom: 1,
        borderColor: 'divider',
        ...(last && { borderBottomLeftRadius: 12, borderBottomRightRadius: 12 }),
        '&:hover': { bgcolor: checked ? 'action.selected' : 'action.hover' },
      }}
    >
      <Checkbox
        checked={checked}
        onClick={(e) => e.stopPropagation()}
        onChange={() => onToggle(file.path)}
        slotProps={{ input: { 'aria-label': t('file.select', { path: file.path }) } }}
      />
      <Thumb path={file.path} box={similar ? 60 : 40} />
      <Box sx={{ flex: 1, minWidth: 0 }} title={file.path}>
        <Typography variant="body2" noWrap sx={{ fontWeight: 500 }}>
          {fileName(file.path)}
        </Typography>
        <Typography variant="caption" color="text.secondary" noWrap component="div">
          {folderOf(file.path)}
        </Typography>
        {meta && (
          <Typography variant="caption" noWrap component="div">
            {meta}
          </Typography>
        )}
        {!wide && (
          <Typography variant="caption" color="text.secondary" noWrap component="div">
            {date}
          </Typography>
        )}
      </Box>
      {wide && (
        <Typography variant="body2" color="text.secondary" noWrap sx={{ flexShrink: 0 }}>
          {date}
        </Typography>
      )}
      <Tooltip title={t('file.showInExplorer')}>
        <IconButton
          size="small"
          onClick={(e) => {
            e.stopPropagation()
            onShowInExplorer(file.path)
          }}
          aria-label={t('file.showInExplorerAria', { path: file.path })}
        >
          <FolderOpenIcon fontSize="small" />
        </IconButton>
      </Tooltip>
    </Stack>
  )
})
