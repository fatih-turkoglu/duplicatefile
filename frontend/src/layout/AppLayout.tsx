import {
  AppBar,
  Badge,
  BottomNavigation,
  BottomNavigationAction,
  Box,
  Container,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Paper,
  Toolbar,
  Typography,
  useMediaQuery,
} from '@mui/material'
import { useTheme } from '@mui/material/styles'
import TravelExploreIcon from '@mui/icons-material/TravelExplore'
import FileCopyIcon from '@mui/icons-material/FileCopyOutlined'
import SettingsIcon from '@mui/icons-material/SettingsOutlined'
import { useEffect, useRef } from 'react'
import ThemeModeButton from '../components/ThemeModeButton'
import { ScrollContainerContext } from './scrollContainer'
import { t } from '../i18n/core'
import { useScan } from '../scan/ScanContext'

export type Page = 'scan' | 'results' | 'settings'

type NavItem = { page: Page; label: string; icon: React.ReactElement }

// Etiketler çizimde çevrilir (dil değişebilir).
const baseNavItems = (): NavItem[] => [
  { page: 'scan', label: t('nav.scan'), icon: <TravelExploreIcon /> },
  { page: 'results', label: t('nav.results'), icon: <FileCopyIcon /> },
  { page: 'settings', label: t('nav.settings'), icon: <SettingsIcon /> },
]

interface Props {
  page: Page
  onNavigate: (page: Page) => void
  children: React.ReactNode
}

// Pencere genişliğine göre gezinme:
//   lg ve üstü → etiketli yan menü
//   md         → dar yan şerit (simge + küçük etiket)
//   sm ve altı → alt gezinme çubuğu
export default function AppLayout({ page, onNavigate, children }: Props) {
  const theme = useTheme()
  const isMd = useMediaQuery(theme.breakpoints.up('md'))
  const isLg = useMediaQuery(theme.breakpoints.up('lg'))
  const scanning = useScan().state.status === 'running'
  const mainRef = useRef<HTMLElement>(null)

  // Yeni sayfa en üstten başlasın.
  useEffect(() => {
    mainRef.current?.scrollTo({ top: 0 })
  }, [page])

  // Tarama sürerken başka sayfadaysak "Tara" simgesinde nokta göster.
  const navItems = baseNavItems().map((n) =>
    n.page === 'scan' && scanning && page !== 'scan'
      ? { ...n, icon: <Badge variant="dot" color="secondary">{n.icon}</Badge> }
      : n,
  )

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      <AppBar position="static" sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Toolbar sx={{ gap: 1.5 }}>
          <FileCopyIcon color="primary" />
          <Typography variant="h6" component="h1" noWrap sx={{ flexGrow: 1 }}>
            DuplicateFile
          </Typography>
          <ThemeModeButton />
        </Toolbar>
      </AppBar>

      <Box sx={{ display: 'flex', flex: 1, minHeight: 0 }}>
        {isMd && <SideNav page={page} onNavigate={onNavigate} compact={!isLg} navItems={navItems} />}
        <Box component="main" ref={mainRef} sx={{ flex: 1, overflow: 'auto', position: 'relative' }}>
          <Container maxWidth="lg" sx={{ py: { xs: 2, sm: 3 } }}>
            <ScrollContainerContext.Provider value={mainRef}>{children}</ScrollContainerContext.Provider>
          </Container>
        </Box>
      </Box>

      {!isMd && (
        <Paper square elevation={0} sx={{ borderTop: 1, borderColor: 'divider' }}>
          <BottomNavigation showLabels value={page} onChange={(_, p: Page) => onNavigate(p)}>
            {navItems.map((n) => (
              <BottomNavigationAction key={n.page} value={n.page} label={n.label} icon={n.icon} />
            ))}
          </BottomNavigation>
        </Paper>
      )}
    </Box>
  )
}

function SideNav({
  page,
  onNavigate,
  compact,
  navItems,
}: Omit<Props, 'children'> & { compact: boolean; navItems: NavItem[] }) {
  return (
    <Box
      component="nav"
      sx={{
        width: compact ? 88 : 232,
        flexShrink: 0,
        borderRight: 1,
        borderColor: 'divider',
        bgcolor: 'background.paper',
        overflowY: 'auto',
      }}
    >
      <List sx={{ px: compact ? 1 : 1.5, py: 1.5 }}>
        {navItems.map((n) => (
          <ListItemButton
            key={n.page}
            selected={page === n.page}
            onClick={() => onNavigate(n.page)}
            sx={{
              borderRadius: compact ? 3 : 10,
              mb: 0.5,
              ...(compact && { flexDirection: 'column', py: 1, gap: 0.5 }),
              '&.Mui-selected': { color: 'primary.main' },
              '&.Mui-selected .MuiListItemIcon-root': { color: 'primary.main' },
            }}
          >
            <ListItemIcon sx={{ minWidth: compact ? 0 : 40 }}>{n.icon}</ListItemIcon>
            <ListItemText
              primary={n.label}
              slotProps={{ primary: compact ? { variant: 'caption' } : { sx: { fontWeight: 500 } } }}
              sx={{ m: 0 }}
            />
          </ListItemButton>
        ))}
      </List>
    </Box>
  )
}
