package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/hippodribble/fynewidgets/gps"
)

func main() {
	ap := app.New()
	w := ap.NewWindow("Zoid")
	w.SetContent(gui())
	// w.Resize(fyne.NewSize(1000, 1000))
	w.SetFullScreen(true)
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	server := gps.NewGPSMonitor()
	return server
}
