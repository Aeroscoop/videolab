package main

import (
    "context"
    "embed"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
    app := NewApp()
    err := wails.Run(&options.App{
        Title: "VideoLab",
        Width: 1280,
        Height: 850,
        MinWidth: 940,
        MinHeight: 620,
        AssetServer: &assetserver.Options{Assets: assets},
        BackgroundColour: &options.RGBA{R: 17, G: 25, B: 23, A: 1},
        OnStartup: func(ctx context.Context) { app.startup(ctx) },
        Bind: []interface{}{app},
    })
    if err != nil { println("Failed to start VideoLab:", err.Error()) }
}
