import { useState } from 'react'
import { Box } from '@mui/material'
import FileIcon from '@mui/icons-material/InsertDriveFileOutlined'
import BrokenImageIcon from '@mui/icons-material/BrokenImageOutlined'
import { isImage, thumbUrl } from '../api'

interface Props {
  path: string
  box: number // ekrandaki kare boyutu (px)
  fit?: 'cover' | 'contain'
}

// Görseller için Go'dan gelen küçük resim; diğer dosyalarda (ya da çözülemeyen
// görselde) simge. Yüksek DPI ekranlar için kutunun iki katı istenir.
export default function Thumb({ path, box, fit = 'cover' }: Props) {
  const [failed, setFailed] = useState(false)
  const frame = {
    width: box,
    height: box,
    flexShrink: 0,
    borderRadius: 1.5,
    bgcolor: 'action.hover',
  }

  if (!isImage(path) || failed) {
    return (
      <Box sx={{ ...frame, display: 'grid', placeItems: 'center', color: 'text.secondary' }}>
        {failed ? <BrokenImageIcon fontSize="small" /> : <FileIcon fontSize="small" />}
      </Box>
    )
  }
  return (
    <Box
      component="img"
      src={thumbUrl(path, Math.min(512, Math.ceil(box * 2)))}
      alt=""
      loading="lazy"
      decoding="async"
      draggable={false}
      onError={() => setFailed(true)}
      sx={{ ...frame, objectFit: fit, display: 'block' }}
    />
  )
}
