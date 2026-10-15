import { useCallback, useEffect, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  LinearProgress,
  MenuItem,
  Skeleton,
  Snackbar,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from '@mui/material'
import { useColorScheme } from '@mui/material/styles'
import FolderOpenIcon from '@mui/icons-material/FolderOpenOutlined'
import DriveFileMoveIcon from '@mui/icons-material/DriveFileMoveOutlined'
import RestoreIcon from '@mui/icons-material/Restore'
import CompressIcon from '@mui/icons-material/Compress'
import CleaningIcon from '@mui/icons-material/CleaningServicesOutlined'
import DeleteSweepIcon from '@mui/icons-material/DeleteSweepOutlined'
import { api, type DatabaseInfo } from '../api'
import { themeModes, type ThemeMode } from '../components/ThemeModeButton'
import { formatBytes } from '../format'
import { useLocalStorage } from '../hooks/useLocalStorage'
import { errorText, t } from '../i18n/core'
import { useI18n } from '../i18n/I18nProvider'
import { AUTO } from '../i18n/select'
import { useResults } from '../results/ResultsContext'
import { defaultScanOptions, SCAN_OPTIONS_KEY } from '../scan/ScanOptionsForm'

const workerOptions = [
  { value: 0, label: () => t('settings.workers.auto') },
  { value: 1, label: () => t('settings.workers.hdd') },
  { value: 2, label: () => '2' },
  { value: 4, label: () => '4' },
  { value: 8, label: () => t('settings.workers.ssd') },
]

export default function SettingsPage() {
  const { mode, setMode } = useColorScheme()
  const i18n = useI18n()
  const { resetResults } = useResults()
  const [scanOptions, setScanOptions] = useLocalStorage(SCAN_OPTIONS_KEY, defaultScanOptions)

  const [info, setInfo] = useState<DatabaseInfo>()
  const [infoError, setInfoError] = useState<string>()
  const [working, setWorking] = useState<string | null>(null) // süren bakım işinin adı
  const [confirmClear, setConfirmClear] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  const refresh = useCallback(() => {
    api.databaseInfo().then(
      (i) => {
        setInfo(i)
        setInfoError(undefined)
      },
      (e) => setInfoError(errorText(e)),
    )
  }, [])
  useEffect(refresh, [refresh])

  const [appVersion, setAppVersion] = useState<string>()
  useEffect(() => {
    api.appInfo().then((i) => setAppVersion(i.version), () => {})
  }, [])

  // Bakım işlerini sırayla çalıştırır; bitince bilgileri yeniler.
  // taskKey: işin adının çeviri anahtarı (hata mesajında kullanılır).
  const run = async (taskKey: string, fn: () => Promise<string>) => {
    setWorking(taskKey)
    try {
      setMessage(await fn())
    } catch (e) {
      setMessage(t('settings.taskFailed', { task: t(taskKey), error: errorText(e) }))
    } finally {
      setWorking(null)
      refresh()
    }
  }

  const moveTo = async (dir: string) =>
    run('settings.task.move', async () => {
      const i = await api.moveDatabase(dir)
      setInfo(i)
      return t('settings.moved', { path: i.path })
    })

  const chooseFolder = async () => {
    const dir = await api.selectFolder(t('settings.moveDbDialog'))
    if (dir) moveTo(dir)
  }

  const autoName = i18n.options.find((o) => o.code === i18n.autoCode)?.name ?? 'English'

  return (
    <>
      <Typography variant="h5" gutterBottom>
        {t('settings.title')}
      </Typography>
      <Stack spacing={2}>
        <Card>
          <CardContent>
            <SectionTitle>{t('settings.appearance')}</SectionTitle>
            <ToggleButtonGroup
              exclusive
              size="small"
              value={mode ?? 'system'}
              onChange={(_, m: ThemeMode | null) => m && setMode(m)}
              aria-label={t('settings.theme')}
              sx={{ mb: 3, flexWrap: 'wrap' }}
            >
              {themeModes().map((m) => (
                <ToggleButton key={m.mode} value={m.mode} sx={{ gap: 1, px: 2 }}>
                  {m.icon}
                  {m.label}
                </ToggleButton>
              ))}
            </ToggleButtonGroup>

            <TextField
              select
              fullWidth
              label={t('settings.language')}
              value={i18n.preference}
              onChange={(e) => i18n.setPreference(e.target.value)}
              helperText={i18n.dir ? t('settings.languageHelp', { dir: i18n.dir }) : ' '}
              sx={{ maxWidth: { sm: 420 } }}
            >
              <MenuItem value={AUTO}>{t('settings.languageAuto', { name: autoName })}</MenuItem>
              {i18n.options.map((o) => (
                // Dil adları kendi dillerinde gösterilir (Deutsch, Français …).
                <MenuItem key={o.code} value={o.code} lang={o.code}>
                  {o.name}
                </MenuItem>
              ))}
            </TextField>
          </CardContent>
        </Card>

        <Card>
          <CardContent>
            <SectionTitle>{t('settings.scan')}</SectionTitle>
            <TextField
              select
              label={t('settings.workers')}
              value={scanOptions.workers}
              onChange={(e) => setScanOptions({ ...scanOptions, workers: Number(e.target.value) })}
              helperText={t('settings.workersHelp')}
              sx={{ maxWidth: { sm: 420 }, mb: 2 }}
              fullWidth
            >
              {workerOptions.map((w) => (
                <MenuItem key={w.value} value={w.value}>
                  {w.label()}
                </MenuItem>
              ))}
            </TextField>
            <Box>
              <Button
                startIcon={<RestoreIcon />}
                onClick={() => {
                  // Seçilen klasörler korunur; yalnızca seçenekler sıfırlanır.
                  setScanOptions({ ...defaultScanOptions, roots: scanOptions.roots })
                  setMessage(t('settings.resetDone'))
                }}
              >
                {t('settings.resetScan')}
              </Button>
            </Box>
          </CardContent>
        </Card>

        <Card>
          <CardContent>
            <SectionTitle>{t('settings.db')}</SectionTitle>
            {infoError && <Alert severity="error">{t('settings.dbOpenFailed', { error: infoError })}</Alert>}
            {info?.warning && (
              <Alert severity="warning" sx={{ mb: 2 }}>
                {t('settings.dbFallback', { error: info.warning })}
              </Alert>
            )}
            {info ? (
              <>
                <Typography
                  variant="body2"
                  sx={{ fontFamily: 'Consolas, monospace', wordBreak: 'break-all', mb: 0.5 }}
                >
                  {info.path}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  {t(info.isDefault ? 'settings.dbDefault' : 'settings.dbCustom')} ·{' '}
                  {t('settings.dbRecords', { count: info.stats.files })} ·{' '}
                  {t('settings.dbScans', { count: info.stats.scans })} · {formatBytes(info.stats.sizeBytes)}
                </Typography>
                <Stack direction="row" useFlexGap spacing={1} sx={{ flexWrap: 'wrap' }}>
                  <Button startIcon={<FolderOpenIcon />} onClick={() => api.showInExplorer(info.path)}>
                    {t('settings.showFolder')}
                  </Button>
                  <Button startIcon={<DriveFileMoveIcon />} onClick={chooseFolder} disabled={!!working}>
                    {t('settings.moveDb')}
                  </Button>
                  {!info.isDefault && (
                    <Button startIcon={<RestoreIcon />} onClick={() => moveTo('')} disabled={!!working}>
                      {t('settings.moveDefault')}
                    </Button>
                  )}
                </Stack>
              </>
            ) : (
              !infoError && <Skeleton width="60%" />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardContent>
            <SectionTitle>{t('settings.maintenance')}</SectionTitle>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              {t('settings.maintenanceText')}
            </Typography>
            {working && <LinearProgress sx={{ mb: 2 }} />}
            <Stack spacing={2}>
              <MaintenanceItem
                icon={<CompressIcon />}
                title={t('settings.compactTitle')}
                description={t('settings.compactText')}
                action={t('settings.compactAction')}
                disabled={!!working || !info}
                onClick={() =>
                  run('settings.task.compact', async () => {
                    const r = await api.compactDatabase()
                    return t('settings.compactDone', { before: formatBytes(r.before), after: formatBytes(r.after) })
                  })
                }
              />
              <MaintenanceItem
                icon={<CleaningIcon />}
                title={t('settings.pruneTitle')}
                description={t('settings.pruneText')}
                action={t('settings.pruneAction')}
                disabled={!!working || !info}
                onClick={() =>
                  run('settings.task.prune', async () => {
                    const n = await api.pruneDatabase()
                    return n > 0 ? t('settings.pruneDone', { count: n }) : t('settings.pruneNone')
                  })
                }
              />
              <MaintenanceItem
                icon={<DeleteSweepIcon />}
                title={t('settings.clearTitle')}
                description={t('settings.clearText')}
                action={t('settings.clearAction')}
                danger
                disabled={!!working || !info}
                onClick={() => setConfirmClear(true)}
              />
            </Stack>
          </CardContent>
        </Card>

        <Card>
          <CardContent>
            <SectionTitle>{t('settings.about')}</SectionTitle>
            <Typography variant="body2">
              DuplicateFile{' '}
              {appVersion
                ? appVersion === 'dev'
                  ? t('settings.devVersion')
                  : t('settings.version', { version: appVersion })
                : ''}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {t('app.tagline')}
            </Typography>
          </CardContent>
        </Card>
      </Stack>

      <Dialog open={confirmClear} onClose={() => setConfirmClear(false)} maxWidth="xs" fullWidth>
        <DialogTitle>{t('settings.clearConfirmTitle')}</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {info ? t('settings.clearConfirmCount', { count: info.stats.files }) + ' ' : ''}
            {t('settings.clearConfirmText')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmClear(false)}>{t('common.cancel')}</Button>
          <Button
            color="error"
            variant="contained"
            onClick={() => {
              setConfirmClear(false)
              run('settings.task.clear', async () => {
                await api.clearDatabase()
                resetResults()
                return t('settings.clearDone')
              })
            }}
          >
            {t('settings.clearConfirm')}
          </Button>
        </DialogActions>
      </Dialog>
      <Snackbar open={!!message} autoHideDuration={6000} onClose={() => setMessage(null)} message={message} />
    </>
  )
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <Typography variant="subtitle1" gutterBottom sx={{ fontWeight: 500 }}>
      {children}
    </Typography>
  )
}

interface ItemProps {
  icon: React.ReactElement
  title: string
  description: string
  action: string // düğme etiketi
  danger?: boolean
  disabled: boolean
  onClick: () => void
}

function MaintenanceItem({ icon, title, description, action, danger, disabled, onClick }: ItemProps) {
  return (
    <Stack direction={{ xs: 'column', sm: 'row' }} sx={{ gap: 1.5, alignItems: { sm: 'center' } }}>
      <Box sx={{ flex: 1 }}>
        <Typography variant="body2" sx={{ fontWeight: 500 }}>
          {title}
        </Typography>
        <Typography variant="caption" color="text.secondary">
          {description}
        </Typography>
      </Box>
      <Button
        variant="outlined"
        color={danger ? 'error' : 'primary'}
        startIcon={icon}
        disabled={disabled}
        onClick={onClick}
        sx={{ flexShrink: 0, alignSelf: { xs: 'flex-start', sm: 'center' } }}
      >
        {action}
      </Button>
    </Stack>
  )
}
