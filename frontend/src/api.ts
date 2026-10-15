// Go tarafıyla konuşan tek katman. Bileşenler wailsjs'i doğrudan değil bu
// modül üzerinden kullanır; böylece tipler ve olay adları tek yerde durur.
import * as App from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { db, fileops, main, scanner } from '../wailsjs/go/models'

export type DatabaseInfo = main.DatabaseInfo

export type FileEntry = db.File
export type DuplicateGroup = db.DuplicateGroup
export type OpsReport = fileops.Report
export type OpResult = fileops.Result
export type RenameItem = { path: string; newName: string }
// fileops.Result.Code ile aynı tutulmalı.
export type OpErrorCode = 'exists' | 'invalid_name' | 'no_recycle_bin' | 'cancelled' | 'last_copy' | 'unverified' | 'error'
export type Conflict = 'skip' | 'rename'

export type ScanMode = 'exact' | 'similar'

export interface ScanOptions {
  roots: string[]
  minSize: number
  extensions: string[]
  excludeDirs: string[]
  skipHidden: boolean
  workers: number
  mode: ScanMode
  threshold: number // benzer görsel eşiği: 64 bitten en fazla kaçı farklı olabilir
}

export interface DeleteOptions {
  permanent: boolean
  similar: boolean
  threshold: number
}

// Aşağıdaki tipler Go'daki karşılıklarıyla (internal/scanner) aynı tutulmalı;
// olaylarla geldikleri için Wails bunları otomatik üretmez.
export type ScanStage = 'walk' | 'partial' | 'full' | 'image' | 'cleanup' | 'save'

export interface ScanProgress {
  stage: ScanStage
  done: number
  total: number
  bytes: number
  totalBytes: number
  current: string
}

export interface ScanError {
  path: string
  message: string
  code?: string // bilinen hatalarda: changed, image_unreadable, image_too_large
}

export interface ScanResult {
  scanId: number
  status: 'completed' | 'cancelled'
  mode: ScanMode
  threshold: number
  roots: string[]
  filesSeen: number
  cacheHits: number
  hashed: number
  bytesRead: number
  removed: number
  groups: DuplicateGroup[]
  errors: ScanError[]
  errorCount: number
  durationMs: number
}

export interface OpsProgress {
  done: number
  total: number
}

export const api = {
  appInfo: () => App.AppInfo(),
  languageSetup: () => App.LanguageSetup(),
  loadLanguage: (code: string) => App.LoadLanguage(code),
  databaseInfo: () => App.DatabaseInfo(),
  moveDatabase: (dir: string) => App.MoveDatabase(dir), // "" = varsayılan konum
  compactDatabase: () => App.CompactDatabase(),
  pruneDatabase: () => App.PruneDatabase(),
  clearDatabase: () => App.ClearDatabase(),

  selectFolder: (title: string) => App.SelectFolder(title),
  showInExplorer: (path: string) => App.ShowInExplorer(path),

  startScan: (opts: ScanOptions) => App.StartScan(scanner.Options.createFrom(opts)),
  cancelScan: () => App.CancelScan(),
  isScanning: () => App.IsScanning(),
  duplicateGroups: (roots: string[], minSize: number) =>
    App.DuplicateGroups(roots, minSize).then((g) => g ?? []), // null değil dizi

  deleteFiles: (paths: string[], opts: DeleteOptions) => App.DeleteFiles(paths, fileops.DeleteOptions.createFrom(opts)),
  moveFiles: (paths: string[], destDir: string, conflict: Conflict) =>
    App.MoveFiles(paths, destDir, conflict),
  renameFiles: (items: RenameItem[]) =>
    App.RenameFiles(items.map((i) => fileops.RenameItem.createFrom(i))),
  validateFileName: (name: string) => App.ValidateFileName(name),
}

// internal/imaging'deki desteklenen uzantılarla aynı tutulmalı.
const IMAGE_EXTS = new Set(['.jpg', '.jpeg', '.jfif', '.png', '.gif', '.bmp', '.tif', '.tiff', '.webp'])

export function isImage(path: string): boolean {
  const i = path.lastIndexOf('.')
  return i >= 0 && IMAGE_EXTS.has(path.slice(i).toLowerCase())
}

// Go'nun asset sunucusundaki /thumb adresinden küçük resim (en uzun kenar `size`).
export function thumbUrl(path: string, size: number): string {
  return `/thumb?s=${size}&path=${encodeURIComponent(path)}`
}

// Olay abonelikleri; dönen fonksiyon aboneliği kaldırır (useEffect temizliği).
export const events = {
  onScanProgress: (cb: (p: ScanProgress) => void) => EventsOn('scan:progress', cb),
  onScanDone: (cb: (r: ScanResult) => void) => EventsOn('scan:done', cb),
  onScanError: (cb: (message: string) => void) => EventsOn('scan:error', cb),
  onOpsProgress: (cb: (p: OpsProgress) => void) => EventsOn('ops:progress', cb),
}
