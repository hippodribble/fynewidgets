package dataflow

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Parameter interface {
	IsInput() bool
}

type FloatParameter struct {
	widget.BaseWidget
	Name  string
	input bool
	Value binding.Float
	box   *canvas.Rectangle
	label *widget.Label
}

func NewFloatParameter(name string, input bool) *FloatParameter {
	box := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	box.StrokeColor = theme.Color(theme.ColorNameForeground)
	box.StrokeWidth = 1
	box.SetMinSize(fyne.NewSize(30, 30))
	boundvalue := binding.NewFloat()
	return &FloatParameter{Name: name, box: box, input: input, Value: boundvalue, label: widget.NewLabel(name)}
}

func (f *FloatParameter) CreateRenderer() fyne.WidgetRenderer {
	editor := widget.NewEntryWithData(binding.FloatToString(f.Value))
	fmt.Println(f.box.StrokeColor, f.box.FillColor)
	if f.input {
		return widget.NewSimpleRenderer(container.NewBorder(nil, nil, f.box, nil, editor))
	} else {
		return widget.NewSimpleRenderer(container.NewBorder(nil, nil, nil, f.box, editor))
	}
}

func (f *FloatParameter) MinSize() fyne.Size { return fyne.NewSize(50, 30) }

func (f *FloatParameter) Info() (string, any) {
	return f.Name, f.Value
}

func (f *FloatParameter) IsInput() bool { return f.input }

func (f *FloatParameter) AsFormItem() *widget.FormItem {
	return widget.NewFormItem(f.Name, f)
}

// type zoid widget.FormItem

func FloatParameterFromFormItem(item *widget.FormItem) *FloatParameter {
	return item.Widget.(*FloatParameter)
}
