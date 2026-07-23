package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	ap := app.New()
	w := ap.NewWindow("Picture Background")
	w.SetContent(gui())
	// w.Resize(fyne.NewSize(1000, 800))
	w.SetFullScreen(true)
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	cv := canvas.NewImageFromFile("/Users/glenn/Downloads/Photos/20251012_190614115.JPG")
	cv.FillMode = canvas.ImageFillCover
	cv.Translucency = 0.8
	me := widget.NewMultiLineEntry()
	me.SetMinRowsVisible(20)
	bottom:=widget.NewLabel("EDITOR")
	return container.NewBorder(nil,bottom,nil,nil,container.NewStack(me, cv))
}
