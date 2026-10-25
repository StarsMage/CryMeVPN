package main

import (
	"CryMeVPN/pkg/storage"
	"embed"
	"log"
	"path/filepath"

	//github
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	goruntime "runtime"
)

//go:embed all:frontend
var assets embed.FS

func main()  {
	path, err := storage.GetClearPath()

	if err != nil {
		log.Fatalf("%v",err)
	}

	app := NewApp()


	err = wails.Run(&options.App{
		Title: "CryMeVpn",
		Height: 740,
		Width: 360,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		DisableResize: true,
		HideWindowOnClose: goruntime.GOOS == "windows",
		OnStartup: app.Startup,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewUserDataPath: filepath.Join(path, "webview"),
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "crymeb346286c-354e-41e9-af74-5acc344b1d2b",
			OnSecondInstanceLaunch: app.OnSecondInstance,
		},
		},
	)

	if err != nil {
		log.Fatalf("main.go fatal error: %v",err)
	}
}