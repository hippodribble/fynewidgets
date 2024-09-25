package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
	"github.com/pkg/errors"
)

const MINCOLUMNS int = 1
const MAXCOLUMNS int = 10

var left, right, top, bottom, centre *fyne.Container
var statuschannel chan interface{} = make(chan interface{})
var columnchannel chan int = make(chan int)

var thumbnailgrid *fynewidgets.ThumbnailGrid
var allornone bool
var slide *widget.Slider
var lastfolder fyne.ListableURI
var details *fynewidgets.Loupe

func main() {

	appp := app.NewWithID("com.github.hippodribble.fynewidgets.demoapp")
	w := appp.NewWindow("com.github.hippodribble.fynewidgets Demonstration " + fynewidgets.VERSION)
	w.SetContent(gui())
	w.Resize(fyne.NewSize(1200, 900))

	w.ShowAndRun()
}

func gui() *fyne.Container {
	left = container.NewStack()
	details = fynewidgets.NewLoupe(image.Pt(300, 300), 1)

	bleft := container.NewBorder(nil, details, nil, nil, left)
	right = container.NewStack()

	b := makeleftbuttons()
	top = container.NewBorder(nil, nil, container.NewHBox(b...), makescreenshotbutton(), nil)
	centre = container.NewStack()
	bottom = container.NewStack(fynewidgets.NewStatusProgress(statuschannel))

	return container.NewBorder(top, bottom, bleft, right, centre)
}

func makescreenshotbutton() fyne.CanvasObject {
	return widget.NewButton("Screenshot", fynewidgets.Screenshot)
}

func makeleftbuttons() []fyne.CanvasObject {
	buttons := make([]fyne.CanvasObject, 0)
	bFolder := widget.NewButtonWithIcon("", theme.FolderIcon(), thumbnailsfromfolder)
	buttons = append(buttons, bFolder)
	bFile := widget.NewButtonWithIcon("", theme.FileIcon(), imagefromfile)
	buttons = append(buttons, bFile)
	bMulti := widget.NewButtonWithIcon("", theme.GridIcon(), gridfromselection)
	buttons = append(buttons, bMulti)
	bAllNone := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), toggleallornone)
	buttons = append(buttons, bAllNone)
	slide = widget.NewSlider(float64(MINCOLUMNS), float64(MAXCOLUMNS))
	slide.SetValue(3)
	slide.Step = 1
	slide.OnChanged = func(f float64) {
		statuschannel <- fmt.Sprintf("%d columns", int(f))
	}
	slide.OnChangeEnded = func(f float64) {
		n := int(f)
		if n >= MINCOLUMNS && n <= MAXCOLUMNS && columnchannel != nil {
			columnchannel <- n
		} else {
			log.Println("columnchannel is nil or out of range", n)
		}
	}
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(100, 1))
	buttons = append(buttons, container.NewStack(r, slide))

	return buttons
}

// put some selectable thumbnails on the left
func thumbnailsfromfolder() {
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

		thumbnailgrid, err = fynewidgets.NewThumbnailGrid(list, 100, 100, 20, statuschannel, thumbnailchannel)
		if err != nil {
			statuschannel <- "Problem creating grid " + err.Error()
		}
		thumbnailgrid.SetThumbnailChannel(thumbnailchannel)
		go func() {
			for {
				uri := <-thumbnailchannel
				log.Println(uri.Path())
				showImage(uri)
			}
		}()
		thumbscroll := container.NewVScroll(thumbnailgrid)
		centre.RemoveAll()
		left.RemoveAll()
		left.Add(thumbscroll)
		left.Refresh()
		slide.SetValue(3)
		lastfolder = lu
		allornone = false

	}, fyne.CurrentApp().Driver().AllWindows()[0])
	if lastfolder != nil {
		dlg.SetLocation(lastfolder)
	}
	dlg.Show()
}

func imagefromfile() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			statuschannel <- "Cancelled: " + err.Error()
			return
		}
		if uc==nil{return}
		if fynewidgets.IsImage(uc.URI()) {
			err:=showImage(uc.URI())
			if err!=nil{
				statuschannel<-errors.Wrap(err,"displaying image").Error()
			}
		}else{
			statuschannel<-fynewidgets.Message{Text:"That's not an image!",Duration: 2}
		}

	}, fyne.CurrentApp().Driver().AllWindows()[0])
	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.Show()
}

func showImage(uri fyne.URI)error {

	if fynewidgets.IsImage(uri) {
		decodedImg, err := fynewidgets.NewPanZoomCanvasFromFile(uri, image.Pt(100, 100), statuschannel)
		if err != nil {
			return errors.Wrap(err,"creating pan zoom canvas")
		}
		if decodedImg == nil {
			return errors.Wrap(err,"empty canvas when creating pan zoom canvas")
		}
		decodedImg.SetLoupe(details)
		centre.RemoveAll()
		centre.Add(decodedImg)
		centre.Refresh()
	}
	return nil
}

func gridfromselection() {

	if thumbnailgrid == nil {
		statuschannel <- "No thumbnails available"
		return
	}

	if len(thumbnailgrid.Thumbnails) == 0 {
		statuschannel <- "No thumbnails available"
		return
	}

	selectedURIs := make([]fyne.URI, 0)
	for i := range thumbnailgrid.Thumbnails {
		if thumbnailgrid.Thumbnails[i].IsSelected() {
			selectedURIs = append(selectedURIs, thumbnailgrid.Thumbnails[i].URI)
		}
	}

	if len(selectedURIs) == 0 {
		statuschannel <- "No thumbnail images selected!"
		return
	}

	g, err := fynewidgets.NewSynchronisedImageGrid(3, statuschannel)
	if err != nil {
		log.Fatalln(err)
	}
	g.SetColumnChannel(columnchannel)
	g.SetImages(selectedURIs)

	// add a loupe to each image for fun!
	items, err := g.Items()
	if err == nil {
		for i := range items {
			if o, ok := items[i].(*fynewidgets.PanZoomCanvas); ok {
				o.SetLoupe(details)
			}
		}
	}

	centre.RemoveAll()
	centre.Add(g)
	centre.Refresh()

}

func toggleallornone() {
	if thumbnailgrid == nil {
		return
	}
	if len(thumbnailgrid.Thumbnails) == 0 {
		return
	}
	allornone = !allornone
	for i := range thumbnailgrid.Thumbnails {
		thumbnailgrid.Thumbnails[i].SetSelected(allornone)
	}
}
