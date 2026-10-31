package main

import (
	"embed"
	"log"

	"certificate-generator/app"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
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
		Windows: application.WindowsOptions{
			UseVisualHosting:              true,
			DisableQuitOnLastWindowClosed: false,
		},
	})

	mainWindow := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Certificate Generator Janaiz",
		Width:     1280,
		Height:    850,
		URL:       "/",
		MinWidth:  1024,
		MinHeight: 700,
	})

	mainWindow.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if appInstance.HasActiveJobs() {
			event.Cancel()

			dialog := wailsApp.Dialog.Question()
			dialog.SetTitle("Konfirmasi Keluar")
			dialog.SetMessage("Proses pembuatan sertifikat sedang berlangsung.\nApakah Anda yakin ingin membatalkan proses dan keluar?")
			dialog.AddButton("Batal").SetAsCancel().SetAsDefault()
			dialog.AddButton("Keluar & Batalkan").OnClick(func() {
				appInstance.CancelAllJobs()
				_ = appInstance.Close()
				wailsApp.Quit()
			})
			dialog.Show()
			return
		}

		// Bersihkan koneksi database SQLite langsung agar tidak ada lock dan tutup seketika
		_ = appInstance.Close()
	})

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}
