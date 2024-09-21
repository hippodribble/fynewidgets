package simpleimagewidget

import (
	"image"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type SimpleImageWidget struct {
	widget.BaseWidget
	canvas    *canvas.Image
	baseimage *image.Image
}

func NewSimpleImageWidget(img *image.Image) *SimpleImageWidget {
	w := new(SimpleImageWidget)
	w.canvas = canvas.NewImageFromImage(*img)
	w.baseimage = img
	w.canvas.FillMode = canvas.ImageFillContain
	// w.canvas.SetMinSize(fyne.NewSize(400,400))

	w.ExtendBaseWidget(w)
	return w
}

func (w *SimpleImageWidget) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(w.canvas)
	return widget.NewSimpleRenderer(c)
}
func (w *SimpleImageWidget) SetCanvas(c *canvas.Image) {
	w.canvas = c
}

func (w *SimpleImageWidget) Resize(size fyne.Size) {

	size1 := w.canvas.MinSize()
	w.BaseWidget.Resize(size)
	w.canvas.SetMinSize(size)
	size2 := w.canvas.MinSize()
	log.Println("Resizing", size1, size2)
	w.Refresh()
}
