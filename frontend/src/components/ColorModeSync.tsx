import { useEffect } from 'react'
import { useColorScheme } from '@mui/material/styles'
import {
  WindowSetDarkTheme,
  WindowSetLightTheme,
  WindowSetSystemDefaultTheme,
} from '../../wailsjs/runtime/runtime'

// Uygulamada seçilen temayı Windows pencere başlık çubuğuna da uygular.
export default function ColorModeSync() {
  const { mode } = useColorScheme()

  useEffect(() => {
    if (mode === 'dark') WindowSetDarkTheme()
    else if (mode === 'light') WindowSetLightTheme()
    else if (mode === 'system') WindowSetSystemDefaultTheme()
  }, [mode])

  return null
}
