package main

import (
	"embed"
	"log"

	"certificate-generator/app"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	appInstance, err := app.NewApp(app.Config{
		FontDir: "assets/fonts",
	})
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
	defer appInstance.Close()

	wailsApp := application.New(application.Options{
		Name:        "Certificate Generator",
		Description: "Professional Desktop Certificate Generator",
		Services: []application.Service{
			application.NewService(appInstance),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Certificate Generator Janaiz",
		Width:     1280,
		Height:    850,
		URL:       "/",
		MinWidth:  1024,
		MinHeight: 700,
	})

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}
