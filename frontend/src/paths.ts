// POSIX names may contain backslashes; only Windows paths treat them as separators.
export function isWindowsPath(path: string): boolean {
  return /^[a-z]:[\\/]/i.test(path) || path.startsWith('\\\\')
}

export function pathKey(path: string): string {
  return isWindowsPath(path) ? path.replaceAll('/', '\\').toLowerCase() : path
}

export function lastSeparator(path: string): number {
  return isWindowsPath(path) ? Math.max(path.lastIndexOf('\\'), path.lastIndexOf('/')) : path.lastIndexOf('/')
}
