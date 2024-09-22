package pz4

import (
	"fmt"
	"image"
	"image/draw"
	"math"

	"fyne.io/fyne/v2"
	"github.com/hippodribble/fynewidgets"
	"github.com/pkg/errors"
)

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

func (d *Datum) FitDevice(size fyne.Size) error {
	if d == nil {
		return errors.New("No datum yet - nil")
	}

	SIZE := d.Pyramid.images[0].Bounds()                                        // maximum extent of image
	scale := min(size.Width/float32(SIZE.Dx()), size.Height/float32(SIZE.Dy())) // scale is DEVICE:IMAGE, and smaller of the two in order to fit the screen
	ticks := fynewidgets.FloatScaleToTicks(scale, d.Sensitivity)                // convert scale to integer ticks
	scale = fynewidgets.TickScaleToFloatScale(ticks, d.Sensitivity)             // convert ticks back to scale, thus creating discrete levels of zoom that are repeatable
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
// - first decide on the layer, then get the sub-image based on a local projection
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
	rSource := image.Rectangle{*TL, *BR}              // rectangle covering image coordinates of canvas corners
	rDest := rSource.Sub(*TL)                         // move the rectangle so its TL is at 0,0 - the TL requested will be at the NW corner
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

	ticks := fynewidgets.FloatScaleToTicks(scalerequested, d.Sensitivity) // convert to discrete ticks
	newscale := fynewidgets.TickScaleToFloatScale(ticks, d.Sensitivity)   // convert back to float

	d.DeviceCoords = &p                         // update device coordinates
	d.ImageCoords = ip                          // update image coordinates
	d.Scale = newscale                          // update scale
	d.Ticks = ticks                             // update ticks
	d.Pyramid.level = d.levelForScale(newscale) // update level

	return nil
}

func (d *Datum) ScaleByTick(p fyne.Position, delta float32) error {
	dir := 1
	if delta < 0 {
		dir = -1
	}
	newticks := d.Ticks + dir
	newscale := fynewidgets.TickScaleToFloatScale(newticks, d.Sensitivity) // convert modified ticks to float
	return d.ChangeProjection(p, newscale)
}

// multiply current scale by the requested factor (good for doubling size, etc)
func (d *Datum) ChangeScale(factor float32) error {

	scalerequested := d.Scale * factor
	ticks := fynewidgets.FloatScaleToTicks(scalerequested, d.Sensitivity) // convert to discrete ticks
	newscale := fynewidgets.TickScaleToFloatScale(ticks, d.Sensitivity)   // convert back to float

	d.Scale = newscale                         // update scale
	d.Ticks = ticks                            // update ticks
	d.Pyramid.level = d.levelForScale(d.Scale) // update level
	return nil
}

func (d *Datum) levelForScale(scale float32) int {
	level := -int(math.Log2(float64(scale))+.31)
	return min(max(level, 0), d.Pyramid.Height()-1) // constrain level to what is available in the pyramid
}
