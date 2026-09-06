package fynewidgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// A Dial gauge widget that can be used to display numerical data.
type Dial struct {
	widget.BaseWidget
	donut                 *canvas.Arc
	display               *canvas.Text
	radius                float32
	leftClick, rightClick func()
	min, max, value       float32
	minsize               fyne.Size
}

func NewDial(c color.Color, cutout float32, min, max float32, fLeft, fRight func()) *Dial {
	d := canvas.NewDoughnutArc(-135, 135, c)
	d.CutoutRatio = cutout
	label := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	label.TextStyle.Bold = true
	label.Alignment = fyne.TextAlignCenter
	donut := &Dial{donut: d, leftClick: fLeft, rightClick: fRight, min: min, max: max, minsize: fyne.NewSize(25, 25), display: label}
	label.TextSize = donut.minsize.Width / 3
	donut.ExtendBaseWidget(donut)
	return donut
}

func (d *Dial) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(d.donut, d.display))
}

func (d *Dial) MinSize() fyne.Size {
	return d.minsize
}

func (d *Dial) SetSize(sz fyne.Size) {
	d.minsize = sz
	d.display.TextSize = d.minsize.Width / 3
	d.Refresh()
}

// SetMin is used to specify the dial range
func (d *Dial) SetMin(v float32) {
	d.donut.StartAngle = v
	d.Refresh()
}

// SetMax is used to specify the dial range
func (d *Dial) SetMax(v float32) {
	d.donut.EndAngle = v
	d.Refresh()
}

func (d *Dial) SetValue(v float32) {
	if v>d.max{v-=d.max}
	d.value = v
	f := (d.value - d.min) / (d.max - d.min)
	d.donut.EndAngle = f*270 - 135
	// fmt.Println(f,d.donut.EndAngle)
	d.display.Text = fmt.Sprintf("%.1f", v)
	d.Refresh()
}

func (d *Dial) SetLeftClick(f func())  { d.leftClick = f }
func (d *Dial) SetRightClick(f func()) { d.rightClick = f }

func (d *Dial) Value() float32 { return d.value }

func (d *Dial) MouseDown(e *desktop.MouseEvent) {
	switch e.Button {
	case desktop.MouseButtonPrimary:
		go func() {d.leftClick()}()
	case desktop.MouseButtonSecondary:
		go func() {d.rightClick()}()
	}
}
func (d *Dial) MouseUp(e *desktop.MouseEvent) {}
