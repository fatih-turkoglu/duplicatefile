//go:build darwin

// Package macos wraps Foundation APIs without launching Finder or AppleScript.
package macos

/*
#cgo LDFLAGS: -framework Foundation -framework UniformTypeIdentifiers
#include <stdlib.h>
char *df_languages(void);
int df_dark_mode(void);
int df_trash(const char *path, char **message);
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"unsafe"
)

var ErrTrashUnavailable = errors.New("this volume does not support Trash")

func PreferredLanguages() []string {
	value := C.df_languages()
	if value == nil {
		return []string{}
	}
	defer C.free(unsafe.Pointer(value))
	languages := []string{}
	_ = json.Unmarshal([]byte(C.GoString(value)), &languages)
	return languages
}

func DarkMode() bool { return C.df_dark_mode() != 0 }

// Trash never falls back to unlinking a file when the native operation fails.
func Trash(path string) error {
	if path == "" || containsNUL(path) {
		return fs.ErrInvalid
	}
	value := C.CString(path)
	defer C.free(unsafe.Pointer(value))
	var message *C.char
	status := C.df_trash(value, &message)
	if message != nil {
		defer C.free(unsafe.Pointer(message))
	}
	if status == 0 {
		return nil
	}
	var cause error
	switch status {
	case 1:
		cause = ErrTrashUnavailable
	case 2:
		cause = fs.ErrPermission
	case 3:
		cause = fs.ErrNotExist
	default:
		cause = errors.New("could not move file to Trash")
	}
	return fmt.Errorf("%w: %s", cause, C.GoString(message))
}

func containsNUL(s string) bool {
	for i := range s {
		if s[i] == 0 {
			return true
		}
	}
	return false
}
