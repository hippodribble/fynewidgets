package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"net/netip"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
	"github.com/hippodribble/fynewidgets/network"
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

	b2, err := fynewidgets.NewPolygonalButtons(tempButtons, 150, 50)
	if err != nil {
		log.Fatalln("Bad Buttons")
	}

	radialcontainer := container.NewAdaptiveGrid(10)
	for range 10 {
		tempButtons2 := []*widget.Button{}
		for _, text := range []string{"1/hello", "2/world", "3/go", "4/lang", "5/blip", "10/plop"} {
			tempButtons2 = append(tempButtons2, widget.NewButton(text, func() { ch <- text }))
		}
		b7 := fynewidgets.NewRadialButtons(tempButtons2, 100)
		radialcontainer.Add(b7)
	}

	iconbuttons := container.NewAdaptiveGrid(15)
	for range 15 {
		b8 := fynewidgets.NewIconButton(widget.NewIcon(theme.CalendarIcon()), func() { ch <- "Calendar" })
		iconbuttons.Add(b8)
	}

	dials := container.NewAdaptiveGrid(10)
	for range 10 {
		red := uint8(rand.IntN(128) + 127)
		green := uint8(rand.IntN(256))
		blue := uint8(rand.IntN(128) + 127)
		c := color.RGBA{red, green, blue, 255}
		f1 := func() {
			fmt.Println("Left Click")
		}
		f2 := func() {
			fmt.Println("Right Click")
		}
		dial := fynewidgets.NewDial(c, .7, -100, 100, f1, f2)
		dial.SetLeftClick(func() {
			ch <- fmt.Sprintf("Left clicked at %.1f", dial.Value())
		})
		dial.SetSize(fyne.NewSize(50, 50))
		dials.Add(dial)
		var base float32 = float32(rand.Float32()*2000 - 1)
		go func() {
			for {
				fyne.Do(func() { dial.SetValue(float32(100 * math.Sin(float64(base)))) })
				base += .02
				time.Sleep(time.Millisecond * 100)
			}
		}()
	}
	// clock:=fynewidgets.NewClock()

	// dkblue := color.RGBA{128, 128, 255, 255}
	// ltblue := color.RGBA{192, 192, 255, 255}
	// gp, err := fynewidgets.NewGraphPaper(50, 10, 401, 401, dkblue, ltblue, color.White)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	add := netip.MustParseAddrPort("127.0.0.1:8080")
	node := network.NewTCPEndPoint(add)
	nv:=network.NewNodeView(&node,color.RGBA{255,255,0,255})

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
			radialcontainer,
			iconbuttons,
			dials,
			nv,
			// clock,
			// gp,
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
