package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	iw "github.com/hippodribble/fynewidgets/adaptiveimagewidget"
	th "github.com/hippodribble/fynewidgets/listablethumbnail"
)

var status *widget.Label
var infiniteProgress *widget.ProgressBarInfinite
var progress *widget.ProgressBar
var centre *fyne.Container
var description *widget.Label
var lastfolder fyne.ListableURI
var pictureURIs []fyne.URI
var listbox,zoombox, top, leftside,b *fyne.Container
var tophidden,lefthidden bool

// var dlg *dialog.FileDialog

func main() {
	ap := app.NewWithID("com.github.hippodribble.fynewidgets")
	w := ap.NewWindow("Check ImageWidget")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(1200, 900))
	w.ShowAndRun()
}

func gui() *fyne.Container {
	bCanvas := widget.NewButton("Canvas.Image", openImageNormally)
	bAdaptive := widget.NewButton("AdaptiveImageWidget", openImage)
	description = widget.NewLabel("")
	description.TextStyle.Italic = true
	bFolder := widget.NewButton("Folder", openFolder)
	top = container.NewBorder(nil, nil, container.NewHBox(bCanvas, bAdaptive), bFolder, description)

	centre = container.NewStack()

	status = widget.NewLabel("Load a large image")
	status.TextStyle.Bold = true
	status.Truncation = fyne.TextTruncateClip
	infiniteProgress = widget.NewProgressBarInfinite()
	infiniteProgress.Hide()
	progress = widget.NewProgressBar()
	progress.Hide()
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(200, 10))
	bottom := container.NewBorder(nil, nil, nil, container.NewStack(r, infiniteProgress, progress), status)

	listbox = container.NewStack()
	zoombox=container.NewStack()
	leftside=container.NewBorder(nil, zoombox, nil, nil, listbox)
	b = container.NewBorder(top, bottom, leftside, nil, centre)
	return b
}

func openImage() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {

		centre.RemoveAll()
		description.SetText("")
		infiniteProgress.Show()
		defer infiniteProgress.Hide()
		if err != nil || uc == nil {
			status.SetText("File dialog error")
			return
		}

		uri := uc.URI()
		status.SetText("Displaying " + uri.Path())
		infiniteProgress.Start()
		
		im,err:=imaging.Open(uri.Path())
		if err!=nil{
			return
		}

		widget, err := iw.NewImageWidget(&im, 200)
		if err != nil {
			status.SetText("Error creating widget: " + err.Error())
			return
		}

		centre.Add(widget)
		centre.Refresh()
		infiniteProgress.Stop()
		status.SetText("")
		dw:=widget.DetailedImage(400,400)
		zoombox.RemoveAll()
		zoombox.Add(dw)
		zoombox.Refresh()

		fldr, err := storage.Parent(uri)
		if err != nil {
			lastfolder = nil
		}

		lastfolder, err = storage.ListerForURI(fldr)
		if err != nil {
			log.Println("No lister for folder", err.Error())
			lastfolder=nil
		}

	}, fyne.CurrentApp().Driver().AllWindows()[0])

	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.SetConfirmText("Yeah")
	dlg.SetDismissText("Nah")
	dlg.SetLocation(lastfolder)
	dlg.Show()
}

func openImageNormally() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		centre.RemoveAll()
		description.SetText("")
		infiniteProgress.Show()
		defer infiniteProgress.Hide()
		if err != nil {
			status.SetText("File dialog error")
			return
		}
		uri := uc.URI()
		status.SetText("Loading file " + uri.Path())
		infiniteProgress.Start()
		f, err := os.Open(uri.Path())
		if err != nil {
			status.SetText("Error opening file")
			return
		}
		cfg, _, err := image.DecodeConfig(f)
		if err != nil {
			status.SetText("File is not an image")
			return
		}
		pixcount := cfg.Height * cfg.Width
		var pix string
		if pixcount < 1000000 {
			pix = "< 1 megapixel"
		} else {
			pix = fmt.Sprintf("%.1f megapixels", float64(pixcount)/1000000)
		}

		description.SetText(fmt.Sprintf("%s - %dx%d (%s) ", uri.Name(), cfg.Width, cfg.Height, pix))
		f.Seek(0, 0)
		im, format, err := image.Decode(f)
		if err != nil {
			status.SetText("File is not an image")
			return
		}
		infiniteProgress.Stop()
		status.SetText("Displaying " + format)
		infiniteProgress.Start()
		widget := canvas.NewImageFromImage(im)
		widget.FillMode = canvas.ImageFillContain
		widget.ScaleMode = canvas.ImageScaleFastest
		centre.Add(widget)
		centre.Refresh()
		infiniteProgress.Stop()
		status.SetText("Done")
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.Show()
}

func openFolder() {
	dlg := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
		if err != nil {
			status.SetText("Folder picker error shocker")
			return
		}
		if lu == nil {
			status.SetText("Cancelled")
			return
		}
		lastfolder = lu
		scanFolder(lu)
	}, fyne.CurrentApp().Driver().AllWindows()[0])
	dlg.SetLocation(lastfolder)
	// dlg.SetConfirmText("Y")
	// dlg.SetDismissText("N")
	dlg.Resize(fyne.NewSize(800,800))

	dlg.Show()
}

func scanFolder(folder fyne.ListableURI) {
	uris, err := folder.List()
	if err != nil {
		status.SetText("Bad folder. Bad Bad Bad.")
		return
	}
	// shorten the list
	pictureURIs = make([]fyne.URI, 0)
	for _, uri := range uris {
		switch strings.ToLower(uri.Extension()) {
		case ".jpg", ".gif", ".png":
			pictureURIs = append(pictureURIs, uri)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(pictureURIs))
	thumbs := make([]*th.Thumbnail, len(pictureURIs))
	n := float64(len(pictureURIs))

	dProg := 1 / n
	progress.SetValue(0)
	progress.Show()

	for i, uri := range pictureURIs {
		go func(w *sync.WaitGroup, u fyne.URI, index int) {
			defer w.Done()
			// status.SetText(uri.Path())
			t, err := th.NewThumbNail(uri, 300)
			if err != nil {
				return
			}
			thumbs[index] = t
			progress.SetValue(progress.Value + dProg)
		}(&wg, uri, i)
	}
	wg.Wait()
	progress.SetValue(0)
	progress.Hide()
	sort.Slice(thumbs, func(i, j int) bool {
		return thumbs[i].Caption < thumbs[j].Caption
	})
	showThumbnails(thumbs)
}

func showThumbnails(thumbs []*th.Thumbnail) {

	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(400, 250))

	l := widget.NewList(
		func() int {
			return len(thumbs)
		},
		func() fyne.CanvasObject {
			return container.NewStack(r)
		},
		func(i widget.ListItemID, co fyne.CanvasObject) {
			if co, ok := co.(*fyne.Container); ok {
				// co.RemoveAll()
				co.RemoveAll()
				co.Add(r)
				co.Add(thumbs[i])
			}
		},
	)
	l.OnSelected = pickApicture
	
	listbox.RemoveAll()
	listbox.Add(container.NewBorder(nil,widget.NewSeparator(),nil,nil,l))
	// listbox.Add(container.NewBorder(r, nil, nil, nil, container.NewScroll(g)))
	listbox.Refresh()

}

func pickApicture(id int) {
	centre.RemoveAll()

	uri := pictureURIs[id]
	// uri := uc.URI()
	status.SetText("Displaying " + uri.Path())
	infiniteProgress.Start()
	
	im,err:=imaging.Open(uri.Path(),imaging.AutoOrientation(true))
	if err!=nil{
		return
	}

	widget, err := iw.NewImageWidget(&im, 200)
	if err != nil {
		status.SetText("Error creating widget: " + err.Error())
		return
	}

	widget.Requestfullscreen.AddListener(binding.NewDataListener(func() {
		tophidden,_=widget.Requestfullscreen.Get()
		lefthidden,_=widget.Requestfullscreen.Get()
		if(tophidden){
			top.Hide()
		}else{top.Show()}
		if(lefthidden){
			leftside.Hide()
		}else{leftside.Show()}
		b.Refresh()
		log.Println(tophidden,"selected")
	}))
	
	centre.Add(widget)
	centre.Refresh()
	infiniteProgress.Stop()
	status.SetText("")
	b:=widget.Pyramid[0].Bounds()
	info:=fmt.Sprintf("%s - %v %.3f Mpixel",uri.Name(),b,float64(b.Dx()*b.Dy())/1000000)
	fyne.CurrentApp().Driver().AllWindows()[0].SetTitle(info)
	dw:=widget.DetailedImage(400,400)
	
	log.Println(dw.MinSize())
	zoombox.RemoveAll()
	zoombox.Add(dw)
	// log.Println("Added detail image")
	zoombox.Refresh()
	// log.Println("refreshed detail image")

	fldr, err := storage.Parent(uri)
	if err != nil {
		lastfolder = nil
	}

	lastfolder, err = storage.ListerForURI(fldr)
	if err != nil {
		log.Println("No lister for folder", err.Error())
		lastfolder=nil
	}

	centre.Add(widget)
	centre.Refresh()
	infiniteProgress.Stop()
	status.SetText("Done")
}
