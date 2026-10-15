import { IconButton, Tooltip } from '@mui/material'
import { useColorScheme } from '@mui/material/styles'
import BrightnessAutoIcon from '@mui/icons-material/BrightnessAuto'
import LightModeIcon from '@mui/icons-material/LightMode'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import { t } from '../i18n/core'

export type ThemeMode = 'system' | 'light' | 'dark'

// Etiketler çizimde çevrilir (dil değişebilir).
export const themeModes = (): { mode: ThemeMode; label: string; icon: React.ReactElement }[] => [
  { mode: 'system', label: t('theme.system'), icon: <BrightnessAutoIcon /> },
  { mode: 'light', label: t('theme.light'), icon: <LightModeIcon /> },
  { mode: 'dark', label: t('theme.dark'), icon: <DarkModeIcon /> },
]

// Üst çubuktaki tema düğmesi: her tıklamada sistem → açık → koyu sırasıyla döner.
export default function ThemeModeButton() {
  const { mode, setMode } = useColorScheme()
  if (!mode) return null

  const modes = themeModes()
  const i = modes.findIndex((m) => m.mode === mode)
  const current = modes[i]
  const next = modes[(i + 1) % modes.length]

  return (
    <Tooltip title={t('theme.tooltip', { current: current.label, next: next.label })}>
      <IconButton color="inherit" onClick={() => setMode(next.mode)} aria-label={t('theme.change')}>
        {current.icon}
      </IconButton>
    </Tooltip>
  )
}
