package fynewidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ToggleButton struct {
	widget.BaseWidget
	OnClicked func()
	Pressed   bool
	text      *canvas.Text
	rect      *canvas.Rectangle
}

func NewToggleButton(text string, onClicked func()) *ToggleButton {
	tb := &ToggleButton{
		OnClicked: onClicked,
		Pressed:   false,
		text:      canvas.NewText(text, theme.Color(theme.ColorNameForeground)),
	}
	tb.text.Alignment = fyne.TextAlignCenter
	tb.text.TextStyle.Bold = true
	tb.ExtendBaseWidget(tb)
	return tb
}

func (tb *ToggleButton) CreateRenderer() fyne.WidgetRenderer {
	tb.rect = canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	tb.rect.CornerRadius = canvas.RadiusMaximum
	return widget.NewSimpleRenderer(container.NewStack(tb.rect, tb.text))
}

func (tb *ToggleButton) Tapped(_ *fyne.PointEvent) {
	tb.Pressed = !tb.Pressed
	if tb.Pressed {
		tb.rect.FillColor = theme.Color(theme.ColorNameForeground)
		tb.text.Color = theme.Color(theme.ColorNameBackground)
		tb.OnClicked()
	} else {
		tb.rect.FillColor = theme.Color(theme.ColorNameBackground)
		tb.text.Color = theme.Color(theme.ColorNameForeground)
	}
	fyne.Do(func() { tb.Refresh() })
}
