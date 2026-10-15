import React, { useMemo } from 'react'
import { createRoot } from 'react-dom/client'
import CssBaseline from '@mui/material/CssBaseline'
import { ThemeProvider } from '@mui/material/styles'
import './fonts.css'
import { createAppTheme } from './theme'
import { locale } from './i18n/core'
import { I18nProvider } from './i18n/I18nProvider'
import ColorModeSync from './components/ColorModeSync'
import ErrorBoundary from './components/ErrorBoundary'
import App from './App'
import { ScanProvider } from './scan/ScanContext'
import { ResultsProvider } from './results/ResultsContext'

// I18nProvider dil değişince içini yeniden kurar; tema da o anki dile göre
// (MUI bileşen metinleri) oluşturulur.
function Themed() {
  const theme = useMemo(() => createAppTheme(locale()), [])
  return (
    <ThemeProvider theme={theme} defaultMode="system" disableTransitionOnChange>
      <CssBaseline />
      <ColorModeSync />
      <ErrorBoundary>
        <App />
      </ErrorBoundary>
    </ThemeProvider>
  )
}

// Tarama ve sonuç durumu dil sağlayıcısının dışındadır: dil değişince
// arayüz yeniden kurulsa da süren tarama ve sonuç listesi kaybolmaz.
createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ScanProvider>
      <ResultsProvider>
        <I18nProvider>
          <Themed />
        </I18nProvider>
      </ResultsProvider>
    </ScanProvider>
  </React.StrictMode>,
)
