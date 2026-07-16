package fynewidgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type Clock struct {
	widget.BaseWidget
	ticker time.Ticker
	label  *widget.Label
	paused bool
}

func NewClock() *Clock {
	C := Clock{}
	C.label = widget.NewLabel("Mon 15:45:06")
	C.label.TextStyle.Bold = true
	C.label.TextStyle.Monospace = true
	C.ticker = *time.NewTicker(time.Millisecond * 100)
	go func() {
		for TIME := range C.ticker.C {
			if C.paused {
				continue
			}
			fyne.Do(func() { C.label.SetText(TIME.Format("Mon 15:04:05")) })
		}
	}()

	C.ExtendBaseWidget(&C)
	return &C

}

func (C *Clock) CreateRenderer() fyne.WidgetRenderer {

	return widget.NewSimpleRenderer(C.label)

}

func (C *Clock) MouseDown(e *desktop.MouseEvent) {
	C.paused = !C.paused
	if C.paused {
		C.label.Importance = widget.DangerImportance
	} else {
				C.label.Importance=widget.MediumImportance
	}
	C.Refresh()
}

func (C *Clock) MouseUp(e *desktop.MouseEvent) {
}
