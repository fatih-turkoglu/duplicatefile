import { chromium } from 'playwright-core'
import fs from 'node:fs'

// Use an installed browser on either development platform; no Playwright download.
export async function launchBrowser() {
  const executablePath = [
    process.env.DUPLICATEFILE_BROWSER,
    '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    '/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge',
    'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
    'C:/Program Files/Microsoft/Edge/Application/msedge.exe',
    'C:/Program Files/Google/Chrome/Application/chrome.exe',
  ].find((file) => file && fs.existsSync(file))
  if (!executablePath) throw new Error('Install Chrome/Edge or set DUPLICATEFILE_BROWSER to the browser executable.')
  return chromium.launch({ executablePath, headless: true })
}
