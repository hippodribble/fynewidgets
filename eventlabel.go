package fynewidgets

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	eventbus "github.com/dtomasi/go-event-bus/v3"
)

type EventLabel struct {
	widget.Label
	bus *eventbus.EventBus
}

func NewEventLabel(bus *eventbus.EventBus) *EventLabel {
	label := &EventLabel{bus: bus}
	label.Label = *widget.NewLabel("")
	label.Label.Truncation = fyne.TextTruncateEllipsis
	lasttime := time.Now()

	ch := bus.Subscribe("status:show")
	go func() {
		for x := range ch {
			defer func() {
				if r := recover(); r != nil {
					log.Println(x.Data)
					log.Println(r)
				}
			}()
			if s, ok := x.Data.(string); ok {
				fyne.Do(func() { label.SetText(s) })
			}
			x.Done()
		}
	}()

	go func() {
		for range time.NewTicker(1 * time.Second).C {
			if time.Since(lasttime) > 2*time.Second {
				fyne.Do(func() { label.SetText("") })
			}
		}
	}()

	return label

}

// EventLabelChannel reads a text channel and displays the data in a label.
// the label is deleted automatically after delay seconds
type EventLabelChannel struct {
	widget.Label
	ch chan string
	lasttime time.Time
	delay time.Duration
}

func NewEventLabelChannel(ch chan string,delay int) *EventLabelChannel {
	label := &EventLabelChannel{ch: ch}
	label.Label = *widget.NewLabel("")
	label.Label.Truncation = fyne.TextTruncateEllipsis
	label.lasttime = time.Now()
	label.delay=time.Millisecond*time.Duration(delay*1000)

	go func() {
		for x := range ch {
			fyne.Do(func() { label.SetText(x) })
			label.lasttime=time.Now()
		}
	}()

	go func() {
		for range time.NewTicker(1 * time.Second).C {
			if time.Since(label.lasttime) > label.delay {
				fyne.Do(func() { label.SetText("") })
			}
		}
	}()

	return label

}
