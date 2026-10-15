import {
  Autocomplete,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  FormControlLabel,
  MenuItem,
  Stack,
  Slider,
  Switch,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from '@mui/material'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import FileCopyIcon from '@mui/icons-material/FileCopyOutlined'
import ImageSearchIcon from '@mui/icons-material/ImageSearch'
import type { ScanMode, ScanOptions } from '../api'
import { formatBytes } from '../format'
import { t } from '../i18n/core'
import FolderList from './FolderList'

// Tarama seçeneklerinin localStorage anahtarı (sonuç ekranı da okur).
export const SCAN_OPTIONS_KEY = 'scanOptions'

export const defaultScanOptions: ScanOptions = {
  roots: [],
  minSize: 1024,
  extensions: [],
  excludeDirs: [],
  skipHidden: true,
  workers: 0,
  mode: 'exact',
  threshold: 8,
}

const minSizeValues = [0, 1024, 100 * 1024, 1024 ** 2, 10 * 1024 ** 2, 100 * 1024 ** 2]

const extPresets = [
  { key: 'scan.ext.images', exts: ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'heic', 'tif', 'tiff'] },
  { key: 'scan.ext.videos', exts: ['mp4', 'mkv', 'avi', 'mov', 'wmv', 'webm', 'm4v'] },
  { key: 'scan.ext.music', exts: ['mp3', 'flac', 'wav', 'aac', 'm4a', 'ogg', 'wma'] },
  { key: 'scan.ext.documents', exts: ['pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt', 'odt'] },
  { key: 'scan.ext.archives', exts: ['zip', 'rar', '7z', 'tar', 'gz', 'iso'] },
]

// "  .JPG " -> "jpg"; boş ve tekrar edenler atılır.
function normalizeExts(list: string[]): string[] {
  const out = list.map((e) => e.trim().toLowerCase().replace(/^\.+/, '')).filter(Boolean)
  return [...new Set(out)]
}

interface Props {
  options: ScanOptions
  onChange: (o: ScanOptions) => void
  onStart: () => void
}

export default function ScanOptionsForm({ options, onChange, onStart }: Props) {
  const set = <K extends keyof ScanOptions>(key: K, value: ScanOptions[K]) => onChange({ ...options, [key]: value })
  const thresholdMarks = [
    { value: 2, label: t('scan.threshold.strict') },
    { value: 8, label: t('scan.threshold.similar') },
    { value: 16, label: t('scan.threshold.loose') },
  ]

  return (
    <Stack spacing={2}>
      <Card>
        <CardContent>
          <SectionTitle>{t('scan.roots.title')}</SectionTitle>
          <FolderList
            folders={options.roots}
            onChange={(roots) => set('roots', roots)}
            addLabel={t('scan.roots.add')}
            dialogTitle={t('scan.roots.dialog')}
            emptyText={t('scan.roots.empty')}
          />
        </CardContent>
      </Card>

      <Card>
        <CardContent>
          <SectionTitle>{t('scan.mode.title')}</SectionTitle>
          <ToggleButtonGroup
            exclusive
            fullWidth
            value={options.mode}
            onChange={(_, m: ScanMode | null) => m && set('mode', m)}
            sx={{ maxWidth: { sm: 520 } }}
          >
            <ToggleButton value="exact" sx={{ gap: 1 }}>
              <FileCopyIcon fontSize="small" />
              {t('scan.mode.exact')}
            </ToggleButton>
            <ToggleButton value="similar" sx={{ gap: 1 }}>
              <ImageSearchIcon fontSize="small" />
              {t('scan.mode.similar')}
            </ToggleButton>
          </ToggleButtonGroup>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1.5 }}>
            {options.mode === 'exact' ? t('scan.mode.exactHelp') : t('scan.mode.similarHelp')}
          </Typography>
          {options.mode === 'similar' && (
            // px: kenardaki işaret etiketleri taşmasın
            <Box sx={{ mt: 2, px: 4, maxWidth: 520 }}>
              <Typography variant="body2" sx={{ fontWeight: 500 }} id="threshold-label">
                {t('scan.threshold.label', { value: options.threshold })}
              </Typography>
              <Slider
                aria-labelledby="threshold-label"
                min={2}
                max={16}
                step={1}
                marks={thresholdMarks}
                value={options.threshold}
                onChange={(_, v) => set('threshold', v as number)}
              />
              <Typography variant="caption" color="text.secondary">
                {t('scan.threshold.help')}
              </Typography>
            </Box>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardContent>
          <SectionTitle>{t('scan.options.title')}</SectionTitle>
          <Stack spacing={3}>
            <TextField
              select
              label={t('scan.minSize.label')}
              value={options.minSize}
              onChange={(e) => set('minSize', Number(e.target.value))}
              helperText={t('scan.minSize.help')}
              sx={{ maxWidth: { sm: 320 } }}
            >
              {minSizeValues.map((v) => (
                <MenuItem key={v} value={v}>
                  {v === 0 ? t('scan.minSize.all') : t('scan.minSize.atLeast', { size: formatBytes(v) })}
                </MenuItem>
              ))}
            </TextField>

            <Box>
              <Autocomplete
                multiple
                freeSolo
                options={[]}
                value={options.extensions}
                onChange={(_, v) => set('extensions', normalizeExts(v as string[]))}
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label={t('scan.ext.label')}
                    placeholder={options.extensions.length ? '' : t('scan.ext.placeholder')}
                    helperText={t('scan.ext.help')}
                  />
                )}
              />
              <Stack direction="row" useFlexGap spacing={1} sx={{ flexWrap: 'wrap', mt: 1 }}>
                {extPresets.map((p) => (
                  <Chip
                    key={p.key}
                    label={`+ ${t(p.key)}`}
                    variant="outlined"
                    size="small"
                    onClick={() => set('extensions', normalizeExts([...options.extensions, ...p.exts]))}
                  />
                ))}
                {options.extensions.length > 0 && (
                  <Chip label={t('scan.ext.clear')} size="small" onClick={() => set('extensions', [])} />
                )}
              </Stack>
            </Box>

            <FormControlLabel
              control={<Switch checked={options.skipHidden} onChange={(e) => set('skipHidden', e.target.checked)} />}
              label={t('scan.skipHidden')}
            />

            <Box>
              <Typography variant="body2" sx={{ fontWeight: 500, mb: 0.5 }}>
                {t('scan.exclude.title')}
              </Typography>
              <FolderList
                folders={options.excludeDirs}
                onChange={(dirs) => set('excludeDirs', dirs)}
                addLabel={t('scan.exclude.add')}
                dialogTitle={t('scan.exclude.dialog')}
                emptyText={t('scan.exclude.empty')}
              />
            </Box>
          </Stack>
        </CardContent>
      </Card>

      <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
        <Button
          variant="contained"
          size="large"
          startIcon={<PlayArrowIcon />}
          disabled={options.roots.length === 0}
          onClick={onStart}
          sx={{ width: { xs: '100%', sm: 'auto' } }}
        >
          {t('scan.start')}
        </Button>
      </Box>
    </Stack>
  )
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <Typography variant="subtitle1" sx={{ fontWeight: 500, mb: 1 }}>
      {children}
    </Typography>
  )
}
