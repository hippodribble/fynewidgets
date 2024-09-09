package imagedetailwidget

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ImageDetailWidget struct {
	widget.BaseWidget
	image *image.NRGBA
}

// uses the highest-reolution image available to show a small area
func NewImageDetailWidget(img *image.NRGBA) *ImageDetailWidget {
	detail := new(ImageDetailWidget)
	detail.image = img
	detail.ExtendBaseWidget(detail)
	return detail
}

func (w *ImageDetailWidget) MinSize() fyne.Size {
	if w.image == nil {
		return fyne.NewSize(100, 100)
	}
	return fyne.NewSize(float32(w.image.Bounds().Dx()), float32(w.image.Bounds().Dy()))
}

func (w *ImageDetailWidget) CreateRenderer() fyne.WidgetRenderer {
	ci := canvas.NewImageFromImage(w.image)
	ci.FillMode = canvas.ImageFillContain
	return widget.NewSimpleRenderer(container.NewBorder(widget.NewLabel("Loupe"), nil, nil, nil, ci))
}

func (w *ImageDetailWidget) Refresh() {
	if w.image == nil {
		return
	}
	w.BaseWidget.Refresh()
}
