package fynewidgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// A RunningTaskButton stays lit as long as the underlying function is active
//
//	It also tracks the press count
type RunningTaskButton struct {
	widget.BaseWidget
	r                       *canvas.Rectangle
	l                       *canvas.Text
	OnClicked               func()
	busycolor, notbusycolor color.Color
	NPresses                int
	busy                    bool
}

func NewRunningTaskButton(text string, onClicked func()) *RunningTaskButton {
	r := canvas.NewRectangle(color.Gray{64})
	r.CornerRadius = canvas.RadiusMaximum
	l := canvas.NewText(text, theme.Color(theme.ColorNameForeground))
	l.Alignment = fyne.TextAlignCenter
	l.TextStyle.Bold = true

	b := &RunningTaskButton{
		r:            r,
		l:            l,
		OnClicked:    onClicked,
		busycolor:    color.RGBA{255, 255, 0, 255},
		notbusycolor: color.Gray{64},
	}
	b.r.FillColor = b.notbusycolor
	b.ExtendBaseWidget(b)
	return b
}

func (r *RunningTaskButton) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(r.r, r.l)
	return widget.NewSimpleRenderer(c)
}

func (r *RunningTaskButton) Text() string {
	return r.l.Text
}

func (r *RunningTaskButton) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonPrimary {
		if r.busy {
			return
		}
		r.busy = true
		r.NPresses++

		go func() {

			fyne.DoAndWait(func() {
				r.r.FillColor = r.busycolor
				r.r.Refresh()
			})

			r.OnClicked()

			fyne.DoAndWait(func() {
				r.r.FillColor = r.notbusycolor
				r.r.Refresh()
			})

			r.busy = false

		}()
	}
}

func (r *RunningTaskButton) MouseUp(e *desktop.MouseEvent) {

}
