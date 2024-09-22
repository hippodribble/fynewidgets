package thumbnail

import (
	"fmt"
	"image"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"github.com/hippodribble/fynewidgets"
)

type Thumbnail struct {
	widget.BaseWidget
	Image              image.Image
	Caption            string
	URI                fyne.URI
	label              *widget.Label
	canvas             *canvas.Image
	w, h               int
	pixels             string
	downPoint, upPoint fyne.Position
	selected           binding.Bool
	commsChannel       chan interface{}
	clickchannel       chan fyne.URI
}

// creates a Thumbnail image lazily, ie it returns immediately, but loads the image to the thumbnail in another goroutine
func NewThumbNail(uri fyne.URI, w, h int, thresholdMegapixels float64, maxlabellength int) (*Thumbnail, error) {
	t := Thumbnail{URI: uri}
	t.Caption = uri.Name()
	if len(uri.Name()) > maxlabellength {
		t.Caption = fynewidgets.ShortenName(t.Caption, maxlabellength)
	}
	t.label = widget.NewLabel(t.Caption)
	t.label.Alignment = fyne.TextAlignCenter
	t.label.TextStyle.Italic = true
	// t.label.Truncation = fyne.TextTruncateEllipsis
	t.canvas = &canvas.Image{}
	t.canvas.SetMinSize(fyne.NewSize(float32(w), float32(h)))
	t.canvas.FillMode = canvas.ImageFillContain

	go func(t *Thumbnail) {
		r, err := os.Open(uri.Path())
		if err != nil {
			return
		}
		defer r.Close()
		config, _, err := image.DecodeConfig(r)
		if err != nil {
			return
		}

		t.w = config.Width
		t.h = config.Height
		pxcount := t.w * t.h
		px := float64(pxcount) / 1000000
		t.pixels = fmt.Sprintf("%d x %d - %.3f Mpixel", t.w, t.h, px)
		if px > thresholdMegapixels {
			t.Image = fynewidgets.MakeFillerImage(w, h)
			t.canvas.Image = t.Image
			return
		}
		r.Close()
		im, err := imaging.Open(uri.Path(), imaging.AutoOrientation(true))
		if err != nil {
			return
		}

		t.Image = imaging.Fit(im, w*2, h*2, imaging.Gaussian)
		t.canvas.Image = t.Image
		t.canvas.FillMode = canvas.ImageFillOriginal

	}(&t)
	t.selected = binding.NewBool()
	t.ExtendBaseWidget(&t)
	return &t, nil
}

func (t *Thumbnail) CreateRenderer() fyne.WidgetRenderer {
	b := container.NewBorder(nil, widget.NewCheckWithData(t.Caption, t.selected), nil, nil, t.canvas)
	return widget.NewSimpleRenderer(b)
}

func (t *Thumbnail) CommChannel() chan interface{}           { return t.commsChannel }
func (t *Thumbnail) SetCommChannel(channel chan interface{}) { t.commsChannel = channel }

func (t *Thumbnail) ClickChannel() chan fyne.URI           { return t.ClickChannel() }
func (t *Thumbnail) SetClickChannel(channel chan fyne.URI) { t.clickchannel = channel }

func (t *Thumbnail) MouseDown(e *desktop.MouseEvent) { t.downPoint = e.Position }
func (t *Thumbnail) MouseUp(e *desktop.MouseEvent) {
	t.upPoint = e.Position
	if t.clickchannel != nil {
		t.clickchannel <- t.URI
	}
}

func (t *Thumbnail) IsSelected() bool {
	b, err := t.selected.Get()
	if err != nil {
		return false
	}
	return b
}

func (t *Thumbnail) SetSelected(b bool) {
	t.selected.Set(b)
}
