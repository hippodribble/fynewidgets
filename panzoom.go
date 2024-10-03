package fynewidgets

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	eventbus "github.com/dtomasi/go-event-bus/v3"
	"github.com/pkg/errors"
)

// A Widget to display a large image with pan and zoom capability
// - solves the problem of overloading the graphics card with large images, by pyramid decomposition
type PanZoomCanvas struct {
	widget.BaseWidget
	datum               *Datum             // datum handles pyramid and projection for display
	canvas              *canvas.Image      // the image is displayed here
	bus                 *eventbus.EventBus // to talk to the application's StatusProgress widget
	busy                bool               // avoids the whole double bounce thing
	mousedown           bool               // for detecting drag etc
	mousedownpoint      fyne.Position      // where the mouse was clicked
	mousedownimagepoint image.Point        // where the image was clicked
	pixelcount          int                // pixels on device (mainly for testing)
	// datumchannel        chan Datum         // when there is a change, this channel can be used to notify other components
	uri   fyne.URI // originating URI, if available
	text  string   // used for labels
	loupe *Loupe   // used for providing a loup image to an application
	// channel             chan interface{} // to talk to the application's StatusProgress widget

}

func NewPanZoomCanvasFromImage(img image.Image, minsize image.Point, bus *eventbus.EventBus, description string) (*PanZoomCanvas, error) {

	widget := &PanZoomCanvas{
		canvas: canvas.NewImageFromImage(img),
		bus:    bus,
		text:   description}
	widget.canvas.FillMode = canvas.ImageFillStretch
	widget.canvas.SetMinSize(fyne.NewSize(100, 100))

	d, err := NewDatum(img, minsize, 5)
	if err != nil {
		widget.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
		return nil, errors.Wrap(err, "creating datum")
	}

	widget.datum = d
	d.FitDevice(fyne.NewSize(widget.canvas.Size().Width, widget.canvas.Size().Height))
	// widget.DatumChanged()

	widget.Refresh()
	return widget, nil
}

// Loads an image lazily from a file, returning a component immediately. When the image is loaded and handled, it is inserted into the component. This is necessary, as large compressed images can take a long time to rasterise
//
//	uri      The location of the image
//	minsize  The minimum size of the image
func NewPanZoomCanvasFromFile(uri fyne.URI, minsize image.Point, bus *eventbus.EventBus) (*PanZoomCanvas, error) {

	widget := &PanZoomCanvas{
		canvas: canvas.NewImageFromImage(MakeUniformColourImage(color.Gray{Y: 32}, 200, 200)),
		uri:    uri,
		busy:   true,
		text:   uri.Name(),
		bus:    bus}
	widget.canvas.FillMode = canvas.ImageFillContain
	widget.canvas.SetMinSize(fyne.NewSize(float32(minsize.X), float32(minsize.Y)))
	widget.uri = uri

	go func(ww *PanZoomCanvas) {
		defer func() { ww.busy = false }()

		if uri == nil {
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			return
		}
		img, err := LoadNRGBA(uri) // load image as NRGBA, even if it's something else (especially JPEG)
		if err != nil {            // if loading fails, replace the placeholder image with a red one
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			return
		}
		d, err := NewDatum(img, minsize, 5)
		if err != nil {
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			return
		}
		ww.datum = d

		d.FitDevice(fyne.NewSize(ww.canvas.Size().Width, ww.canvas.Size().Height))

		// ww.DatumChanged()

		ww.canvas.Image, _, err = d.GetCurrentImage(ww.canvas.Size()) // show the smallest image by default. Large images can take too long to show, and will most likely be replaced by lower reoslution versions anyway.
		if err != nil {
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			return
		}
		ww.Refresh()
	}(widget)

	widget.ExtendBaseWidget(widget)
	return widget, nil // return immediately to keep the UI snappy like a crocodile
}

func (p *PanZoomCanvas) URI() fyne.URI { return p.uri }

func (p *PanZoomCanvas) CreateRenderer() fyne.WidgetRenderer {
	if p.uri != nil {

		label := widget.NewLabelWithStyle(p.text, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		label.Truncation = fyne.TextTruncateEllipsis
		c := container.NewBorder(nil, label, nil, nil, nil)
		c2 := container.NewStack(p.canvas, c)
		return widget.NewSimpleRenderer(c2)

	}
	c := container.NewBorder(nil, nil, nil, nil, p.canvas)
	return widget.NewSimpleRenderer(c)
}

// set this component to use an external loupe to show full detail around the cursor.
func (p *PanZoomCanvas) SetLoupe(loupe *Loupe) {
	p.loupe = loupe
}

// returns the current image portion displayed in the component, at full resolution
func (p *PanZoomCanvas) CurrentImage() (image.Image, error) {
	if p.canvas.Image != nil {
		return p.canvas.Image, nil
	}
	return nil, errors.New("no image exists")
}

func (p *PanZoomCanvas) CurrentCanvas() (*canvas.Image, error) {
	if p.canvas.Image != nil {
		return p.canvas, nil
	}
	return nil, errors.New("no image exists")
}

func (p *PanZoomCanvas) SetMinSize(size fyne.Size) {
	p.canvas.SetMinSize(size)
	p.canvas.Refresh()
}

func (p *PanZoomCanvas) MinSize() fyne.Size {
	return p.canvas.MinSize()
}

// func (p *PanZoomCanvas) SetDatumChannel(channel chan Datum) {
// 	p.datumchannel = channel
// }

// func (p *PanZoomCanvas) DatumChanged() {
// if p.datumchannel != nil {
// 	p.datumchannel <- *p.datum
// }
// }

func (p *PanZoomCanvas) SetDatum(datum Datum) {
	p.datum = &datum
	p.Refresh()
}

func (p *PanZoomCanvas) Datum() *Datum {
	return p.datum
}

func (p *PanZoomCanvas) SetLoupeAtPoint(point *image.Point) error {
	if p.loupe == nil {
		return errors.New("no loup to use")
	}
	if p.datum == nil {
		return errors.New("no datum to use")
	}
	if p.datum.Pyramid == nil {
		return errors.New("no pyramid to use")
	}
	P := image.Pt(p.loupe.dimensions.X/2, p.loupe.dimensions.Y/2)
	tl := point.Sub(P)
	br := point.Add(P)

	srect := image.Rectangle{tl, br}
	smallimage := image.NewNRGBA(srect.Sub(tl))
	draw.Draw(smallimage, srect.Sub(tl), p.datum.Pyramid.images[0], srect.Min, draw.Over)

	p.loupe.canvas.Image = smallimage
	p.loupe.Refresh()

	return nil
}

// When the window is resized, show the full image and broadcast this datum change
func (p *PanZoomCanvas) Resize(size fyne.Size) {
	p.BaseWidget.Resize(size)
	err := p.datum.FitDevice(p.canvas.Size())
	if err != nil {
		return
	}
	p.Refresh()
	p.bus.PublishAsync("datum:changed", p.datum)
}

func (p *PanZoomCanvas) Refresh() {
	p.BaseWidget.Refresh()
	if p.datum == nil {
		return
	}
	if p.datum.Scale < 0 {
		p.datum.FitDevice(p.canvas.Size())
		p.bus.PublishAsync("datum:changed", p.datum)
	}
	img, pixelscount, err := p.datum.GetCurrentImage(p.canvas.Size())
	if err != nil {
		return
	}
	p.pixelcount = pixelscount
	p.canvas.Image = img

	text := fmt.Sprintf("L: %d | Scale: %d%% | %.2f MPix", p.datum.Pyramid.level, int(p.datum.Scale*100), float32(p.pixelcount)/1000000.0)

	p.bus.Publish("text:status", text)

	p.canvas.Refresh()
}

func (p *PanZoomCanvas) MouseOut() {}

func (p *PanZoomCanvas) MouseMoved(e *desktop.MouseEvent) {

	if p.datum == nil {
		return
	}
	point, err := p.datum.TransformDeviceToFullImage(e.Position)
	if err != nil {
		return
	}

	TL, _ := p.datum.TransformDeviceToFullImage(fyne.NewPos(0, 0))
	BR, _ := p.datum.TransformDeviceToFullImage(fyne.NewPos(p.canvas.Size().Width, p.canvas.Size().Height))
	SIZE := BR.Sub(*TL)

	p.bus.PublishAsync("text:status", fmt.Sprintf("L: %d | Scale: %d%% | %.2f MPix | M: %.1f %.1f | W: %d %d | Full View: %d x %d ",
		p.datum.Pyramid.level, int(p.datum.Scale*100), float32(p.pixelcount)/1000000.0,
		e.Position.X, e.Position.Y, point.X, point.Y, SIZE.X, SIZE.Y))

	if p.mousedown {
		if p.busy {
			return
		}

		go func(*PanZoomCanvas) {
			p.busy = true
			defer func() {
				p.busy = false
			}()
			p.datum.ImageCoords = &p.mousedownimagepoint
			p.datum.DeviceCoords = &e.Position
			p.Refresh()
			p.bus.PublishAsync("datum:changed", p.datum)
		}(p)
	}
	p.SetLoupeAtPoint(point)

}

func (p *PanZoomCanvas) MouseIn(e *desktop.MouseEvent) {

}

func (p *PanZoomCanvas) MouseUp(e *desktop.MouseEvent) {
	p.mousedown = false
	if e.Button == desktop.MouseButtonSecondary {

		err := p.datum.FitDevice(p.canvas.Size())
		if err != nil {
			return
		}
		p.bus.PublishAsync("datum:changed", p.datum)
		p.Refresh()
	}
	p.Refresh()
}

func (p *PanZoomCanvas) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonPrimary {
		p.bus.Publish("text:status", "Mouse Down")
		p.mousedown = true
		pt, err := p.datum.TransformDeviceToFullImage(e.Position)
		if err != nil {
			return
		}
		p.mousedownpoint = e.Position
		p.mousedownimagepoint = *pt
	}
}

func (p *PanZoomCanvas) Scrolled(e *fyne.ScrollEvent) {
	if p.busy {
		return
	}
	go func(p *PanZoomCanvas) {
		p.busy = true
		defer func() {
			p.busy = false
		}()
		p.datum.ScaleByTick(e.Position, e.Scrolled.DY)
		p.Refresh()
		p.bus.PublishAsync("datum:changed", p.datum)
		// p.DatumChanged()

	}(p)
}

func (p *PanZoomCanvas) Cursor() desktop.Cursor {
	return desktop.CrosshairCursor
}

func (p *PanZoomCanvas) TypedRune(r rune) {
	switch r {
	case '2':
		if p.busy {
			return
		}
		p.busy = true
		go func(p *PanZoomCanvas) {
			defer func() { p.busy = false }()
			p.datum.ChangeScale(2.0)
			p.Refresh()

		}(p)

	case '1':
		if p.busy {
			return
		}
		p.busy = true
		go func(p *PanZoomCanvas) {
			defer func() { p.busy = false }()
			p.datum.ChangeScale(0.5)
			p.Refresh()

		}(p)

	}
}

func (p *PanZoomCanvas) TypedKey(event *fyne.KeyEvent) {
}
