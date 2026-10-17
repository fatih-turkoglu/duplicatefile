import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'

// Büyük kütüphaneler ayrı parçalara bölünür; tek paket 500 kB uyarı sınırını aşmasın.
const vendorChunks: Record<string, string[]> = {
  react: ['react', 'react-dom', 'scheduler'],
  mui: ['@mui/material', '@mui/system', '@mui/utils', '@mui/styled-engine', '@mui/private-theming', '@mui/core-downloads-tracker', '@popperjs/core'],
  'mui-icons': ['@mui/icons-material'],
  emotion: ['@emotion'],
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    // macOS 12 ships with Safari 15's WKWebView. Vite's default targets 16.4+.
    target: 'safari15',
    cssTarget: 'safari15',
    rollupOptions: {
      output: {
        manualChunks(id) {
          const match = id.match(/node_modules[\\/]((?:@[^\\/]+[\\/])?[^\\/]+)/)
          if (!match) return
          const pkg = match[1].replace('\\', '/')
          for (const [chunk, pkgs] of Object.entries(vendorChunks)) {
            if (pkgs.some(p => pkg === p || pkg.startsWith(p + '/'))) return chunk
          }
          return 'vendor'
        },
      },
    },
  },
})
