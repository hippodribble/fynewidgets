package gps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Receiver struct{
	widget.BaseWidget
	c *fyne.Container
}

func NewReceiver(ID int)*Receiver{
	r:=&Receiver{}
	r.ExtendBaseWidget(r)
	return r
}

func (r *Receiver)CreateRenderer()fyne.WidgetRenderer{
	return widget.NewSimpleRenderer(r.c)
}