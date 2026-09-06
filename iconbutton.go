package fynewidgets

import (
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type IconButton struct {
	widget.BaseWidget
	icon *widget.Icon
	c    *canvas.Circle
	f    func()
	fill color.Color
}

func NewIconButton(icon *widget.Icon, f func()) *IconButton {
	fill := color.RGBA{64, 32, 0, 255}
	b := &IconButton{icon: icon, f: f, c: canvas.NewCircle(fill)}
	b.ExtendBaseWidget(b)
	return b
}

type icblayout struct {
	b *IconButton
}

func (l icblayout) MinSize(obs []fyne.CanvasObject) fyne.Size {
	return l.b.icon.MinSize().AddWidthHeight(25, 25)
}

func (l icblayout) Layout(obs []fyne.CanvasObject, sz fyne.Size) {
	for _, o := range obs {
		wi, hi := o.Size().Width, o.Size().Height
		w, h := sz.Width, sz.Height

		if p, ok := o.(*canvas.Circle); ok {
			d := min(sz.Width, sz.Height)
			p.Resize(fyne.NewSize(d, d))
			p.Move(fyne.NewPos(w/2-wi/2, h/2-hi/2))
		}

		if p, ok := o.(*widget.Icon); ok {
			// fmt.Println("Icon")
			p.Move(fyne.NewPos(w/2-wi/2, h/2-hi/2))
			p.Resize(p.MinSize())
			// fmt.Println(p.Size(), p.MinSize())
		}
	}
}



func (b *IconButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.New(icblayout{b}, b.c, b.icon))
}

func (b *IconButton) MouseDown(e *desktop.MouseEvent) {
	if b.isInCircle(e) {
		go func() {
			store := b.c.FillColor
			b.c.FillColor = theme.Color(theme.ColorNameForeground)
			fyne.Do(func() { b.Refresh() })
			time.Sleep(time.Millisecond * 50)
			b.c.FillColor = store
			fyne.Do(func() { b.Refresh() })
		}()
	}
}
func (b *IconButton) MouseUp(e *desktop.MouseEvent) {b.f()}

func (b *IconButton) isInCircle(e *desktop.MouseEvent) bool {
	x, y := e.Position.X, e.Position.Y
	w, h := b.c.Size().Width/2, b.c.Size().Height/2
	C := b.c.Position1.AddXY(w, h)
	dx, dy := x-C.X, y-C.Y
	r := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if r < w && r < h {
		return true
	}
	return false
}
