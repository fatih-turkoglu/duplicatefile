package main

import "runtime"

// version, derlemede build.sh/build.ps1 tarafından wails.json'daki info.productVersion
// ile doldurulur (-ldflags "-X main.version=..."). Doğrudan `wails build` ya
// da `wails dev` ile "dev" kalır.
var version = "dev"

// AppInfo, Ayarlar'daki "Hakkında" bölümünün bilgisidir.
type AppInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"goVersion"`
}

// AppInfo uygulama sürümünü döner.
func (a *App) AppInfo() AppInfo {
	return AppInfo{Version: version, GoVersion: runtime.Version()}
}
