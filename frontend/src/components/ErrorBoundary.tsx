import { Component, type ErrorInfo, type ReactNode } from 'react'
import { Alert, AlertTitle, Box, Button, Typography } from '@mui/material'
import { t } from '../i18n/core'

interface State {
  error: Error | null
}

// Arayüzde beklenmeyen bir hata olursa React bütün ağacı kaldırır ve pencere
// boş kalır. Bunun yerine hatayı ve yeniden yükleme düğmesini gösterir.
export default class ErrorBoundary extends Component<{ children: ReactNode }, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Arayüz hatası:', error, info.componentStack)
  }

  render() {
    const { error } = this.state
    if (!error) return this.props.children
    return (
      <Box sx={{ p: 3, maxWidth: 720, mx: 'auto' }}>
        <Alert
          severity="error"
          action={
            <Button color="inherit" onClick={() => window.location.reload()}>
              {t('crash.reload')}
            </Button>
          }
        >
          <AlertTitle>{t('crash.title')}</AlertTitle>
          {t('crash.text')}
          <Typography
            component="pre"
            variant="caption"
            sx={{ mt: 1, whiteSpace: 'pre-wrap', wordBreak: 'break-all', fontFamily: 'Consolas, monospace' }}
          >
            {error.message}
          </Typography>
        </Alert>
      </Box>
    )
  }
}
