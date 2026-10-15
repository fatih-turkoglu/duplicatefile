import { useState } from 'react'
import AppLayout, { type Page } from './layout/AppLayout'
import ScanPage from './pages/ScanPage'
import ResultsPage from './pages/ResultsPage'
import SettingsPage from './pages/SettingsPage'

const PAGE_KEY = 'page'
const pages: Page[] = ['scan', 'results', 'settings']

// Açık sayfa oturum boyunca saklanır: dil değişince arayüz yeniden kurulsa da
// kullanıcı aynı sayfada kalır.
function initialPage(): Page {
  try {
    const p = sessionStorage.getItem(PAGE_KEY) as Page | null
    return p && pages.includes(p) ? p : 'scan'
  } catch {
    return 'scan'
  }
}

export default function App() {
  const [page, setPageState] = useState<Page>(initialPage)
  const setPage = (p: Page) => {
    setPageState(p)
    try {
      sessionStorage.setItem(PAGE_KEY, p)
    } catch {
      // saklanamazsa yalnızca yeniden kurulumda Tara'ya döner
    }
  }

  return (
    <AppLayout page={page} onNavigate={setPage}>
      {page === 'scan' && <ScanPage onNavigate={setPage} />}
      {page === 'results' && <ResultsPage onNavigate={setPage} />}
      {page === 'settings' && <SettingsPage />}
    </AppLayout>
  )
}
