package listablethumbnail

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
)

type Thumbnail struct {
	widget.BaseWidget
	Image   image.Image
	Caption string
	URI     fyne.URI
	label   *widget.Label
	canvas  *canvas.Image
	w, h    int
	pixels  string
}

func NewThumbNail(uri fyne.URI, w, h int) (*Thumbnail, error) {
	t := Thumbnail{URI: uri}
	im, err := imaging.Open(uri.Path(), imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	// t.Image=imaging.Thumbnail(im,w,h,imaging.Gaussian)
	t.Image = imaging.Fit(im, w, h, imaging.Gaussian)
	

	t.Caption = uri.Name()
	t.label = widget.NewLabel(t.Caption)
	t.label.Alignment = fyne.TextAlignCenter
	t.label.Truncation = fyne.TextTruncateEllipsis
	t.canvas = canvas.NewImageFromImage(t.Image)
	t.canvas.SetMinSize(fyne.NewSize(300, 200))
	t.canvas.FillMode = canvas.ImageFillContain
	t.w = im.Bounds().Dx()
	t.h = im.Bounds().Dy()
	pxcount := t.w * t.h
	px := float64(pxcount) / 1000000
	t.pixels = fmt.Sprintf("%d x %d - %.3f Mpixel", t.w, t.h, px)

	t.ExtendBaseWidget(&t)
	return &t, nil
}

func (t *Thumbnail) CreateRenderer() fyne.WidgetRenderer {
	blackrect:=canvas.NewRectangle(color.NRGBA{0,0,0,132})
	b := container.NewStack(t.canvas, container.NewBorder(nil, container.NewStack(blackrect, t.label), nil, nil, nil))
	return widget.NewSimpleRenderer(b)
}

func (t *Thumbnail) MouseIn(e *desktop.MouseEvent) {
	t.label.SetText(t.pixels)
}

func (t *Thumbnail) MouseMoved(e *desktop.MouseEvent) {
}

func (t *Thumbnail) MouseOut() {
	t.label.SetText(t.Caption)
}
