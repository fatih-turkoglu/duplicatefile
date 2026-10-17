package appdirs

import (
	"os"
	"path/filepath"
	"runtime"
)

// Data keeps persistent macOS settings out of the disposable Library/Caches.
// Windows retains its existing LocalAppData location.
func Data() (string, error) {
	base, err := os.UserCacheDir()
	if runtime.GOOS == "darwin" {
		base, err = os.UserConfigDir()
	}
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "DuplicateFile"), nil
}
