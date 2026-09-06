package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/hippodribble/fynewidgets/gps"
)

func main() {
	ap := app.NewWithID("com.github.hippodribble.fynewidgets.gps")
	w := ap.NewWindow("Zoid")
	w.SetContent(gui())
	// w.Resize(fyne.NewSize(1000, 1000))
	// w.SetFullScreen(true)
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	server := gps.NewGPSMonitor()
	return server
}
