package main

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"sync"

	"fyne.io/cloud"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	eventbus "github.com/dtomasi/go-event-bus/v3"
	"github.com/hippodribble/fynewidgets"
)

var progress *fyne.Container = container.NewStack()
var status *fynewidgets.EventLabel
var myprefs fyne.Preferences
var myapp fyne.App
var notes *fynewidgets.PermanentNotes
var Bus *eventbus.EventBus
var left, centre *fyne.Container = container.NewStack(), container.NewStack()
var loupe *fynewidgets.Loupe
var thumbs *fynewidgets.ThumbnailGrid

func main() {

	startbus()

	myapp = app.NewWithID("com.glenn.fynewidget.widgettester")
	cloud.Enable(myapp)
	myprefs = myapp.Preferences()

	w := myapp.NewWindow("Test Progress Bars")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(600, 600))
	w.SetOnClosed(savepreferences)

	w.ShowAndRun()
}

func startbus() {
	Bus = eventbus.NewEventBus()
	Bus.SubscribeCallback("stringlist:notes", func(topic string, data interface{}) {
		myprefs.SetStringList("notes", data.([]string))
		savepreferences()
	})
}

func gui() *fyne.Container {

	prefnotes := myprefs.StringListWithFallback("notes", []string{})
	notes = fynewidgets.NewPermanentNotes(prefnotes, Bus)

	status = fynewidgets.NewEventLabel(Bus)


	top := container.NewHBox(
		widget.NewButtonWithIcon("All/None",theme.GridIcon(),allornone),
		widget.NewButtonWithIcon("Notes", theme.DocumentIcon(), shownotes),
		// widget.NewButton("Test Progress Widget", testProgressBars),
		widget.NewButton("Test Thumbnails", testThumbnailLoading),
		// widget.NewButton("Set Cloud Provider", cloudstorage),
		widget.NewButton("Test Grid",testimagegrid),
	)

	progresscell := container.NewBorder(nil, nil, nil, fynewidgets.NewMachineInfo(Bus), progress)
	bottom := container.NewBorder(nil, nil, nil, progresscell, status)
	loupe = fynewidgets.NewLoupe(image.Pt(200,200), 1)
	return container.NewBorder(top, bottom, container.NewBorder(nil, loupe, nil, nil, left), nil, centre)
}

func cloudstorage() {
	// Bus.Publish("text:status", fyne.CurrentApp().Storage().RootURI().Path())

	description := myapp.CloudProvider().ProviderDescription()
	log.Println(description)
	storage := myapp.Storage()
	log.Println(storage.RootURI().Path())
	

	// Bus.Publish("text:status", fmt.Sprintf("%d notes found in %s", len(notes.Notes), myapp.Storage().RootURI().Path()))

	cloud.ShowSettings(myapp, fyne.CurrentApp().Driver().AllWindows()[0])

}

func testProgressBars() {
	n := 10

	names := []string{}
	for i := 0; i < n; i++ {
		names = append(names, fmt.Sprintf("Phase iTask with a long name, really long %d", i))
	}

	progressbar := fynewidgets.NewMultiTaskProgress(names, Bus)

	progress.RemoveAll()
	progress.Add(progressbar)

	wg := &sync.WaitGroup{}
	wg.Add(len(names))

	for i := range names {
		go func() {
			delay := rand.Intn(100) + 10
			for j := 0; j <= 100; j += 1 {
				if Bus.HasSubscribers("progress:multi") {
					Bus.Publish("progress:multi", &fynewidgets.TaskProgressMessage{Task: i, Value: float64(j) / 100.0})
					testBusyWork(100000 * delay)
				}
			}
		}()
	}
}

func testBusyWork(N int) {
	var a, b int
	array := make([]int, N)
	for i := 0; i < len(array); i++ {
		a = array[i]
		b = array[len(array)-1-i]
		array[i], array[len(array)-1-i] = b, a
	}
	array = []int{}
}

func savepreferences() {
	myprefs.SetStringList("notes", notes.Notes)
}

func shownotes() {

	if notes == nil {
		return
	}

	box := container.NewStack(notes)

	dlg := dialog.NewCustomConfirm("Your notes", "Save to File", "Dismiss", box, func(b bool) {
		if !b {
			return
		}
		notes.ExportNotes()

	}, fyne.CurrentApp().Driver().AllWindows()[0])

	dlg.Resize(fyne.NewSize(600, 400))
	dlg.Show()

}

func testThumbnailLoading() {
	dialog.ShowFolderOpen(func(uc fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, myapp.Driver().AllWindows()[0])
			return
		}

		if uc == nil {
			return
		}

		log.Println(uc.Path())
		lister, err := storage.ListerForURI(uc)
		if err != nil {
			dialog.ShowError(err, myapp.Driver().AllWindows()[0])
			return
		}

		uris, err := lister.List()
		if err != nil {
			dialog.ShowError(err, myapp.Driver().AllWindows()[0])
			return
		}

		thumbs, err = fynewidgets.NewThumbnailGrid(uris, 200,200, 12, Bus)
		if err != nil {
			dialog.ShowError(err, myapp.Driver().AllWindows()[0])
			return
		}

		ch := Bus.Subscribe("image:thumbnail")
		go func() {
			for x := range ch {
				uri := x.Data.(fyne.URI)
				log.Println("Thumbnail selected:", uri.Path())
				pz, err := fynewidgets.NewPanZoomCanvasFromFile(uri, image.Pt(100, 100), Bus)
				if err != nil {
					dialog.ShowError(err, myapp.Driver().AllWindows()[0])
					return
				}
				log.Println("Opened", pz.URI().Path())
				pz.SetLoupe(loupe)
				centre.RemoveAll()
				centre.Add(pz)
				centre.Refresh()
				log.Println("refreshed")
				x.Done()
			}
		}()

		left.RemoveAll()
		left.Add(container.NewVScroll(thumbs))
		left.Refresh()
		log.Println("left refreshed")

	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

func allornone(){

	if thumbs==nil{
		return
	}
	thumbs.ToggleSelection()
}

func testimagegrid(){

	if thumbs==nil{
		return
	}

	if len(left.Objects)==0{
		return
	}

	if len(thumbs.SelectedFiles())==0{
		return
	}

	g,err:=fynewidgets.NewSynchronisedImageGrid(4,Bus)
	if err!=nil{
		dialog.ShowError(err,fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	centre.RemoveAll()
	centre.Add(g)
	g.SetImages(thumbs.SelectedFiles())
	g.Refresh()

}