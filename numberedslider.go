package fynewidgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SliderLabel struct {
	widget.BaseWidget
	value  binding.Float
	slider *widget.Slider
	label  *canvas.Text
}

func NewSliderLabel(min, max, step float64) *SliderLabel {
	v := binding.NewFloat()
	colour := color.RGBA{128, 128, 0, 255}
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantDark {
		colour = color.RGBA{255,255,0, 255}
	}
	l := canvas.NewText("", colour)
	l.TextSize = 36
	l.TextStyle.Bold = true
	l.Alignment = fyne.TextAlignCenter
	l.SetMinSize(fyne.NewSize(50, 36))
	// l.TextStyle.Italic = true
	// l.Importance = widget.DangerImportance

	s := widget.NewSliderWithData(min, max, v)
	s.Step = step
	sl := &SliderLabel{
		value:  v,
		slider: s,
		label:  l,
	}
	s.OnChanged = func(f float64) { l.Text = fmt.Sprintf("%.1f", f); l.Refresh() }
	s.OnChangeEnded = func(f float64) { l.Text = ""; l.Refresh() }

	sl.ExtendBaseWidget(sl)
	return sl
}

func (s *SliderLabel) CreateRenderer() fyne.WidgetRenderer {
	lay := &SliderLabelLayout{s}
	c := container.New(lay, s.slider, s.label)
	return widget.NewSimpleRenderer(c)
}

type SliderLabelLayout struct {
	sl *SliderLabel
}

func (s *SliderLabelLayout) Layout(os []fyne.CanvasObject, sz fyne.Size) {

	// H := sz.Height
	W := sz.Width

	middle := fyne.NewPos(W/2-os[1].Size().Width/2-12, 0)

	os[0].Resize(sz)
	os[0].Move(fyne.NewPos(0, 0))
	os[1].Move(middle)
	fmt.Printf("Widget:%v\nLabel: %v\nSlider: %v\nMiddle: %v\n\n", sz, os[1].Size(), os[0].Size(), middle)
}

func (s *SliderLabelLayout) MinSize(os []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 36)
}
