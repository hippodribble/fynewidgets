package panzoomwidgetb

import (
	"errors"
	"fmt"
	"image"
	"log"
	"math"

	"fyne.io/fyne/v2"
	"github.com/hippodribble/fynewidgets"
)

// For an image represented by a Gaussian pyramid, datum transformation between image and device coordinates needs to take into account which layer of the pyramid is being used.
//
//Local datum and scale refer to the current layer, whereas image datum and scale refer to the full image (layer 0 of the pyramid)

// Scaling requests are generally received via a mouse scroll event. In order to be able to return to zoom level 1 with perfect precision, zoom is tracked as integer mouse wheel or trackpad clicks. These are converted to floating point scale using a sensitivity in clicks per octave, eg 7 clicks cause size to halve or double.
type PyramidTransform struct {
	DeviceDatum *fyne.Position // datum device coordinates
	ImageDatum  *image.Point   // global datum image coordinates, ie relative to layer 0
	Scale       float32        // scale is Device to Image scale factor - for large images, this is generally less than unity. It refers to level 0 of the pyramid
	Ticks       int            // number of clicks from scale 1.0 using 2-power scaling using zoom = 2^(clicks/sensitivity)
	Sensitivity int            // clicks per octave when zooming
	Level       int            // current pyramid level
	Height      int            // height of pyramid
	Pyramid     []*image.NRGBA // pyramid of images
}

func (t *PyramidTransform) String() string {
	return fmt.Sprintf("Scale %.2f on level %d, Image Datum: %d,%d, Device Datum: %.1f,%.1f, Scroll Sensitivity:%d Clicks:%d",
		t.Scale, t.Level, t.ImageDatum.X, t.ImageDatum.Y, t.DeviceDatum.X, t.DeviceDatum.Y, t.Sensitivity, t.Ticks)
}

// change the display scale for the transform (relative to layer 0). This may cause the pyramid level to change.
func (t *PyramidTransform) SetScale(globalscale float64) error {
	if globalscale <= 0 {
		return errors.New("scales must be positive")
	}
	newLevel := -int(math.Log2(globalscale))
	t.Level = min(max(newLevel, 0), t.Height)

	return nil
}

// converts device coordinates to image coordinates for the current layer
func (t *PyramidTransform) ToImage(devicepoint fyne.Position) *image.Point {
	
	if t.Scale == 0 ||t.DeviceDatum==nil{
		return &image.Point{0,0}
	}

	A := devicepoint.Subtract(t.DeviceDatum)
	A.X /= t.Scale
	A.Y /= t.Scale
	
	levelScalar := float32(math.Pow(2, float64(t.Level)))
	log.Println("SCALAR",levelScalar)
	
	B := image.Pt(int(A.X), int(A.Y))
	C := B.Add(*t.ImageDatum)
	return &C
}

// response to a zoom level tick (mouse scroll tick, trackpad drag, etc) - transform and possibly level are modified
func (t *PyramidTransform) Zoom(i float32, pos fyne.Position) {
	if i > 0 {
		t.Ticks++
	}
	if i < 0 {
		t.Ticks--
	}

	t.DeviceDatum = &pos
	newimagedatum:=t.ToImage(pos)
	t.ImageDatum=newimagedatum

	t.Scale = fynewidgets.TickScaleToFloatScale(t.Ticks, t.Sensitivity)
	LEVEL:=-int(math.Log2(float64(t.Scale)))
	t.Level=min(max(LEVEL,0),len(t.Pyramid)-1)
}

func (t *PyramidTransform) FitToScreen(p fyne.Size) {
	log.Println("fitting to screen")
	w := p.Width
	h := p.Height
	W := t.Pyramid[0].Bounds().Dx()
	H := t.Pyramid[0].Bounds().Dy()

	pos := fyne.NewPos(w/2, h/2)
	POS := image.Pt(W/2, H/2)

	t.DeviceDatum = &pos
	t.ImageDatum = &POS

	s := min(w/float32(W), h/float32(H))

	ticks := fynewidgets.FloatScaleToTicks(s, t.Sensitivity)
	s2 := fynewidgets.TickScaleToFloatScale(ticks, t.Sensitivity)
	l := -int(math.Log2(float64(s2)))

	t.Ticks = ticks
	t.Scale = s2
	t.Level = min(max(0, l), len(t.Pyramid)-1)
	log.Println("New transform level is", t.Level,"with scale", t.Scale, t.ImageDatum,*t.DeviceDatum)

}
