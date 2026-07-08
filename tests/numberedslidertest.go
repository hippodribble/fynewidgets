package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/hippodribble/fynewidgets"
)

func main() {
	ap := app.New()
	w := ap.NewWindow("Test SliderLabel")
	sl := fynewidgets.NewSliderLabel(90,99.9,.1)
	w.SetContent(sl)
	w.Resize(fyne.NewSize(200, 100))
	w.ShowAndRun()
}
