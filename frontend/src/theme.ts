import { createTheme } from '@mui/material/styles'
import { bgBG, deDE, enUS, frFR, ptBR, ptPT, roRO, trTR } from '@mui/material/locale'
import { isMac } from './platform'

// MUI'nin kendi bileşen metinleri (ör. "Seçenek yok") için dil paketleri.
// Dil dosyasının @locale değerine göre seçilir; bilinmeyen dilde İngilizce.
const muiLocales: Record<string, typeof enUS> = {
  'pt-pt': ptPT,
  bg: bgBG,
  de: deDE,
  en: enUS,
  fr: frFR,
  pt: ptBR,
  ro: roRO,
  tr: trTR,
}

function muiLocale(locale: string) {
  const tag = locale.toLowerCase()
  return muiLocales[tag] ?? muiLocales[tag.split('-')[0]] ?? enUS
}

// Açık ve koyu şemalar CSS değişkenleriyle tanımlanır; mod değişince sayfa
// yeniden çizilmeden yalnızca <html> sınıfı değişir. Seçilen mod
// (sistem/açık/koyu) MUI tarafından localStorage'da saklanır.
export function createAppTheme(locale: string) {
  return createTheme(
    {
      cssVariables: { colorSchemeSelector: 'class' },
      colorSchemes: {
        light: {
          palette: {
            primary: { main: '#3949ab' },
            secondary: { main: '#00897b' },
            background: { default: '#f5f6fa', paper: '#ffffff' },
          },
        },
        dark: {
          palette: {
            primary: { main: '#9fa8da' },
            secondary: { main: '#80cbc4' },
            background: { default: '#121318', paper: '#1c1d23' },
          },
        },
      },
      shape: { borderRadius: 12 },
      typography: {
        fontFamily: isMac ? '-apple-system, BlinkMacSystemFont, system-ui, sans-serif' : 'Roboto, "Segoe UI", system-ui, sans-serif',
        button: { textTransform: 'none', fontWeight: 500 },
      },
      components: {
        // WebView2'nin varsayılan oklu kaydırma çubuğu yerine ince, temaya uyan çubuk.
        MuiCssBaseline: {
          styleOverrides: {
            '*': {
              scrollbarWidth: 'thin',
              scrollbarColor: 'var(--mui-palette-text-disabled) transparent',
            },
          },
        },
        MuiButton: { defaultProps: { disableElevation: true } },
        MuiCard: { defaultProps: { variant: 'outlined' } },
        MuiAppBar: { defaultProps: { elevation: 0, color: 'inherit' } },
        MuiTooltip: { defaultProps: { arrow: true } },
      },
    },
    muiLocale(locale),
  )
}
