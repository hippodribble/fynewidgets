package fynewidgets

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
	eventbus "github.com/dtomasi/go-event-bus/v3"
	"github.com/pkg/errors"
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
	bus                *eventbus.EventBus
}

// creates a Thumbnail image lazily, ie it returns immediately, but loads the image to the thumbnail in another goroutine
func NewThumbNail(uri fyne.URI, w, h int, thresholdMegapixels float64, maxlabellength int, bus *eventbus.EventBus) (*Thumbnail, error) {
	t := Thumbnail{URI: uri, bus: bus}
	t.Caption = uri.Name()
	if len(uri.Name()) > maxlabellength {
		t.Caption = ShortenName(t.Caption, maxlabellength)
	}
	t.label = widget.NewLabel(t.Caption)
	t.label.Alignment = fyne.TextAlignCenter
	t.label.TextStyle.Italic = true
	// t.label.Truncation = fyne.TextTruncateEllipsis
	t.canvas = &canvas.Image{}
	sz:=fyne.NewSize(float32(w), float32(h))
	t.canvas.SetMinSize(sz)
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
			t.Image = MakeFillerImage(w, h)
			t.canvas.Image = t.Image
			t.canvas.Resize(fyne.NewSize(float32(w), float32(h)))
			return
		}
		r.Close()
		im, err := imaging.Open(uri.Path(), imaging.AutoOrientation(true))
		if err != nil {
			return
		}

		t.Image = imaging.Fit(im, w, h, imaging.Gaussian)
		t.canvas.Image = t.Image
		t.canvas.FillMode = canvas.ImageFillOriginal
		t.canvas.SetMinSize(sz)
		t.canvas.Resize(sz)

	}(&t)
	t.selected = binding.NewBool()
	t.ExtendBaseWidget(&t)
	return &t, nil
}

func (t *Thumbnail) CreateRenderer() fyne.WidgetRenderer {
	b := container.NewBorder(nil, widget.NewCheckWithData(t.Caption, t.selected), nil, nil, t.canvas)
	return widget.NewSimpleRenderer(b)
}



func (t *Thumbnail) MouseDown(e *desktop.MouseEvent) { t.downPoint = e.Position }
func (t *Thumbnail) MouseUp(e *desktop.MouseEvent) {
	t.upPoint = e.Position
	if t.bus != nil {
		t.bus.Publish("image:thumbnail",t.URI)
		
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

// A grid of thumbnail images from a slice of URIs
type ThumbnailGrid struct {
	widget.BaseWidget
	grid       *fyne.Container
	uris       []fyne.URI
	Thumbnails []*Thumbnail
	allnone    binding.Bool
}

func NewThumbnailGrid(uris []fyne.URI, w, h int, maxlabellength int, bus *eventbus.EventBus) (*ThumbnailGrid, error) {

	if uris == nil {
		return nil, errors.New("nil URI list")
	}

	if len(uris) == 0 {
		return nil, errors.New("emptyURI list")
	}

	g := &ThumbnailGrid{uris: uris, Thumbnails: make([]*Thumbnail, 0)}

	for _, uri := range uris {
		if !IsImage(uri) {
			continue
		}
		t, err := NewThumbNail(uri, w, h, 1000, maxlabellength, bus)
		if err != nil {
			continue
		}
		g.Thumbnails = append(g.Thumbnails, t)
	}

	ncols := 1
	// if len(g.Thumbnails) > 10 {
	// 	ncols = 2
	// }
	g.grid = container.NewGridWithColumns(ncols)

	for _, t := range g.Thumbnails {
		g.grid.Add(t)
	}

	g.allnone = binding.NewBool()

	g.ExtendBaseWidget(g)
	return g, nil
}

func (g *ThumbnailGrid) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(g.grid)
	return widget.NewSimpleRenderer(c)
}

func (g *ThumbnailGrid) SelectedFiles() []fyne.URI {
	selectedlist := make([]fyne.URI, 0)
	for _, t := range g.Thumbnails {
		if t.IsSelected() {
			selectedlist = append(selectedlist, t.URI)
		}
	}
	return selectedlist
}

func (g *ThumbnailGrid) ToggleSelection() {
	b, _ := g.allnone.Get()
	g.allnone.Set(!b)
	for _, w := range g.Thumbnails {
		w.SetSelected(!b)
	}
}
