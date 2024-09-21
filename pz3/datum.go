package pz3

import (
	"errors"
	"fmt"
	"image"
	"log"

	"fyne.io/fyne/v2"
	"github.com/hippodribble/fynewidgets"
)

// A general 2D Datum and Scale type, to convert between image and device coordinates
type Datum struct {
	ImagePoint  image.Point   // pixel location
	DevicePoint fyne.Position // device location of image pixel
	Scale       float32       // scale is DEVICE:IMAGE
	Ticks       int           // mouse or trackpad tracking, distance from zero (positive or negative)
	Sensitivity int           // scroll sensitivity, in ticks per octave
}

func (d *Datum) String() string {
	s := fmt.Sprintf("Datum:\nImage Point:  %d , %d\nDevice Point: %.1f , %.1f\nScale:        %.1f\n", d.ImagePoint.X, d.ImagePoint.Y, d.DevicePoint.X, d.DevicePoint.Y, d.Scale)
	return s
}

// returns a new Datum with negative unity scale (as a check for whether it has been modified)
func NewDatum() *Datum {

	return &Datum{Scale: -1}

}

// Set the parameters of the datum, viz the same location in image and device space, as well as the scale
func (d *Datum) Set(imagepoint image.Point, devicepoint fyne.Position, scale float32) {
	d.ImagePoint = imagepoint
	d.DevicePoint = devicepoint
	d.Scale = scale

}

// converts device to image coordinates
func (d Datum) ToImage(p fyne.Position) image.Point {

	q := p.Subtract(d.DevicePoint)
	return image.Pt(int(q.X/d.Scale)+d.ImagePoint.X, int(q.Y/d.Scale)+d.ImagePoint.Y)

}

// converts image to device coordinates
func (d Datum) ToDevice(P image.Point) fyne.Position {

	Q := P.Sub(d.ImagePoint)
	return fyne.NewPos(d.Scale*float32(Q.X)+d.DevicePoint.X, d.Scale*float32(Q.Y)+d.DevicePoint.Y)

}

// modifies the datum to represent the new device point and scale. Used when zooming about a mouse cursor, etc.
func (d *Datum) ScaleAt(devicepoint fyne.Position, scale float32) error {

	if scale == 0 {
		return errors.New("scale cannot be zero")
	}

	imagepoint := d.ToImage(devicepoint)
	d.Scale = scale
	d.DevicePoint = devicepoint
	d.ImagePoint = imagepoint

	return nil

}

// PyramidDatum extends Datum by adding functions to convert and scale, within the active Level
type PyramidDatum struct {
	Datum
	Level int
}

func (d *PyramidDatum) String() string {
	s := fmt.Sprintf("Datum:\nImage Point:  %d , %d\nDevice Point: %.1f , %.1f\nScale:        %.1f\nLevel:        %d", d.ImagePoint.X, d.ImagePoint.Y, d.DevicePoint.X, d.DevicePoint.Y, d.Scale, d.Level)

	if d.Level == -1 {
		return "No level defined"
	}
	return s
}

// creates a PyramidDatum containing an internal Datum with scale factor -1
func NewPyramidDatum() *PyramidDatum {

	p := PyramidDatum{Datum: *NewDatum()}
	return &p

}

func (pd *PyramidDatum) Set(imagepoint image.Point, devicepoint fyne.Position, scale float32, level int) {
	s:=fynewidgets.FloatScaleToTicks(scale, pd.Sensitivity)
	scale=fynewidgets.TickScaleToFloatScale(s, pd.Sensitivity)
	pd.Datum.Set(imagepoint, devicepoint, scale)
	pd.Level = level
}

// converts device to image coordinates, taking the level into account
func (pd PyramidDatum) ToImageForLevel(p fyne.Position) image.Point {

	log.Println("Converting device to point at",p,"level",pd.Level)

	power := 1 << pd.Level                                                                         // calculate scale multiplier from level
	q := p.Subtract(pd.DevicePoint)                                                                // move device point to origin
	s := fyne.NewPos(q.X/pd.Scale+float32(pd.ImagePoint.X), q.Y/pd.Scale+float32(pd.ImagePoint.Y)) // scale the image and move it to the image datum from the origin
	Q := image.Pt(int(s.X/float32(power)+.5), int(s.Y/float32(power)+.5))                          // scale the pixels by the power and convert to nearest integer pixel locations

	log.Println("Power is",power)
	log.Println("q is",q)
	log.Println("s is",s)
	log.Println("Q is",Q, "after powering")
	return Q

}

// converts image to device coordinates, taking the level into account
// chances are, this will never get used, so it should be treated as suspect :-)
func (pd PyramidDatum) ToDeviceForLevel(P image.Point) fyne.Position {

	power := 1 << pd.Level                                                                           // calculate scale multiplier from level
	s := image.Pt(P.X*power, P.Y*power)                                                              // descale using the power of the current level to level 0
	Q := s.Sub(pd.ImagePoint)                                                                        // remove the image datum
	q := fyne.NewPos(pd.Scale*float32(Q.X)+pd.DevicePoint.X, pd.Scale*float32(Q.Y)+pd.DevicePoint.Y) // scale the image and move it to the device datum from the origin

	return q

}
