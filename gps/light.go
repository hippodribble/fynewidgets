package gps

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Light struct {
	widget.BaseWidget
	c          *canvas.Circle
	background *canvas.Rectangle
	color      color.Color
}

func NewLight(c color.Color, size float32) *Light {
	circle := canvas.NewCircle(c)
	background := canvas.NewRectangle(color.Transparent)
	background.SetMinSize(fyne.NewSize(size, size))

	circle.StrokeColor = c
	circle.StrokeWidth = 1
	l := &Light{c: circle, color: c, background: background}
	l.ExtendBaseWidget(l)
	return (*Light)(l)
}

func (l *Light) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(l.background, l.c))
}

func (l *Light) Off() {
	l.c.Hide()
}

func (l *Light) On() {
	if l.c.Hidden {
		l.c.Show()
	}
}

func (l *Light) Flash(milliseconds int) {
	go func() {
		fyne.Do(l.On)
		time.Sleep(time.Millisecond * time.Duration(milliseconds))
		fyne.Do(l.Off)
	}()
}
func (l *Light) MinSize() fyne.Size { return fyne.NewSize(20, 20) }
