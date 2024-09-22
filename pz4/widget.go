package pz4

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
	"github.com/pkg/errors"
)

// A Widget to display a large image with pan and zoom capability
// - solves the problem of overloading the graphics card with large images, by pyramid decomposition
type PZ4 struct {
	widget.BaseWidget
	datum               *Datum           // datum handles pyramid and projection for display
	canvas              *canvas.Image    // the miage is displayed here
	channel             chan interface{} // to talk to the application's StatusProgress widget
	busy                bool             // for future expansion
	mousedown           bool             // for detecting drag etc
	mousedownpoint      fyne.Position    // where the mouse was clicked
	mousedownimagepoint image.Point      // where the image was clicked
	pixelcount          int              // pixels on device (mainly for testing)
}

// Loads an image lazily from a file, returning a component immediately. When the image is loaded and handled, it is inserted into the component. This is necessary, as large compressed images can take a long time to rasterise
//
//	uri      The location of the image
//	minsize  The minimum size of the image
func NewPZFromFile(uri fyne.URI, minsize image.Point, info chan interface{}) (*PZ4, error) {

	widget := &PZ4{canvas: canvas.NewImageFromImage(fynewidgets.MakeUniformColourImage(color.Gray{Y: 32}, 200, 200)), busy: true}
	widget.channel = info
	widget.canvas.FillMode = canvas.ImageFillContain
	widget.tellApp(2.0)

	go func(ww *PZ4) {
		defer func() { ww.busy = false }()
		t0:=time.Now()
		img, err := fynewidgets.LoadNRGBA(uri) // load image as NRGBA, even if it's something else (especially JPEG)
		if err != nil {                        // if loading fails, replace the placeholder image with a red one
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			ww.tellApp(errors.Wrap(err, "when loading NRGBA image"))        // report the error
			return
		}
		d, err := NewDatum(img, minsize, 5)
		if err != nil {
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			ww.tellApp(errors.Wrap(err, "when making datum"))
			return
		}

		ww.datum = d
		d.FitDevice(fyne.NewSize(ww.canvas.Size().Width, ww.canvas.Size().Height))

		ww.canvas.Image, _, err = d.GetCurrentImage(ww.canvas.Size()) // show the smallest image by default. Large images can take too long to show, and will most likely be replaced by lower reoslution versions anyway.
		if err != nil {
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			ww.tellApp(errors.Wrap(err, "when getting current image"))
			return
		}
		fmt.Printf("Loaded image in %v ms\n", time.Since(t0).Milliseconds())
		ww.tellApp(-2.0)
		ww.Refresh()
	}(widget)

	widget.ExtendBaseWidget(widget)

	return widget, nil // return immediately to keep the UI snappy like a crocodile

}

func (p *PZ4) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(nil, nil, nil, nil, p.canvas)
	return widget.NewSimpleRenderer(c)
}

func (p *PZ4) tellApp(information interface{}) {
	if p.channel != nil {
		p.channel <- information
	}
}

func (p *PZ4) Resize(size fyne.Size) {
	p.BaseWidget.Resize(size)
	err := p.datum.FitDevice(p.canvas.Size())
	p.Refresh()
	if err != nil {
		return
	}
}

func (p *PZ4) Refresh() {
	p.BaseWidget.Refresh()
	if p.datum == nil {
		return
	}
	if p.datum.Scale < 0 {
		p.datum.FitDevice(p.canvas.Size())
	}

	// img, err := p.datum.GetCurrentImage(p.canvas.Size())
	img, pixelscount, err := p.datum.GetCurrentImage(p.canvas.Size())
	if err != nil {
	}
	p.pixelcount = pixelscount
	p.canvas.Image = img
	p.channel <- fmt.Sprintf("L: %d | Scale: %d%% | %.2f MPix", p.datum.Pyramid.level, int(p.datum.Scale*100), float32(p.pixelcount)/1000000.0)

	p.canvas.Refresh()

}

func (p *PZ4) MouseOut() {}

func (p *PZ4) MouseMoved(e *desktop.MouseEvent) {

	if p.datum == nil {
		return
	}
	point, err := p.datum.TransformDeviceToFullImage(e.Position)
	if err != nil {
		return
	}

	p.channel <- fmt.Sprintf("L: %d | Scale: %d%% | %.2f MPix | M: %.1f %.1f | W: %d %d", p.datum.Pyramid.level, int(p.datum.Scale*100), float32(p.pixelcount)/1000000.0, e.Position.X, e.Position.Y, point.X, point.Y)
	if p.mousedown {
		if p.busy {
			return
		}
		p.busy = true

		go func(*PZ4) {
			defer func() {
				p.busy = false
			}()
			p.datum.ImageCoords = &p.mousedownimagepoint
			p.datum.DeviceCoords = &e.Position
			p.Refresh()
		}(p)
	}
}

func (p *PZ4) MouseIn(e *desktop.MouseEvent) {

}

func (p *PZ4) MouseUp(e *desktop.MouseEvent) {
	p.mousedown = false
	if e.Button == desktop.MouseButtonSecondary {

		err := p.datum.FitDevice(p.canvas.Size())
		if err != nil {
		}
		p.Refresh()
	}
	p.Refresh()
}

func (p *PZ4) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonPrimary {
		p.channel <- "Mouse Down"
		p.mousedown = true
		pt, err := p.datum.TransformDeviceToFullImage(e.Position)
		if err != nil {
			return
		}
		p.mousedownpoint = e.Position
		p.mousedownimagepoint = *pt
	}
}

func (p *PZ4) Scrolled(e *fyne.ScrollEvent) {
	if p.busy {
		return
	}
	p.busy = true
	go func(p *PZ4) {
		defer func() { p.busy = false }()
		p.datum.ScaleByTick(e.Position, e.Scrolled.DY)
		p.Refresh()

	}(p)
}

func (p *PZ4) Cursor() desktop.Cursor {
	return desktop.CrosshairCursor
}

func (p *PZ4) TypedRune(r rune) {
	switch r {
	case '2':
		if p.busy {
			return
		}
		p.busy = true
		go func(p *PZ4) {
			defer func() { p.busy = false }()
			p.datum.ChangeScale(2.0)
			p.Refresh()

		}(p)

	case '1':
		if p.busy {
			return
		}
		p.busy = true
		go func(p *PZ4) {
			defer func() { p.busy = false }()
			p.datum.ChangeScale(0.5)
			p.Refresh()

		}(p)

	}
}

func (p *PZ4) TypedKey(event *fyne.KeyEvent) {
}
