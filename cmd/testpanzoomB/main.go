package main

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	smallwidget "github.com/hippodribble/fynewidgets/smallimagewidget"
	"github.com/hippodribble/fynewidgets/statusprogress"
)

var top, bottom, centre *fyne.Container
var ch chan interface{}
var w fyne.Window

func main() {
	ap := app.New()
	w = ap.NewWindow("PZ2 Test 0.1.0")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(1200, 900))
	w.ShowAndRun()
}

func gui() *fyne.Container {
	bFileOpen := widget.NewButton("File", openfile)

	top = container.NewHBox()
	top.Add(bFileOpen)
	ch = make(chan interface{})
	sp := statusprogress.NewStatusProgress(ch)
	bottom = container.NewStack(sp)
	centre = container.NewStack()
	return container.NewBorder(top, bottom, nil, nil, centre)
}

func openfile() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			return
		}
		if uc == nil {
			return
		}
		uri := uc.URI()
		loadimage(uri)
	}, w)
	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.Show()
}

func loadimage(uri fyne.URI) {
	img, err := imaging.Open(uri.Path())
	if err != nil {
		ch <- "Can't open this file"
	}
	var imnrgba image.Image = imaging.Clone(img)

	centre.RemoveAll()
	// widget, err := panzoomwidgetb.NewPZWidget(imnrgba, uri.Name(), 7, ch)

	widget := smallwidget.NewSmallWidget(&imnrgba, "a description of the file goes here")
	if err != nil {
		ch <- err.Error()
	}

	centre.Add(widget)
	centre.Refresh()
}
