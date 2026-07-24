package main

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	eventbus "github.com/dtomasi/go-event-bus/v3"
	"github.com/hippodribble/fynewidgets"
	elevationImage "github.com/hippodribble/fynewidgets/image"
)

var APPSIZE = fyne.NewSize(2000, 1000)
var prefs fyne.Preferences
var top, bottom, left, right *fyne.Container
var centre *container.MultipleWindows
var bus *eventbus.EventBus
var w fyne.Window
var loupe *fynewidgets.Loupe

func main() {
	ap := app.NewWithID("com.github.hippodribble.fynewidgets.imcalc")
	prefs = ap.Preferences()
	bus = eventbus.NewEventBus()

	w = ap.NewWindow("Image Calc v0.1")
	w.SetContent(gui())
	w.SetMainMenu(menu())
	w.Resize(APPSIZE)
	w.ShowAndRun()

}

func gui() fyne.CanvasObject {

	loupe = fynewidgets.NewLoupe(image.Pt(100, 100), 2)
	centre = container.NewMultipleWindows()
	top = container.NewStack()
	bottom = container.NewStack()
	left = container.NewStack(
		container.NewVBox(loupe),
	)
	right = container.NewStack()
	b := container.NewBorder(top, bottom, left, right, centre)
	return b

}

func menu() *fyne.MainMenu {
	itemFileOpen := fyne.NewMenuItem("Open", fileOpen)
	itemFileOpen.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierShortcutDefault}
	menuFile := fyne.NewMenu("File", itemFileOpen)
	mm := fyne.NewMainMenu(menuFile)
	return mm
}

// load an image to a pan zoom widget
func fileOpen() {
	dlg := dialog.NewFileOpen(
		func(uc fyne.URIReadCloser, err error) {

			if err != nil {
				fmt.Println(err)
				return
			}

			if uc == nil {
				fmt.Println("Cancelled")
				return
			}

			s, _ := parentFolderToString(uc)
			prefs.SetString("lastfolder", s)
			img := loadImage(uc.URI().Path())
			pz, err := fynewidgets.NewPanZoomCanvasFromImage(img, image.Pt(200, 200), bus, uc.URI().Name())
			if err != nil {
				return
			}

			pz.SetLoupe(loupe)

			win := container.NewInnerWindow("File Name", pz)
			win.OnMaximized = func() { win.Resize(centre.Size()) }
			win.OnMinimized = func() { win.Resize(fyne.NewSize(250, 250)) }
			win.Resize(fyne.NewSize(750, 750))
			centre.Windows = []*container.InnerWindow{win}
			centre.Refresh()
		},w,
	)
	dlg.SetLocation(stringToFolderURI(prefs.String("lastfolder")))
	// if dlg==nil{return}
	fyne.Do(func() {dlg.Resize(fyne.NewSize(750, 750))})
	dlg.Show()
}

// returns the string path of the parent folder of the file URI
func parentFolderToString(uc fyne.URIReadCloser) (string, error) {
	u := uc.URI()
	p, err := storage.Parent(u)
	if err != nil {
		return "", err
	}
	return p.Path(), nil
}

func stringToFolderURI(s string) fyne.ListableURI {
	if s == "" {
		return nil
	}
	uri := storage.NewFileURI(s)
	lu, err := storage.ListerForURI(uri)
	if err != nil {
		return nil
	}
	return lu
}

func loadImage(path string) *elevationImage.ElevationImage {
	e, err := elevationImage.NewElevationImage(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	e.Rescale(.99)
	return e
}
