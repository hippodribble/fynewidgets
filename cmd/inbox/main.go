package main

import (
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/go-loremipsum/loremipsum"
	"github.com/hippodribble/fynewidgets"
)

var w fyne.Window

func main() {

	ap := app.New()
	w = ap.NewWindow("Inbox v0.1")
	// w.SetFullScreen(true)
	w.SetContent(gui())
	w.Resize(fyne.NewSize(1000, 1000))
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	whiteboard := fynewidgets.NewDesktop()

	// fg := theme.Color(theme.ColorNameForeground)
	// bg := theme.Color(theme.ColorNameBackground)

	whiteboard.Add(fynewidgets.NewDefaultItem())
	l := loremipsum.NewWithSeed(time.Now().Unix())
	catlist := []string{}
	catlist = append(catlist, "Inbox")
	catlist = append(catlist, "Navigation")
	catlist = append(catlist, "Positioning")
	catlist = append(catlist, "HSE")
	catlist = append(catlist, "Client")
	catlist = append(catlist, "Agent")
	for range 37 {
		cat := catlist[rand.Intn(len(catlist))]
		d := fynewidgets.NewItemData(l.Word(), l.Paragraph(), cat)
		whiteboard.Add(fynewidgets.NewItem(d))
	}
	return whiteboard
}
