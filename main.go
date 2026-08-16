package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed all:backend/migrations/sqlite
var sqliteMigrations embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  app.cfg.App.WindowTitle,
		Width:  app.cfg.App.Width,
		Height: app.cfg.App.Height,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
			app.bindings,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
