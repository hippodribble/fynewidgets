package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var left, right, top, bottom, centre *fyne.Container
var statuschannel chan interface{} = make(chan interface{})

func main() {


	appp := app.NewWithID("com.github.hippodribble.fynewidgets.vipstest")
	w := appp.NewWindow("VIPSTest")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(1200, 900))

	w.ShowAndRun()
}
