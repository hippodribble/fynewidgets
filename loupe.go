package fynewidgets

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type Loupe struct {
	widget.BaseWidget
	image      *image.NRGBA
	Updated    binding.Bool
	dimensions image.Point
	canvas     *canvas.Image
	label      *widget.Label
	zoom       int
}

// uses the highest-reolution image available to show a small area
func NewLoupe(size image.Point, zoom int) *Loupe {

	w := new(Loupe)
	w.dimensions = size
	w.Updated = binding.NewBool()
	w.ExtendBaseWidget(w)
	w.image = image.NewNRGBA(image.Rect(0, 0, 50, 50))
	w.canvas = canvas.NewImageFromImage(w.image)
	w.canvas.FillMode = canvas.ImageFillContain
	if zoom >= 1 && zoom < 4 {
		w.zoom = zoom
	} else {
		w.zoom = 1
	}
	w.canvas.SetMinSize(fyne.NewSize(float32(size.X)*float32(w.zoom), float32(size.Y)*float32(w.zoom)))
	w.label = widget.NewLabel(fmt.Sprintf("Loupe ( %d X )", w.zoom))

	return w
}

func (w *Loupe) CreateRenderer() fyne.WidgetRenderer {

	return widget.NewSimpleRenderer(container.NewBorder(w.label, nil, nil, nil, w.canvas))
}

func (w *Loupe) Refresh() {
	if w.image == nil {
		return
	}
	w.BaseWidget.Refresh()
}