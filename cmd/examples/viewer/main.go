package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	eventbus "github.com/dtomasi/go-event-bus/v3"
)

var myapp fyne.App
var mainwindow fyne.Window
var myprefs fyne.Preferences
var bus eventbus.EventBus=*eventbus.NewEventBus() 

func main(){
	myapp=app.NewWithID("com.github.hippodribble.examples.viewer")
	mainwindow=myapp.NewWindow("Image Viewer")
	myprefs=myapp.Preferences()
	loadprefs()
	mainwindow.SetCloseIntercept(shutdown)
	makegui()
	mainwindow.ShowAndRun()
}

func loadprefs(){
	w:=myprefs.FloatWithFallback("mainwindow.width",800)
	h:=myprefs.FloatWithFallback("mainwindow.height",800)
	mainwindow.Resize(fyne.NewSize(float32(w),float32(h)))
}

func shutdown(){
	saveprefs()
	mainwindow.Close()
}

func saveprefs(){
	myprefs.SetFloat("mainwindow.width",float64(mainwindow.Canvas().Size().Width))
	myprefs.SetFloat("mainwindow.height",float64(mainwindow.Canvas().Size().Height))
}

