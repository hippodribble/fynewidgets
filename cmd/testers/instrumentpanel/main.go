package main

import (
	"fmt"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/hippodribble/fynewidgets"
)

var lamps *fynewidgets.InstrumentPanel
var list, names []string
var ch chan *fynewidgets.InstrumentLamp

func main() {
	list = makeTriplets()
	go makechannel()

	N := len(list)
	names = []string{}
	for range 12 {
		names = append(names, list[rand.Intn(N)])
	}

	ap := app.New()
	w := ap.NewWindow("Test")
	// lamp := fynewidgets.NewInstrumentLamp(name, fyne.NewSize(100, 100))
	lamps = fynewidgets.NewInstrumentPanel(names, fyne.NewSize(80, 60), 12)
	lamps.SetChannel(ch)
	w.SetContent(lamps)
	w.Resize(fyne.NewSize(100, 100))
	w.ShowAndRun()
}

func makechannel() {
	ch = make(chan *fynewidgets.InstrumentLamp)
	for {
		lamp := <-ch
		fmt.Println("monitorChannel: ", lamp.Name, lamp.Status.Text)
	}
}

func makeTriplets() []string {
	start := int32('A')
	end := int32('Z')
	testblk := []rune{}

	for i := start; i <= end; i++ {
		testblk = append(testblk, rune(i))
	}
	// fmt.Println(string(testblk))
	n := end - start + 1
	list := make([]string, n*n*n)
	count := 0
	for i := start; i <= end; i++ {
		for j := start; j <= end; j++ {
			for k := start; k <= end; k++ {
				array := []rune{i, j, k}
				list[count] = string(array)
				count++
			}
		}
	}
	return list
}
