package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Arayüz yüklenene kadar görünen arka plan, temanın background.default
	// rengiyle aynı; koyu temada açılışta beyaz parlama olmasın.
	bg := &options.RGBA{R: 245, G: 246, B: 250, A: 255}
	if systemPrefersDark() {
		bg = &options.RGBA{R: 18, G: 19, B: 24, A: 255}
	}

	// Create application with options
	opts := &options.App{
		Title:     "DuplicateFile",
		Width:     1100,
		Height:    760,
		MinWidth:  400,
		MinHeight: 480,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: thumbMiddleware(newThumbCache(64 << 20)), // küçük resimler, en fazla 64 MB
		},
		BackgroundColour: bg,
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			Theme: windows.SystemDefault,
		},
	}
	if runtime.GOOS == "darwin" {
		// Native menus also enable the standard Cmd+C/V/X/Z/A shortcuts.
		opts.Menu = menu.NewMenu()
		opts.Menu.Append(menu.AppMenu())
		opts.Menu.Append(menu.EditMenu())
		opts.Mac = &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			About:    &mac.AboutInfo{Title: "DuplicateFile " + version, Message: "Duplicate and similar file finder"},
		}
	}

	if err := wails.Run(opts); err != nil {
		println("Error:", err.Error())
	}
}
