package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	lazywidget "github.com/hippodribble/fynewidgets/lazyloadingimagewidget"
)

var top, centre *fyne.Container

func main() {
	
	log.Println("Welcome!")
	ap := app.New()
	w := ap.NewWindow("Lazy Load")
	w.Resize(fyne.NewSize(800, 600))
	w.SetContent(gui())
	w.ShowAndRun()

}

func gui() fyne.CanvasObject {

	top = container.NewHBox()
	centre = container.NewStack()
	bFile := widget.NewButton("File", openfile)
	top.Add(bFile)
	b := container.NewBorder(top, nil, nil, nil, centre)
	return b

}

func openfile() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {

		centre.RemoveAll()
		uri := uc.URI()
		log.Println("Displaying " + uri.Path())
		widget := lazywidget.NewLazyWidget(uri)
		centre.Add(widget)
		centre.Refresh()
		log.Println("Updated main UI")

	}, fyne.CurrentApp().Driver().AllWindows()[0])

	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.Show()
}
