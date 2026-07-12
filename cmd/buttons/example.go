package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
)

var ch = make(chan interface{})

func main() {
	ap := app.New()
	w := ap.NewWindow("Test Button")
	w.Resize(fyne.NewSize(800, 800))
	w.SetContent(gui())
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	status := fynewidgets.NewStatusProgress(ch)
	progchan := make(chan float64)
	outchan := make(chan any)
	go func() {
		for v := range outchan {
			ch <- fmt.Sprintf("Long-run progress result was %v", v)
		}
	}()
	ns := []any{}
	for i := range 100 {
		ns = append(ns, float64(i))
	}
	pf := fynewidgets.NewProgressFunc(Sum, progchan, outchan, true, ns...)
	b1 := fynewidgets.NewTaskButton("Add Stuff", pf)
	b3 := fynewidgets.NewPolygonButton("5", 2, 5, func() { ch <- "You clicked on a polygon button" }, color.RGBA{255, 128, 0, 255})
	b4 := fynewidgets.NewBulletButton("Press this bullet!", 3, func() { ch <- "Bullet button pressed" })
	b5 := fynewidgets.NewSliderLabel(0, 1, .01)
	b6 := fynewidgets.NewRunningTaskButton("Running Task", func() {
		ch <- "Running Task!"
		time.Sleep(time.Millisecond * 500)
	})
	tempButtons := []*widget.Button{}
	for _, text := range []string{"1/hello", "2/world", "3/go", "4/lang", "5/blip", "10/plop"} {
		tempButtons = append(tempButtons, widget.NewButton(text, func() { ch <- text }))
	}
	b2, err := fynewidgets.NewPolygonalButtons(tempButtons, 150,50)
	if err != nil {
		log.Fatalln("Bad Buttons")
	}

	tempButtons2 := []*widget.Button{}
	for _, text := range []string{"1/hello", "2/world", "3/go", "4/lang", "5/blip", "10/plop"} {
		tempButtons2 = append(tempButtons2, widget.NewButton(text, func() { ch <- text }))
	}
	b7 := fynewidgets.NewRadialButtons(tempButtons2, 200)

	return container.NewBorder(
		nil, status,
		nil, nil,
		container.NewVBox(
			container.NewGridWithColumns(1,
				b1,
				b3,
				container.NewHBox(b4),
				b5,
				b6,
			),
			b2,
			b7,
		),
	)
}

func Sum(cp chan float64, vars ...any) any {
	sum := 0.0
	n := float64(len(vars))
	for i, v := range vars {
		cp <- float64(i+1) / n
		x := v.(float64)
		sum += x
		time.Sleep(time.Millisecond * 5)
	}
	return sum
}
