//go:build darwin

package pathutil

import (
	"path/filepath"
	"strings"
)

// IsProtected identifies macOS system data and application/library packages.
// Resolve aliases too, so selecting a symlinked directory cannot bypass it.
func IsProtected(path string) bool {
	if protectedMacPath(path) {
		return true
	}
	resolved, err := filepath.EvalSymlinks(path)
	return err == nil && resolved != path && protectedMacPath(resolved)
}

func protectedMacPath(path string) bool {
	path = strings.ToLower(filepath.Clean(path))
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for _, part := range parts {
		switch filepath.Ext(part) {
		case ".app", ".framework", ".bundle", ".plugin", ".appex", ".xpc", ".kext", ".prefpane", ".photoslibrary", ".photolibrary":
			return true
		}
	}
	// Recognise system folders on mounted disks as well as the startup disk.
	if len(parts) > 2 && parts[0] == "volumes" {
		parts = parts[2:]
		path = "/" + strings.Join(parts, "/")
	}
	for _, root := range []string{"/system", "/library", "/applications", "/usr", "/bin", "/sbin", "/dev", "/etc", "/cores", "/network", "/private/etc", "/private/var/db", "/private/var/root", "/private/var/vm", "/private/var/protected", "/var/db", "/var/root", "/var/vm"} {
		if Within(path, root) {
			return true
		}
	}
	// Skip these trees when walking /, while still allowing explicitly chosen
	// disposable temporary folders beneath them (including Go/E2E fixtures).
	if path == "/private" || path == "/var" || path == "/tmp" {
		return true
	}
	return len(parts) >= 3 && parts[0] == "users" && (parts[2] == "library" || parts[2] == "applications")
}
