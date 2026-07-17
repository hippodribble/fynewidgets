package fynewidgets

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type URIButton struct {
	widget.BaseWidget
	i         *widget.FileIcon
	r         *canvas.Rectangle
	busy      bool
	OnClicked func()
}

func NewURIButton(uri fyne.URI, f func()) *URIButton {
	i := widget.NewFileIcon(uri)
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(i.Size())
	b := &URIButton{OnClicked: f, i: i}
	b.ExtendBaseWidget(b)
	return b
}

func (u *URIButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(u.i)
}

func (b *URIButton) MouseDown(e *desktop.MouseEvent) {
	if b.busy {
		return
	}
	b.busy = true
	go func() {
		b.OnClicked()
		b.busy = false
	}()
	go func() {
		fyne.DoAndWait(func() { b.i.Hidden = true; b.Refresh() })
		time.Sleep(time.Millisecond * 50)
		fyne.DoAndWait(func() { b.i.Show(); b.Refresh() })
	}()
}

func (b *URIButton) MouseUp(e *desktop.MouseEvent) {}
