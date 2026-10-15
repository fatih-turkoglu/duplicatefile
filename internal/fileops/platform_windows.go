//go:build windows

package fileops

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	shell32              = syscall.NewLazyDLL("shell32.dll")
	procSHFileOperationW = shell32.NewProc("SHFileOperationW")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetDriveTypeW    = kernel32.NewProc("GetDriveTypeW")
)

const (
	foDelete           = 0x0003
	fofSilent          = 0x0004
	fofNoConfirmation  = 0x0010
	fofAllowUndo       = 0x0040
	fofNoErrorUI       = 0x0400
	fofWantNukeWarning = 0x4000 // geri dönüşüm yerine kalıcı silinecekse sor

	driveFixed = 3

	errorNotSameDevice = syscall.Errno(17)
)

// shFileOpStruct, SHFILEOPSTRUCTW'nin 64 bit karşılığıdır (shellapi.h 64 bitte
// doğal hizalama kullanır).
type shFileOpStruct struct {
	hwnd                  uintptr
	wFunc                 uint32
	pFrom                 *uint16
	pTo                   *uint16
	fFlags                uint16
	fAnyOperationsAborted int32
	hNameMappings         uintptr
	lpszProgressTitle     *uint16
}

// moveToRecycleBin dosyayı Geri Dönüşüm Kutusu'na gönderir.
func moveToRecycleBin(path string) error {
	// Geri Dönüşüm Kutusu olmayan sürücülerde (ağ, USB bellek) SHFileOperation
	// dosyayı sessizce kalıcı siler; bunu engellemek için önceden reddedilir.
	if !hasRecycleBin(path) {
		return ErrNoRecycleBin
	}

	from, err := syscall.UTF16FromString(path)
	if err != nil {
		return err
	}
	from = append(from, 0) // pFrom çift NUL ile biten bir listedir

	op := shFileOpStruct{
		wFunc:  foDelete,
		pFrom:  &from[0],
		fFlags: fofAllowUndo | fofNoConfirmation | fofSilent | fofNoErrorUI | fofWantNukeWarning,
	}
	r, _, _ := procSHFileOperationW.Call(uintptr(unsafe.Pointer(&op)))
	runtime.KeepAlive(from)

	if r != 0 {
		return fmt.Errorf("could not move to the Recycle Bin (code 0x%X)", r)
	}
	if op.fAnyOperationsAborted != 0 {
		return errors.New("moving to the Recycle Bin was cancelled")
	}
	// SHFileOperation bazı hatalarda yine 0 döner; dosya duruyorsa başarısızdır.
	if _, err := os.Lstat(path); err == nil {
		return errors.New("file could not be moved to the Recycle Bin")
	}
	return nil
}

// hasRecycleBin, yolun bulunduğu sürücünün sabit disk olup olmadığını söyler.
// UNC yolları (\\sunucu\paylaşım) uzak sürücü sayılır.
func hasRecycleBin(path string) bool {
	root, err := syscall.UTF16PtrFromString(filepath.VolumeName(path) + `\`)
	if err != nil {
		return false
	}
	t, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(root)))
	return t == driveFixed
}

func isCrossDevice(err error) bool {
	return errors.Is(err, errorNotSameDevice)
}
