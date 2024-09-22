package fynewidgets

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
)

func MakePyramid(img *image.NRGBA, minsize int) ([]*image.NRGBA, error) {
	h := (*img).Bounds().Dx()
	if (*img).Bounds().Dy() < h {
		h = (*img).Bounds().Dy()
	}

	var pyramid []*image.NRGBA = []*image.NRGBA{img}

	for h > minsize {
		lastlayer := pyramid[len(pyramid)-1]
		b := (*lastlayer).Bounds()
		newlayer := imaging.Resize(lastlayer, b.Dx()/2, b.Dy()/2, imaging.Gaussian)
		pyramid = append(pyramid, newlayer)
		h /= 2
	}
	return pyramid, nil
}

type BoxCursor struct {
	desktop.Cursor
	w int
	h int
}

func NewBoxCursor(w, h int) *BoxCursor {
	b := new(BoxCursor)
	b.w = w
	b.h = h
	return b
}

func (b *BoxCursor) Image() (image.Image, int, int) {
	if b.w < 2 {
		b.w = 2
	}
	if b.h < 2 {
		b.h = 2
	}
	// log.Println(b.w, b.h)
	im := image.NewNRGBA(image.Rect(0, 0, b.w, b.h))
	fc := color.NRGBA{0, 0, 0, 64}
	for i := 0; i < b.w; i++ {
		for j := 0; j < b.h; j++ {
			im.Set(i, j, fc)
		}
	}
	// draw.Draw(im,im.Bounds(),image.NewUniform(fc),image.Pt(0,0),draw.Over)

	return im, b.w / 2, b.h / 2
}

// Very large images can be stored as Gaussian pyramids. Panning and zooming require that the number of pixels appearing on screen be limited, to prevent UI lag.
//
// Pyramid Scale determines the level and scale required to efficiently display an image at the requested absolute scale.
//
//	absoluteScale   requested ratio of screen scaling factor, ie screen pixels per image pixel
//	sensitivity     number of steps between scale doubling (logarithmic)
//	nlevels         height of the pyramid
//
// returns: pyramid level image to use, and scale to apply to it, in order to achieve the requested absolute scaling. The global scale is also returned (may be revised in tis function later)
func PyramidScale(absoluteScale float32, sensitivity, nlevels int) (int, float32, float32) {
	l := math.Log2(float64(absoluteScale))
	level := -int(math.Floor(l))
	level -= 1
	level = max(level, 0)
	level = min(level, nlevels-2)
	scalar := l + float64(level)
	scalar = math.Pow(2, scalar)
	return level, float32(scalar), absoluteScale
}

// A Transform maintains the constants required to scale between device coordinates and underlying image coordinates
type PyramidTransform struct {
	GlobalScale  float32 `default:"1.0"` // scale is Device/Image < 1 for large images
	LayerScale   float32 // scale is Device/Image < 1 for large images
	ImageCentre  *image.Point
	DeviceCentre *fyne.Position
	NumLayers    int
	CurrentLayer int
	Sensitivity  int // clicks per octave
	Clicks       int // current number of clicks applied
}

func (t *PyramidTransform) Rescale(centre fyne.Position, newScale float64) {

	ip, err := t.FromDevice(centre) // image coordinates
	if err != nil {
		return
	}
	t.DeviceCentre = &centre
	t.ImageCentre = ip
	fLevel := math.Log2(newScale)
	newlevel := -int(math.Floor(fLevel + .9))

	newlevel = max(newlevel, 0)
	newlevel = min(newlevel, t.NumLayers-1)

	newscalar := fLevel + float64(newlevel)
	newscalar = math.Pow(2, newscalar)
	d := t.CurrentLayer - newlevel
	t.LayerScale = float32(newscalar)
	t.GlobalScale = float32(newScale)
	// if pyramid level changes, change the image coordinate too
	if d != 0 {
		x := ip.X
		y := ip.Y
		if d > 0 {
			x = x << d
			y = y << d
		} else {
			x = x >> -d
			y = y >> -d
		}
		t.ImageCentre = &image.Point{x, y}
		t.CurrentLayer = newlevel
	}
}

func (t *PyramidTransform) ToDevice(P image.Point) (*fyne.Position, error) {
	x := float32(P.X)
	y := float32(P.Y)
	x -= float32(t.ImageCentre.X)
	y -= float32(t.ImageCentre.Y)
	x *= t.LayerScale
	y *= t.LayerScale
	x += t.DeviceCentre.X
	y += t.DeviceCentre.Y
	p := fyne.NewPos(x, y)
	return &p, nil
}

func (t *PyramidTransform) FromDevice(M fyne.Position) (*image.Point, error) {
	if t.DeviceCentre == nil {
		return nil, errors.New("transform not ready")
	}
	x := M.X - t.DeviceCentre.X
	y := M.Y - t.DeviceCentre.Y
	x /= t.LayerScale
	y /= t.LayerScale
	x += float32(t.ImageCentre.X)
	y += float32(t.ImageCentre.Y)
	p := image.Pt(int(x+.5), int(y+.5))
	return &p, nil
}

func (t *PyramidTransform) String() string {
	return fmt.Sprintf("Transform2D: Scale=%.3g\tImage Point: %v\tDevice Point %v (Sensitivity=%d), current layer=%d", t.GlobalScale, t.ImageCentre, t.DeviceCentre, t.Sensitivity, t.CurrentLayer)
}

// converts a number of clicks at a given sensitivity (clicks per octave) to a float ratio
func TickScaleToFloatScale(ticks, sensitivity int) float32 {
	return float32(math.Pow(2, float64(ticks)/float64(sensitivity)))
}

// converts a float ratio to a number of clicks at a given sensitivity (clicks per octave)
func FloatScaleToTicks(scale float32, sensitivity int) int {
	m := 1
	ticks := int(0.5 + math.Log2(float64(scale))*float64(sensitivity))

	if scale < 1 {
		m = -1
		ticks = int(0.5 + math.Log2(float64(1/scale))*float64(sensitivity))
	}
	return ticks * m
}

func MakeFillerImage(w, h int) image.Image {
	return MakeUniformColourImage(color.NRGBA{128, 128, 0, 255}, w, h)
}

func MakeUniformColourImage(c color.Color, w, h int) image.Image {
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Pt(0, 0), draw.Over)
	return im
}

func ShortenName(longname string, length int) string {

	N := len(longname)

	if N <= length || length < 10 {
		// log.Println("RETURN",longname)
		return longname
	}
	ellipsis := ".."
	a := longname[:length/2] + ellipsis + longname[len(longname)-2-length/2:]
	// log.Println("RETURN",a,"for",longname)
	return a
}

// returns a number of rows and columns that will be as close as possible to square. layouts are bigger horizontally, as are most photos, so no adjustment is made
func BestRowsColumns(N int) (int, int) {
	bestRows, bestCols := 1, 1
	bestRatio := 400.0
	desiredRatio := 1.0
	var rows, cols int
	for rows = 1; rows < N; rows++ {
		cols = N / rows
		if cols*rows < N {
			cols += 1
		}
		ratio := math.Abs(float64(rows)/float64(cols) - desiredRatio)
		if ratio < bestRatio {
			bestRatio = ratio
			bestRows, bestCols = rows, cols
		}
	}
	// log.Printf("%d images, %d rows x %d columns, %.2f vs 0.75\n",N, bestRows,bestCols, bestRatio)
	return bestRows, bestCols
}

type Datum2D struct {
	ImageDatum  image.Point
	DeviceDatum fyne.Position
	Scale       float32
}

func (t Datum2D) ToImage(pos *fyne.Position) image.Point {
	a := pos.Subtract(t.DeviceDatum)
	b := fyne.Position{}
	b.X = a.X / float32(t.Scale)
	b.Y = a.Y / float32(t.Scale)
	C := b.AddXY(float32(t.ImageDatum.X), float32(t.ImageDatum.Y))
	return image.Pt(int(C.X), int(C.Y))
}

func (t Datum2D) ToDevice(pos *fyne.Position) image.Point {
	a := pos.SubtractXY(float32(t.ImageDatum.X), float32(t.ImageDatum.Y))
	b := fyne.Position{}
	b.X = a.X * float32(t.Scale)
	b.Y = a.Y * float32(t.Scale)
	C := b.Add(t.DeviceDatum)

	return image.Pt(int(C.X), int(C.Y))
}

// scale in a transform is DEVICE/IMAGE
type Transform struct {
	Datum              *Datum2D // define the datum
	Ticks, Sensitivity int      // mouse or trackpad tracking & sensitivity
}

// modifies the Transform's datum to take account of the scale and zoom centre.
func (t *Transform) Zoom(p fyne.Position, DY float32) {
	q := t.Datum.ToImage(&p) // a new point in image coordinates
	if DY > 0 {
		t.Ticks++
	}
	if DY < 0 {
		t.Ticks--
	}
	t.Datum.Scale = TickScaleToFloatScale(t.Ticks, t.Sensitivity)
	t.Datum.ImageDatum = q
	t.Datum.DeviceDatum = p
}

func LoadImage(uri fyne.URI) (*image.Image, error) {

	img, err := imaging.Open(uri.Path())
	if err != nil {
		return nil, errors.Wrap(err, "LoadImage")
	}

	return &img, nil

}

func LoadNRGBA(uri fyne.URI) (*image.NRGBA, error) {

	img, err := LoadImage(uri)
	if err != nil {
		return nil, errors.Wrap(err, "LoadNRGBA")
	}
	return imaging.Clone(*img), nil

}

func IsImage(uri fyne.URI) bool {
	return uri.MimeType() == "image/jpeg" || uri.MimeType() == "image/png" || uri.MimeType() == "image/gif"
}
