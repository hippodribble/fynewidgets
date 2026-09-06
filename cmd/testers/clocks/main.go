package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
	"golang.org/x/image/colornames"
)

func main() {
	ap := app.New()
	w := ap.NewWindow("Clock Tests")
	w.Resize(fyne.NewSize(400, 400))
	w.SetContent(gui())
	w.ShowAndRun()

}

func gui() fyne.CanvasObject {
	clock := fynewidgets.NewClock()
	dial := fynewidgets.NewDial(colornames.Yellow, .85, 0, 100, func() { fmt.Println("left") }, func() { fmt.Println("right") })
	dial.SetValue(0)
	dial.SetSize(fyne.NewSize(100, 100))
	go func() {
		for {
			time.Sleep(time.Second / 20)
			fyne.Do(func() {
				dial.SetValue(dial.Value() + 1)
			})
		}
	}()

	di := fynewidgets.NewDialInfinite(60, 5, 200, .85, .9, colornames.Olive, colornames.Yellow, 360, "")
	v := 0.0
	go func() {
		for {
			time.Sleep(time.Second / 20)
			v += 3.6
			fyne.Do(func() { di.SetValue(v) })
		}
	}()

	bb := fynewidgets.NewBulletButton("Press me", 2, func() { fmt.Println("pressed") })
	bb.OnClicked = func() { fmt.Println(bb.NPresses, "presses") }

	ch := make(chan string)
	ec := fynewidgets.NewEventLabelChannel(ch,2)

	ib := fynewidgets.NewIconButton(widget.NewIcon(theme.CalendarIcon()), func() {
		ch<-"pressed icon button"
	})


	return container.NewVBox(container.NewHBox(clock, dial, di, bb, ib), ec)
}
