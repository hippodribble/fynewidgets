package main

import (
	"image"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
	"github.com/hippodribble/fynewidgets/pz4"
	"github.com/hippodribble/fynewidgets/statusprogress"
	"github.com/hippodribble/fynewidgets/thumbnail"
)

func gui() *fyne.Container {
	left = container.NewStack()
	right = container.NewStack()
	top = container.NewStack(container.NewHBox(makebuttons()...))
	centre = container.NewStack()
	bottom = container.NewStack(statusprogress.NewStatusProgress(statuschannel))
	return container.NewBorder(top, bottom, left, right, centre)
}

func makebuttons() []fyne.CanvasObject {
	buttons := make([]fyne.CanvasObject, 0)
	bFolder := widget.NewButtonWithIcon("", theme.FolderIcon(), openfolder)
	buttons = append(buttons, bFolder)
	bFile := widget.NewButtonWithIcon("", theme.FileIcon(), openfile)
	buttons = append(buttons, bFile)
	return buttons
}

// put some selectable thumbnails on the left
func openfolder() {
	log.Println("Open folder")
	dlg := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
		if err != nil {
			return
		}
		if lu == nil {
			return
		}

		list, err := lu.List()
		if err != nil {
			log.Println(err)
			return
		}
		thumbnailchannel := make(chan fyne.URI)

		grid, err := thumbnail.NewThumbnailGrid(list, 100, 100, 20, statuschannel, thumbnailchannel)
		if err != nil {
			statuschannel <- "Problem creating grid " + err.Error()
		}
		grid.SetThumbnailChannel(thumbnailchannel)
		go func() {
			for {
				uri := <-thumbnailchannel
				log.Println(uri.Path())
				loadimage(uri)
			}
		}()
		thumbscroll := container.NewVScroll(grid)
		left.RemoveAll()
		left.Add(thumbscroll)
		left.Refresh()

	}, fyne.CurrentApp().Driver().AllWindows()[0])

	dlg.Show()
}

func openfile() {
	log.Println("Open file")
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			return
		}
		loadimage(uc.URI())

	}, fyne.CurrentApp().Driver().AllWindows()[0])
	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.Show()
}

func loadimage(uri fyne.URI) {

	log.Println("Load image")
	if fynewidgets.IsImage(uri) {
		log.Println("Loading IMAGE")
		decodedImg, err := pz4.NewPZFromFile(uri, image.Pt(100, 100), statuschannel)
		if err != nil {
			log.Fatalln(err)
		}
		if decodedImg == nil {
			log.Println("Failed to decode image")
			return
		}

		centre.RemoveAll()
		centre.Add(decodedImg)
		centre.Refresh()
	}
}
