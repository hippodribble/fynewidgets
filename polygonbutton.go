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

// A rounded polygonal button which flashes on press, and tracks the total press count
type PolygonButton struct {
	widget.BaseWidget
	r                       *canvas.Polygon
	l                       *canvas.Text
	OnClicked               func()
	busycolor, notbusycolor color.Color
	nflashes, NPresses      int
	busy                    bool
}

func NewPolygonButton(text string, nFlashes, nSides int, onClicked func(), c color.Color) *PolygonButton {
	r := canvas.NewPolygon(uint(nSides), c)
	r.CornerRadius = 5
	r.SetMinSize(fyne.NewSize(50, 50))
	l := canvas.NewText(string(text[0]), theme.Color(theme.ColorNameForeground))
	l.Alignment = fyne.TextAlignCenter
	l.TextStyle.Bold = true

	b := &PolygonButton{
		r:            r,
		l:            l,
		OnClicked:    onClicked,
		busycolor:    c,
		notbusycolor: color.Gray{64},
		nflashes:     nFlashes,
	}
	b.r.FillColor = b.notbusycolor
	b.ExtendBaseWidget(b)
	return b
}

func (p *PolygonButton) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(p.r, p.l)
	return widget.NewSimpleRenderer(c)
}

func (p *PolygonButton) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonPrimary {
		if p.busy {
			return
		}
		p.busy = true
		p.NPresses++
		go func() {
			p.OnClicked()
			p.busy = false
		}()

		go func() {
			for range p.nflashes {
				fyne.DoAndWait(func() {
					p.r.FillColor = p.busycolor
					p.r.Refresh()
				})
				time.Sleep(time.Millisecond * 50)
				fyne.DoAndWait(func() {
					p.r.FillColor = p.notbusycolor
					p.r.Refresh()
				})
				time.Sleep(time.Millisecond * 50)
			}
		}()
	}
}

func (p *PolygonButton) MouseUp(e *desktop.MouseEvent) {

}

func (p *PolygonButton) MinSize() fyne.Size {
	return fyne.NewSize(20, 20)
}
