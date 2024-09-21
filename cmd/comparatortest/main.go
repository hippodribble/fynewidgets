package main

import (
	"fmt"
	"image/color"
	"log"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"github.com/hippodribble/fynewidgets/comparatorwidget"
	"github.com/hippodribble/fynewidgets/listablethumbnail"
	"github.com/hippodribble/fynewidgets/panzoomwidget"
)

const VERSION string = "0.0.1"

// communicate data make share etc
var statuschannel chan string = make(chan string)
var progresschannel chan float64 = make(chan float64)
var thumbnailCommChan chan interface{} = make(chan interface{})

var lastfolder fyne.ListableURI
var pictureURIs []fyne.URI
var mainlayout, top, bottom, left, right, centre *fyne.Container
var thumbwindow *listablethumbnail.ThumbnailGrid

// var thumbs []*listablethumbnail.Thumbnail
var allnone binding.Bool = binding.NewBool()

func main() {
	ap := app.New()
	w := ap.NewWindow("Image Comparator v" + VERSION)
	w.Resize(fyne.NewSize(1200, 800))
	w.SetContent(gui())
	w.ShowAndRun()
}

func gui() *fyne.Container {
	top = container.NewHBox()
	top.Add(widget.NewButtonWithIcon("Set Image Folder", theme.FolderOpenIcon(), openFolder))
	bInfo := widget.NewButtonWithIcon("Information", theme.InfoIcon(), showSystemInformation)
	top.Add(bInfo)
	bCompare := widget.NewButton("Compare", loadComparator)
	top.Add(bCompare)
	bSelectAllNone := widget.NewButton("All/None", selectAllNone)
	top.Add(bSelectAllNone)

	welcomelabel := canvas.NewText("Welcome...", color.NRGBA{192, 192, 0, 255})
	welcomelabel.SetMinSize(fyne.NewSize(500, 500))
	welcomelabel.Alignment = fyne.TextAlignCenter
	welcomelabel.TextSize = 72
	welcomelabel.TextStyle.Bold = true
	centre = container.NewStack(welcomelabel)

	left = container.NewStack()
	right = container.NewStack()
	progress := widget.NewProgressBar()
	progress.Hide()
	status := widget.NewLabel("Ready")
	status.TextStyle.Bold = true
	statuschannel = make(chan string)
	progresschannel = make(chan float64)
	thumbnailCommChan = make(chan interface{})
	go func() {
		for {
			select {
			case p := <-progresschannel:
				if p == 0 {
					progress.Hide()
				} else {
					progress.Show()
					if p > 0 {
						progress.SetValue(p)
					} else {
						progress.SetValue(progress.Value - p)
					}
				}
			case t := <-statuschannel:
				status.SetText(t)
			case o := <-thumbnailCommChan:
				if uri, ok := o.(fyne.URI); ok {
					log.Println("Selected a humbnail URI")
					go loadMainImage(uri)
				}
			}
		}
	}()

	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(100, 10))
	bottom = container.NewBorder(nil, nil, nil, container.NewStack(r, progress), status)

	mainlayout = container.NewBorder(top, bottom, left, right, centre)

	return mainlayout
}

func openFolder() {
	dlg := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
		if err != nil {
			statuschannel <- "Folder picker error shocker"
			return
		}
		if lu == nil {
			statuschannel <- "Cancelled"
			return
		}
		lastfolder = lu
		scanFolder(lu)
	}, fyne.CurrentApp().Driver().AllWindows()[0])
	dlg.SetLocation(lastfolder)
	dlg.Resize(fyne.NewSize(800, 800))

	dlg.Show()
}

func scanFolder(folder fyne.ListableURI) {
	left.RemoveAll()
	centre.RemoveAll()
	statuschannel <- "Scanning folder"
	uris, err := folder.List()
	if err != nil {
		statuschannel <- "Bad folder. Bad Bad Bad."
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

	thumbwindow, err = listablethumbnail.NewThumbnailGrid(pictureURIs, 100, 100, 18, thumbnailCommChan)
	if err != nil {
		statuschannel <- "Error making thumbnails: " + err.Error()
	}

	left.RemoveAll()
	left.Add(container.NewVScroll(thumbwindow))
	left.Refresh()
	statuschannel <- fmt.Sprintf("Loaded information for %d images", len(thumbwindow.Thumbnails))
}

func loadMainImage(uri fyne.URI) {

	statuschannel <- "Loading file " + uri.Path()
	centre.RemoveAll()
	im, err := imaging.Open(uri.Path(), imaging.AutoOrientation(true))
	if err != nil {
		statuschannel <- err.Error()
		return
	}

	// GENERAL APPROACH:

	// 1. Create widget
	// 2. Attach to any channels
	// 3. Set up a request for Full Screen
	widget, err := panzoomwidget.NewPanZoomWidget(&im, 200, 7)
	if err != nil {
		statuschannel <- "Error creating widget: " + err.Error()
		return
	}
	ch := widget.GetMessageChannel()
	go func() {
		for {
			select {
			case t := <-ch:
				statuschannel <- t
			}
		}
	}()

	widget.Requestfullscreen.AddListener(binding.NewDataListener(func() {
		hider, _ := widget.Requestfullscreen.Get()
		if hider {
			top.Hide()
			left.Hide()
			bottom.Hide()
		} else {
			top.Show()
			left.Show()
			bottom.Show()
		}
		mainlayout.Refresh()
	}))
	centre.Add(widget)
	centre.Refresh()
}

func showSystemInformation() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	statuschannel <- fmt.Sprintf("System Memory System: %d MB , Allocated: %d MB, GC fraction %.2f\n", stats.Sys/1000000, stats.Alloc/1000000, stats.GCCPUFraction)
}

func loadComparator() {
	// log.Println("Load images")
	if thumbwindow == nil {
		return
	}

	statuschannel <- "Loading images..."
	// urisToShow := make([]fyne.URI, 0)
	// for _, thumb := range thumbs {
	// 	if thumb.IsSelected() {
	// 		// log.Println(thumb.URI, "is selected")
	// 		urisToShow = append(urisToShow, thumb.URI)
	// 	}
	// }
	urisToShow := thumbwindow.SelectedFiles()

	c, err := comparatorwidget.NewComparatorWidget(urisToShow, 3, thumbnailCommChan)
	if err != nil {
		statuschannel <- err.Error()
	}
	mchan := c.MouseChannel()
	go func() {
		for {
			select {
			case p := <-*mchan:
				statuschannel <- fmt.Sprintf("Mouse %.1f,%.1f", p.X, p.Y)
			}
		}
	}()

	centre.RemoveAll()
	centre.Add(c)
	centre.Refresh()
	statuschannel <- fmt.Sprintf("Loaded %d images", len(urisToShow))
}

func selectAllNone() {
	if thumbwindow == nil {
		return
	}
	thumbwindow.ToggleSelection()
	// if len(thumbwindow.SelectedFiles()) == 0 {
	// 	return
	// }
	// b, _ := allnone.Get()
	// for _, t := range thumbs {
	// 	t.SetSelected(!b)
	// }
	// allnone.Set(!b)
}
