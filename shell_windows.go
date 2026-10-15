//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"DuplicateFile/internal/apperr"
)

// showInExplorer, Gezgin'i açıp dosyayı seçili gösterir.
func showInExplorer(path string) error {
	if !filepath.IsAbs(path) {
		return apperr.Errorf("path_not_absolute", "path must be absolute: %s", path)
	}
	if _, err := os.Lstat(path); err != nil {
		return err
	}
	cmd := exec.Command("explorer.exe")
	// explorer kendi komut satırı ayrıştırıcısını kullanır; Go'nun argüman
	// tırnaklaması "/select," kısmını da tırnağa aldığı için satır elle kurulur.
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf(`explorer.exe /select,"%s"`, path)}
	// explorer başarılı olsa da sıfırdan farklı çıkış kodu döner; beklenmez.
	return cmd.Start()
}
