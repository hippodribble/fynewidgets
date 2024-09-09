package listablethumbnail

import (
	"fmt"
	"image"
	"log"

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

func NewThumbNail(uri fyne.URI, minsize int) (*Thumbnail, error) {
	t := Thumbnail{URI: uri}
	im, err := imaging.Open(uri.Path(), imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	// t.Image=imaging.Thumbnail(im,minsize,minsize,imaging.Gaussian)
	t.Image=imaging.Fill(im,minsize,minsize,imaging.Center,imaging.Gaussian)
	
	t.Caption = uri.Name()
	t.label = widget.NewLabel(uri.Name())
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
	b := container.NewBorder( t.label,nil, nil, nil, t.canvas)
	return widget.NewSimpleRenderer(b)
}

func (t *Thumbnail) MouseIn(e *desktop.MouseEvent) {
	// widget.NewPopUp(widget.NewLabel("TEST"), fyne.CurrentApp().Driver().CanvasForObject(t))
	log.Println("In Detail")
}

func (t *Thumbnail) MouseOver(e *desktop.MouseEvent) {
	log.Println("moved detail")
}

func (t *Thumbnail) MouseOut() {
	log.Println("Out detail")
}
