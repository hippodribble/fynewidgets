package fynewidgets

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type BulletButton struct {
	widget.BaseWidget
	r                       *canvas.Rectangle
	l                       *canvas.Text
	OnClicked               func()
	done                    bool
	timer                   *time.Timer
	busycolor, notbusycolor color.Color
	nflashes, NPresses      int
	busy                    bool
}

func NewBulletButton(text string, nFlashes int, onClicked func()) *BulletButton {
	r := canvas.NewRectangle(color.Gray{64})
	r.CornerRadius = canvas.RadiusMaximum
	l := canvas.NewText(text, theme.Color(theme.ColorNameForeground))
	l.Alignment = fyne.TextAlignCenter
	l.TextStyle.Bold = true

	b := &BulletButton{
		r:            r,
		l:            l,
		OnClicked:    onClicked,
		busycolor:    color.RGBA{255, 255, 0, 255},
		notbusycolor: color.Gray{64},
		nflashes:     nFlashes,
	}
	b.r.FillColor = b.notbusycolor
	b.ExtendBaseWidget(b)
	return b
}

func (r *BulletButton) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(r.r, r.l)
	return widget.NewSimpleRenderer(c)
}

func (r *BulletButton) Text() string {
	return r.l.Text
}

func (r *BulletButton) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonPrimary {
		if r.busy {
			return
		}
		r.busy = true
		r.NPresses++
		go func() {
			r.OnClicked()
			r.busy = false
		}()
		go func() {
			for range r.nflashes {
				fyne.DoAndWait(func() {
					r.r.FillColor = r.busycolor
					r.r.Refresh()
				})
				time.Sleep(time.Millisecond * 50)
				fyne.DoAndWait(func() {
					r.r.FillColor = r.notbusycolor
					r.r.Refresh()
				})
				time.Sleep(time.Millisecond * 50)
			}
		}()
	}
}

func (r *BulletButton) MouseUp(e *desktop.MouseEvent) {

}
