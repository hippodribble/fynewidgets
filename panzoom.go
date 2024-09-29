package fynewidgets

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
)

// A Widget to display a large image with pan and zoom capability
// - solves the problem of overloading the graphics card with large images, by pyramid decomposition
type PanZoomCanvas struct {
	widget.BaseWidget
	datum               *Datum           // datum handles pyramid and projection for display
	canvas              *canvas.Image    // the image is displayed here
	channel             chan interface{} // to talk to the application's StatusProgress widget
	busy                bool             // avoids the whole double bounce thing
	mousedown           bool             // for detecting drag etc
	mousedownpoint      fyne.Position    // where the mouse was clicked
	mousedownimagepoint image.Point      // where the image was clicked
	pixelcount          int              // pixels on device (mainly for testing)
	datumchannel        chan Datum       // when there is a change, this channel can be used to notify other components
	uri                 fyne.URI         // originating URI, if available
	text                string           // used for labels
	loupe               *Loupe           // used for providing a loup image to an application

}

func NewPanZoomCanvasFromImage(img image.Image, minsize image.Point, info chan interface{}, description string) (*PanZoomCanvas, error) {

	widget := &PanZoomCanvas{canvas: canvas.NewImageFromImage(img), channel: info, text: description}
	widget.canvas.FillMode = canvas.ImageFillStretch
	widget.canvas.SetMinSize(fyne.NewSize(100, 100))

	widget.tellApp(2.0)

	d, err := NewDatum(img, minsize, 5)
	if err != nil {
		widget.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
		return nil, errors.Wrap(err, "creating datum")
	}

	widget.datum = d
	d.FitDevice(fyne.NewSize(widget.canvas.Size().Width, widget.canvas.Size().Height))
	widget.DatumChanged()

	widget.tellApp(-2.0)
	widget.Refresh()
	return widget, nil
}

// Loads an image lazily from a file, returning a component immediately. When the image is loaded and handled, it is inserted into the component. This is necessary, as large compressed images can take a long time to rasterise
//
//	uri      The location of the image
//	minsize  The minimum size of the image
func NewPanZoomCanvasFromFile(uri fyne.URI, minsize image.Point, info chan interface{}) (*PanZoomCanvas, error) {

	widget := &PanZoomCanvas{canvas: canvas.NewImageFromImage(MakeUniformColourImage(color.Gray{Y: 32}, 200, 200)), busy: true, text: uri.Name()}
	widget.channel = info
	widget.canvas.FillMode = canvas.ImageFillContain
	widget.canvas.SetMinSize(fyne.NewSize(100, 100))
	widget.uri = uri
	widget.tellApp(2.0)

	go func(ww *PanZoomCanvas) {
		defer func() { ww.busy = false }()

		if uri == nil {
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			return
		}
		img, err := LoadNRGBA(uri) // load image as NRGBA, even if it's something else (especially JPEG)
		if err != nil {            // if loading fails, replace the placeholder image with a red one
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
		// ww.DatumChanged()

		ww.canvas.Image, _, err = d.GetCurrentImage(ww.canvas.Size()) // show the smallest image by default. Large images can take too long to show, and will most likely be replaced by lower reoslution versions anyway.
		if err != nil {
			ww.canvas.Image = image.NewUniform(color.NRGBA{255, 0, 0, 255}) // replace the placeholder image with a red one
			ww.tellApp(errors.Wrap(err, "when getting current image"))
			return
		}
		ww.tellApp(-2.0)
		ww.Refresh()
	}(widget)

	widget.ExtendBaseWidget(widget)
	return widget, nil // return immediately to keep the UI snappy like a crocodile

}

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

func (p *PanZoomCanvas) tellApp(information interface{}) {
	if p.channel != nil {
		p.channel <- information
	}
}

func (p *PanZoomCanvas) SetDatumChannel(channel chan Datum) {
	p.datumchannel = channel
}

func (p *PanZoomCanvas) DatumChanged() {
	if p.datumchannel != nil {
		p.datumchannel <- *p.datum
	}
}

func (p *PanZoomCanvas) SetDatum(datum Datum) {
	p.datum = &datum
	p.Refresh()
}

func (p *PanZoomCanvas) Datum() *Datum {
	return p.datum
}

// When the window is resized, show the full image and broadcast this datum change
func (p *PanZoomCanvas) Resize(size fyne.Size) {
	p.BaseWidget.Resize(size)
	err := p.datum.FitDevice(p.canvas.Size())
	if err != nil {
		return
	}
	p.Refresh()
	p.DatumChanged()
}

func (p *PanZoomCanvas) Refresh() {
	p.BaseWidget.Refresh()
	if p.datum == nil {
		return
	}
	if p.datum.Scale < 0 {
		p.datum.FitDevice(p.canvas.Size())
		p.DatumChanged()
	}
	img, pixelscount, err := p.datum.GetCurrentImage(p.canvas.Size())
	if err != nil {
		return
	}
	p.pixelcount = pixelscount
	p.canvas.Image = img
	p.channel <- fmt.Sprintf("L: %d | Scale: %d%% | %.2f MPix", p.datum.Pyramid.level, int(p.datum.Scale*100), float32(p.pixelcount)/1000000.0)

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

	p.channel <- fmt.Sprintf("L: %d | Scale: %d%% | %.2f MPix | M: %.1f %.1f | W: %d %d | Full View: %d x %d ",
		p.datum.Pyramid.level, int(p.datum.Scale*100), float32(p.pixelcount)/1000000.0,
		e.Position.X, e.Position.Y, point.X, point.Y, SIZE.X, SIZE.Y)

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
			p.DatumChanged()
		}(p)
	}
	p.SetLoupeAtPoint(point)

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

func (p *PanZoomCanvas) MouseIn(e *desktop.MouseEvent) {

}

func (p *PanZoomCanvas) MouseUp(e *desktop.MouseEvent) {
	p.mousedown = false
	if e.Button == desktop.MouseButtonSecondary {

		err := p.datum.FitDevice(p.canvas.Size())
		if err != nil {
			return
		}
		p.DatumChanged()
		p.Refresh()
	}
	p.Refresh()
}

func (p *PanZoomCanvas) MouseDown(e *desktop.MouseEvent) {
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
		p.DatumChanged()

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

// Pyramid decomposition of an image.Image
//   - Each level is half the width and height of the previous level
//   - A projection PyramidDatum stores datum and scale for the top level of the pyramid, as well as the active level of the pyramid
type Pyramid struct {
	images []*image.NRGBA
	level  int
}

func (p Pyramid) String() string {
	s := fmt.Sprintf("\n\n------------------------\nPyramid\n%d\tLevels:\n", p.Height())

	for i := 0; i < p.Height(); i++ {
		if p.images[i] == nil {
			s += fmt.Sprintf("Level %2d : nil\n", i)
		} else {
			w := p.images[i].Bounds().Dx()
			h := p.images[i].Bounds().Dy()
			s += fmt.Sprintf("Level %2d : %4d x %4d", i, w, h)
			if i == p.level {
				s += " (Active)"
			}
			s += "\n"
		}
	}

	return s

}

// creates a pyramid from an image,Image, with the smallest possible size given. Active level will be the last one (lowest resolution)
func NewPyramid(img image.Image, smallestsize image.Point) (*Pyramid, error) {
	newpyramid := Pyramid{}
	W := img.Bounds().Dx()
	H := img.Bounds().Dy()
	ww := smallestsize.X
	hh := smallestsize.Y
	newpyramid.images = make([]*image.NRGBA, 0)
	newpyramid.images = append(newpyramid.images, imaging.Clone(img))
	W, H = W/2, H/2 // next level down the pyramid

	for W > ww && H > hh { // add level and scale down by 2 in x,y, while remaining above the minimum required size
		newpyramid.images = append(newpyramid.images, imaging.Resize(newpyramid.images[len(newpyramid.images)-1], W, H, imaging.Gaussian))
		W, H = W/2, H/2
	}
	newpyramid.level = len(newpyramid.images) - 1 // for safety in case someone tries to display a massive image

	return &newpyramid, nil

}

// returns the current level of the pyramid to use
func (p *Pyramid) Level() int {

	return p.level
}

func (p *Pyramid) SetLevel(level int) {
	if level >= 0 && level < p.Height() {
		p.level = level
	}
}

// returns the number of levels in the pyramid, whcih is at least 1 if the normal constructor was used.
func (p Pyramid) Height() int {

	return len(p.images)

}

// image at level i of the pyramid. Level 0 corresponds to the full image, with level 1 at half width and height, etc.
func (p Pyramid) CurrentImage() (*image.NRGBA, error) {

	if len(p.images) == 0 {
		return nil, errors.New("no images in pyramid")
	}
	if p.level < 0 || p.level >= len(p.images) {
		return nil, errors.New("level out of range")
	}
	return p.images[p.level], nil

}

type Datum struct {
	ImageCoords  *image.Point   // pixel location
	DeviceCoords *fyne.Position // device location of image pixel
	Scale        float32        // scale is DEVICE:IMAGE
	Ticks        int            // mouse or trackpad tracking, distance from zero (positive or negative) - creates discrete, repeatable levels of scaling
	Sensitivity  int            // scroll sensitivity, in ticks per octave
	Pyramid      *Pyramid       // pyramid of images with an associated current level
}

func (d Datum) String() string {
	if d.Scale < 0 {
		return "no datum has been set yet"
	}

	a := fmt.Sprintf("Datum -----------------------\nDevice: %.1f,%.1f\nImage:  %d,%d\nScale:  %.2f (%d ticks @ %d ticks per octave)\n", d.DeviceCoords.X, d.DeviceCoords.Y, d.ImageCoords.X, d.ImageCoords.Y, d.Scale, d.Ticks, d.Sensitivity)
	a += d.Pyramid.String()
	return a

}

func NewDatum(img image.Image, smallestsize image.Point, scrollsensitivity int) (*Datum, error) {
	p, err := NewPyramid(img, smallestsize)
	if err != nil {
		return nil, errors.Wrap(err, "f: NewDatum")
	}
	return &Datum{Scale: -1, Pyramid: p, Sensitivity: scrollsensitivity}, nil

}

// FitDevice resizes the image to fit the device by changing its datum.
func (d *Datum) FitDevice(size fyne.Size) error {
	if d == nil {
		return errors.New("No datum yet - nil")
	}

	SIZE := d.Pyramid.images[0].Bounds()                                        // maximum extent of image
	scale := min(size.Width/float32(SIZE.Dx()), size.Height/float32(SIZE.Dy())) // scale is DEVICE:IMAGE, and smaller of the two in order to fit the screen
	ticks := FloatScaleToTicks(scale, d.Sensitivity)                            // convert scale to integer ticks
	scale = TickScaleToFloatScale(ticks, d.Sensitivity)                         // convert ticks back to scale, thus creating discrete levels of zoom that are repeatable
	mid := fyne.NewPos(size.Width/2, size.Height/2)                             // centre of device
	MID := image.Pt(SIZE.Dx()/2, SIZE.Dy()/2)                                   // centre of image

	d.Pyramid.level = d.levelForScale(scale)
	d.Scale = scale
	d.DeviceCoords = &mid
	d.ImageCoords = &MID
	d.Ticks = ticks

	return nil
}

func (d *Datum) TransformDeviceToImage(devicepoint fyne.Position) (*image.Point, error) {
	if d.Pyramid == nil {
		return nil, errors.New("f:ImagePoint - no pyramid")
	}
	if d.Scale < 0 {
		return nil, errors.New("f:ImagePoint - no scale")
	}
	power := float32(math.Pow(2, float64(d.Pyramid.level)))                     // calculate pyramid level scalar - each layer's dimensions are half that of the previous
	P := devicepoint.Subtract(d.DeviceCoords)                                   // shift device point to origin
	s := fyne.NewPos(P.X/d.Scale, P.Y/d.Scale)                                  // scale
	s = fyne.NewPos(s.X+float32(d.ImageCoords.X), s.Y+float32(d.ImageCoords.Y)) // translate origin to image point
	s = fyne.NewPos(s.X/power+.5, s.Y/power+.5)                                 // nearest integer pixel needs an offset before truncation
	Q := image.Pt(int(s.X), int(s.Y))                                           // truncate to integer

	return &Q, nil
}
func (d *Datum) TransformDeviceToFullImage(devicepoint fyne.Position) (*image.Point, error) {
	if d.Pyramid == nil {
		return nil, errors.New("f:ImagePoint - no pyramid")
	}
	if d.Scale < 0 {
		return nil, errors.New("f:ImagePoint - no scale")
	}
	P := devicepoint.Subtract(d.DeviceCoords)                                   // shift device point to origin
	s := fyne.NewPos(P.X/d.Scale, P.Y/d.Scale)                                  // scale
	s = fyne.NewPos(s.X+float32(d.ImageCoords.X), s.Y+float32(d.ImageCoords.Y)) // translate origin to image point
	s = fyne.NewPos(s.X+.5, s.Y+.5)                                             // nearest integer pixel needs an offset before truncation
	Q := image.Pt(int(s.X), int(s.Y))                                           // truncate to integer

	return &Q, nil
}

// gets the image to be displayed using this datum, from the pyramid
func (d *Datum) GetCurrentImage(size fyne.Size) (*image.NRGBA, int, error) {

	w := size.Width                         // REDRAWING THE OUTPUT
	h := size.Height                        // dimensions of canvas
	tl := fyne.NewPos(0, 0)                 // top left in device coordinates
	TL, err := d.TransformDeviceToImage(tl) // top left in image coordinates
	if err != nil {
		return nil, 0, errors.Wrap(err, "Getting sub-image at TL")
	}
	br := fyne.NewPos(w, h)                 // bottom right in device
	BR, err := d.TransformDeviceToImage(br) // bottom right in image coordinates
	if err != nil {
		return nil, 0, errors.Wrap(err, "Getting sub-image at BR")
	}
	rSource := image.Rectangle{*TL, *BR} // rectangle covering image coordinates of canvas corners
	rDest := rSource.Sub(*TL)            // move the rectangle so its TL is at 0,0 - the TL requested will be at the NW corner
	if rDest.Dx() > 10000 || rDest.Dy() > 10000 || rDest.Dx() <= 1 || rDest.Dy() <= 1 {
		return nil, 0, errors.New("image too big")
	}
	nrgba := image.NewNRGBA(rDest)                    // create something the right size, ie, as big as the pictures that will be on the screen
	imageSrc := d.Pyramid.images[d.Pyramid.level]     // take the pixels to be drawn on screen from the current pyramid level
	draw.Draw(nrgba, rDest, imageSrc, *TL, draw.Over) // draw to the new image from the origin to the size of the requested image, the base image beginning at the requested top left.

	return nrgba, rSource.Dx() * rSource.Dy(), nil
}

// change the projection in response to a change of datum or scale. Most often used in mouse-centred zoom, or in panning
func (d *Datum) ChangeProjection(p fyne.Position, scalerequested float32) error {
	ip, err := d.TransformDeviceToFullImage(p)
	if err != nil {
		return errors.Wrap(err, "TransformDeviceToFullImage, in Datum.ChangeProjection")
	}

	ticks := FloatScaleToTicks(scalerequested, d.Sensitivity) // convert to discrete ticks
	newscale := TickScaleToFloatScale(ticks, d.Sensitivity)   // convert back to float

	d.DeviceCoords = &p                         // update device coordinates
	d.ImageCoords = ip                          // update image coordinates
	d.Scale = newscale                          // update scale
	d.Ticks = ticks                             // update ticks
	d.Pyramid.level = d.levelForScale(newscale) // update level

	return nil
}

// changes scale in response to a discrete change request (generally a single click of a mouse wheel, etc)
func (d *Datum) ScaleByTick(p fyne.Position, delta float32) error {
	dir := 1
	if delta < 0 {
		dir = -1
	}
	newticks := d.Ticks + dir
	newscale := TickScaleToFloatScale(newticks, d.Sensitivity) // convert modified ticks to float
	return d.ChangeProjection(p, newscale)
}

// multiply current scale by the requested factor (good for doubling size, etc)
func (d *Datum) ChangeScale(factor float32) error {

	scalerequested := d.Scale * factor
	ticks := FloatScaleToTicks(scalerequested, d.Sensitivity) // convert to discrete ticks
	newscale := TickScaleToFloatScale(ticks, d.Sensitivity)   // convert back to float

	d.Scale = newscale                         // update scale
	d.Ticks = ticks                            // update ticks
	d.Pyramid.level = d.levelForScale(d.Scale) // update level
	return nil
}

func (d *Datum) levelForScale(scale float32) int {
	level := -int(math.Log2(float64(scale)) + .31)
	return min(max(level, 0), d.Pyramid.Height()-1) // constrain level to what is available in the pyramid
}
