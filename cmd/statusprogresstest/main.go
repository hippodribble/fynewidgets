package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets/statusprogress"
)

var sp *statusprogress.StatusProgressWidget
var statuschannel chan interface{} = make(chan interface{})

func main() {
	ap := app.New()
	w := ap.NewWindow("Test Status + Progress Bar Widget")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(1200, 900))
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	bProgress := widget.NewButton("Test Progress Bar", testProgressBar)
	bStatus := widget.NewButton("Test Message", testMessage)
	bmessagequeue := widget.NewButton("Test Message Queue", testMessageQueue)
	bInfiniteProgress := widget.NewButton("Test Ongoing Progress", testInfiniteProgress)
	top := container.NewHBox(bProgress, bStatus, bmessagequeue, bInfiniteProgress)

	sp = statusprogress.NewStatusProgress(statuschannel)

	log.Println(sp.CurrentMessage())

	b := container.NewBorder(top, sp, nil, nil, nil)
	return b
}

func testProgressBar() {
	statuschannel <- 0.0
	for i := 0; i < 10; i++ {
		statuschannel <- -0.1
		time.Sleep(time.Millisecond * 100)
	}
	statuschannel <- 0.0
}

func testMessage() {
	go func() {
		msg := statusprogress.Message{Text: fmt.Sprintf("%v Here is a status update, 3 seconds timer", time.Now()), Duration: 3}
		statuschannel <- msg
		log.Println("Sent status update")
	}()
}

func testMessageQueue() {
	go func() {
		for i := 0; i < 10; i++ {
			delay := int(rand.Intn(5000) + 500)
			msg := statusprogress.Message{Text: fmt.Sprintf("%v New Message will be here for %d ms", time.Now(), delay), Duration: 5}
			statuschannel <- msg
			time.Sleep(time.Millisecond * time.Duration(delay))
		}
	}()
}

func testInfiniteProgress() {
	go func() {
		statuschannel <- statusprogress.Message{Text: "Start", Duration: 3}
		statuschannel <- 2.0
		time.Sleep(time.Second * 3)
		statuschannel <- statusprogress.Message{Text: "Stop", Duration: 3}
		statuschannel <- -2.0
	}()
}
