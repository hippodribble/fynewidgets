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

	ch := bus.Subscribe("text:status")
	go func() {
		for x := range ch {
			defer func() {
				if r := recover(); r != nil {
					log.Println(x.Data)
					log.Println(r)
				}
			}()
			if s, ok := x.Data.(string); ok {
				label.SetText(s)
			}
			x.Done()
		}
	}()
	// bus.SubscribeCallback("text:status", func(topic string, data interface{}) {

	// })
	go func() {
		for range time.NewTicker(3 * time.Second).C {
			if time.Since(lasttime) > 30*time.Second {
				label.SetText("")
			}
		}
	}()

	return label

}
