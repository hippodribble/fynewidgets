package pz3

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
)

type PZ struct {
	widget.BaseWidget
	pyramid     *Pyramid
	canvas      *canvas.Image
	description string
	channel     chan interface{}
}

// Loads an image lazily from a file, returning a component immediately. When the image is loaded and handled, it is inserted into the component. This is necessary, as large compressed images can take a long time to rasterise
//
//	uri      The location of the image
//	minsize  The minimum size of the image
func NewPZFromFile(uri fyne.URI, minsize image.Point) (*PZ, error) {

	widget := &PZ{description: uri.Name(), canvas: canvas.NewImageFromImage(fynewidgets.MakeFillerImage(200, 200))}
	widget.canvas.FillMode = canvas.ImageFillContain
	widget.tellApp("Opening the file")
	widget.tellApp(2.0)

	go func(ww *PZ) {
		img, err := fynewidgets.LoadNRGBA(uri) // load image as NRGBA, even if it's something else (especially JPEG)
		if err != nil {                        // if loading fails, replace the placeholder image with a red one
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			return
		}
		pyr, err := NewPyramid(img, minsize) // why would it not be possible to create a pyramid? We don't know yet
		ww.pyramid = pyr                     // store the pyramid
		if err != nil {
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			return
		}
		ww.canvas.Image = pyr.images[pyr.Datum.Level] // show the smallest image by default. Large images can take too long to show, and will most likely be replaced by lower reoslution versions anyway.

		ww.Fit()
		ww.tellApp(-2.0)
	}(widget)

	return widget, nil // return immediately to keep the UI snappy like a crocodile

}

func (p *PZ) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(nil, widget.NewLabel(p.description), nil, nil, p.canvas)

	return widget.NewSimpleRenderer(c)

}

// Allows a channel to be added that can report information back to the calling application (status, progress, etc)
func (p *PZ) SetChannel(c chan interface{}) {
	p.channel = c
}

// Fits image to device extent by scaling and translating
//   - the datum is adjusted
//   - the canvas is refreshed
//   - new datum is centre of both
//   - the widget is refreshed
func (p *PZ) Fit() {
	log.Println("Fit called")
	if p.pyramid == nil {
		log.Println("No pyramid")
		return
	}
	if p.pyramid.Datum == nil {
		log.Println("No datum")
		return
	}
	datum := p.pyramid.Datum
	datum.Sensitivity = 5

	size := p.canvas.Size()
	SIZE := p.pyramid.images[0].Bounds()
	scale := min(size.Width/float32(SIZE.Dx()), size.Height/float32(SIZE.Dy()))
	if math.Abs(float64(scale)-float64(p.pyramid.Datum.Scale)) < .01 {
		log.Println("No scale change. Returning!!", size, SIZE)
		return
	}

	ticks := fynewidgets.FloatScaleToTicks(scale, datum.Sensitivity)
	scale = fynewidgets.TickScaleToFloatScale(ticks, datum.Sensitivity)
	mid := fyne.NewPos(size.Width/2, size.Height/2)
	MID := image.Pt(SIZE.Dx()/2, SIZE.Dy()/2)
	level := -int(math.Log2(float64(scale)))
	level = min(max(level, 0), p.pyramid.Height()-1)

	var newdatum *PyramidDatum = NewPyramidDatum()
	newdatum.Sensitivity = datum.Sensitivity
	newdatum.Level = level
	newdatum.Scale = scale
	newdatum.DevicePoint = mid
	newdatum.ImagePoint = MID
	newdatum.Ticks = ticks

	p.pyramid.Datum = newdatum

	log.Println(p.pyramid.Datum, size, SIZE)
	log.Println(newdatum, size, SIZE)

	// p.canvas.Image = p.pyramid.images[level]
	p.Refresh()
}

func (p *PZ) Refresh() {
	log.Println("Refresh called")
	if p.pyramid == nil {
		return
	}
	if p.pyramid.Datum == nil {
		return
	}
	if p.pyramid.Datum.Scale == -1 {
		p.Fit()
	}

	w := p.canvas.Size().Width                          // REDRAWING THE OUTPUT
	h := p.canvas.Size().Height                         // dimensions of canvas
	tl := fyne.NewPos(0, 0)                             // top left in device
	TL := p.pyramid.Datum.ToImageForLevel(tl)           // top left on image
	br := fyne.NewPos(w, h)                             // bottom right in device
	BR := p.pyramid.Datum.ToImageForLevel(br)           // image coordinates of canvas corners
	rSource := image.Rectangle{TL, BR}                  // rectangle covering image coordinates of canvas corners
	rDest := rSource.Sub(TL)                            // move the rectangle so its TL is at 0,0 - the TL requested will be at the NW corner
	nrgba := image.NewNRGBA(rDest)                      // create something the right size, ie, as big as the pictures that will be on the screen
	imageSrc := p.pyramid.images[p.pyramid.Datum.Level] // take the pixels to be drawn on screen from the current pyramid level
	draw.Draw(nrgba, rDest, imageSrc, TL, draw.Over)    // draw to the new image from the origin to the size of the requested image, the base image beginning at the requested top left.
	p.canvas.Image = nrgba                              // put it in the canvas.Image
	newcanvassize := fyne.NewSize(float32(rSource.Dx()), float32(rSource.Dy()))     // canvas should be same size as the pixels it draws. FillMode will take care of the scaling
	// p.canvas.Resize(newcanvassize)                                                  // resize the canvas to fit the image
	p.canvas.SetMinSize(newcanvassize) // resize the canvas to fit the image
	p.canvas.Refresh() // refresh the canvas.Image
	log.Println(p.pyramid.Datum, p.canvas.Size())

}

func (p *PZ) Resize(size fyne.Size) {
	p.BaseWidget.Resize(size)
	// p.canvas.Resize(fyne.NewSize(float32(p.canvas.Image.Bounds().Dx()), float32(p.canvas.Image.Bounds().Dy())))
	p.Fit()
}

func (p *PZ) MouseIn(e *desktop.MouseEvent) {}
func (p *PZ) MouseOut()                     {}
func (p *PZ) MouseMoved(e *desktop.MouseEvent) {
	// p.tellApp(fmt.Sprintf("Mouse: %.1f,%.1f - Image: %v",e.Position.X,e.Position.Y,p.pyramid.Datum.ToImageForLevel(e.Position)))
}                                             // report mouse position to app
func (p *PZ) MouseDown(e *desktop.MouseEvent) {}
func (p *PZ) MouseUp(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonSecondary {
		p.Fit()
	}
}

func (p *PZ) tellApp(information interface{}) {
	if p.channel != nil {
		p.channel <- information
	}
}
