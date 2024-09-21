package main

import (
	"image"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets/pz4"
	"github.com/hippodribble/fynewidgets/statusprogress"
)

var w fyne.Window
var info chan interface{} = make(chan interface{})
var top, bottom, centre *fyne.Container

func main() {
	log.Println("Welcome ----------------------------------------------------------------------------")
	appp := app.New()
	w = appp.NewWindow("PZ3 v0.1")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	log.Println("GUI ----------------------------------------------------------------------------")
	filebutton := widget.NewButtonWithIcon("Get a File", theme.FileImageIcon(), Openfile)
	top = container.NewHBox(filebutton)
	bottom = container.NewStack(statusprogress.NewStatusProgress(info))
	centre = container.NewStack()

	return container.NewBorder(top, bottom, nil, nil, centre)

}

func Openfile() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			return
		}
		if uc == nil {
			return
		}
		uri := uc.URI()

		pz4, err := pz4.NewPZFromFile(uri, image.Pt(100, 100), info)
		if err != nil {
			info <- err.Error()
			return
		}

		centre.RemoveAll()
		// centre.Add(canvas.NewImageFromImage(fynewidgets.MakeFillerImage(100, 100)))
		// centre.Add(canvas.NewImageFromImage(image.NewUniform(color.Gray16{Y: 32})))
		centre.Add(pz4)

	}, w)
	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.Show()
}
