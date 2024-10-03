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
}

func NewClock() *Clock {

	C := Clock{}
	C.label = widget.NewLabel("Mon 15:45:06")
	C.label.TextStyle.Bold = true
	C.label.TextStyle.Monospace = true

	C.ticker = *time.NewTicker(time.Millisecond * 500)

	go func() {
		for TIME := range C.ticker.C {
			C.label.SetText(TIME.Format("Mon 15:04:05"))

		}
	}()

	C.ExtendBaseWidget(&C)
	return &C

}

func (C *Clock) CreateRenderer() fyne.WidgetRenderer {

	return widget.NewSimpleRenderer(C.label)

}

func (C *Clock) MouseDown(e *desktop.MouseEvent) {
}

func (C *Clock) MouseUp(e *desktop.MouseEvent) {
}
