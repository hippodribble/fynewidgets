package main

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/go-loremipsum/loremipsum"
	"github.com/hippodribble/fynewidgets"
)

func main() {
	ap := app.New()
	if drv, ok := ap.Driver().(desktop.Driver); ok {
		splash := drv.CreateSplashWindow()
		splash.SetContent(gui())
		splash.Resize(fyne.NewSize(800, 800))
		splash.ShowAndRun()
	}
} 

func gui() fyne.CanvasObject {
	l := loremipsum.New()
	c := container.NewAdaptiveGrid(5)
	for range 10 {
		word := l.Word()
		b := fynewidgets.NewBulletButton(word, 2, func() { fmt.Println("A Button") })
		b.OnClicked = func() { fmt.Println(b.NPresses, "presses"); time.Sleep(time.Second) }
		c.Add(b)
	}
	for i := range 10 {
		b := fynewidgets.NewPolygonButton(l.Word(), 2, i+3, func() { fmt.Println("A Polygon Button") }, color.RGBA{255, 255, 0, 255})
		b.OnClicked = func() { fmt.Println(b.NPresses, "presses") }
		c.Add(b)
	}

	for i := range 10 {
		b := fynewidgets.NewRunningTaskButton(l.Word(), func() { fmt.Println("A Polygon Button") })
		b.OnClicked = func() { fmt.Println(b.NPresses, "presses"); time.Sleep(time.Millisecond * time.Duration(i*100)) }
		c.Add(b)
	}
	return container.NewPadded(container.NewVBox(c))
}
