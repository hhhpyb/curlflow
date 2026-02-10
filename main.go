package main

import (
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"curlflow/cmd"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cmd.StartApp = func() {
		// Create an instance of the app structure
		app := NewApp()

		// Create application with options
		err := wails.Run(&options.App{
			Title:         "curlflow",
			Width:         1024,
			Height:        768,
			MinWidth:      800,
			MinHeight:     600,
			DisableResize: false,
			AssetServer: &assetserver.Options{
				Assets: assets,
			},
			BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
			OnStartup:        app.startup,
			OnBeforeClose:    app.beforeClose,
			Mac: &mac.Options{
				TitleBar:             mac.TitleBarHiddenInset(),
				Appearance:           mac.NSAppearanceNameDarkAqua,
				WebviewIsTransparent: true,
				WindowIsTranslucent:  true,
			},
			Bind: []interface{}{
				app,
			},
		})

		if err != nil {
			// Log error to file since GUI apps don't have stdout/stderr on Windows
			f, _ := os.OpenFile("startup_error.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				defer f.Close()
				f.WriteString(fmt.Sprintf("%s - Error: %s\n", time.Now().Format(time.RFC3339), err.Error()))
			}
			println("Error:", err.Error())
		}
	}

	cmd.Execute()
}
