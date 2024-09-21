package smallwidget

import (
	"image"
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"github.com/hippodribble/fynewidgets"
)

// A SmallWidget displays a simple image and allows mouse pan and zoom
//
//	This becomes slow with very large images when zoomed so that the entire image is visible, because all image pixels will be pushed to the graphics card.
//	The alternative to this is to use image pyramids, for which the PZWidget is available.
type SmallWidget struct {
	widget.BaseWidget
	canvas    *canvas.Image
	label     *widget.Label
	baseimage image.NRGBA
	transform fynewidgets.Transform
}

func (s *SmallWidget) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(nil, s.label, nil, nil, container.NewStack(s.canvas))
	return widget.NewSimpleRenderer(c)
}

func NewSmallWidget(img *image.Image, description string) *SmallWidget {

	nrgba := *imaging.Clone(*img)
	s := &SmallWidget{label: widget.NewLabel(description)}
	s.canvas = canvas.NewImageFromImage(&nrgba)
	s.canvas.FillMode = canvas.ImageFillContain
	s.baseimage = nrgba
	s.transform = fynewidgets.Transform{}
	s.transform.Datum = &fynewidgets.Datum2D{}
	s.ExtendBaseWidget(s)
	return s
}

// refresh is accomplished by drawing only part of the image to the canvas, to save time and pixels at the graphics card
func (s *SmallWidget) Refresh() {

	if s.transform.Datum.Scale == 0 {
		s.Fit()
	}
	w := s.canvas.Size().Width                           // REDRAWING THE OUTPUT
	h := s.canvas.Size().Height                          // dimensions of canvas
	tl := fyne.NewPos(0, 0)                              // top left in device
	TL := s.transform.Datum.ToImage(&tl)                 // top left on image
	br := fyne.NewPos(w, h)                              // bottom right in device
	BR := s.transform.Datum.ToImage(&br)                 // image coordinates of canvas corners
	rSource := image.Rectangle{TL, BR}                   // rectangle covering image coordinates of canvas corners
	rDest := rSource.Sub(TL)                             // move the rectangle so its TL is at 0,0 - the TL requested will be at the NW corner
	nrgba := image.NewNRGBA(rDest)                       // create something the right size, ie, as big as the pictures that will be on the screen
	draw.Draw(nrgba, rDest, &s.baseimage, TL, draw.Over) // draw to the new image from the origin to the size of the requested image, the base image beginning at the requested top left.
	s.canvas.Image = nrgba                               // put it in the canvas.Image
	s.canvas.Refresh()                                   // refresh the canvas.Image

}

// Fits the current image to the device display by computing an appropriate datum
//	An integer number of mousewheel or trackpad ticks is used to set the scale, so the image may be slightly smaller than the full extent of device space available.
//	This is so that the zoom can always return to precisely 100%, which may not happen with floating-point zoom if the image is repeatedly scaled up and down.
func (s *SmallWidget) Fit() {

	sensitivity := 7
	w := s.canvas.Size().Width
	h := s.canvas.Size().Height
	W := s.baseimage.Bounds().Dx()
	H := s.baseimage.Bounds().Dy()
	scale := min(w/float32(W), h/float32(H))
	ticks := fynewidgets.FloatScaleToTicks(scale, sensitivity)
	s.transform.Datum.Scale = fynewidgets.TickScaleToFloatScale(ticks, sensitivity)
	devdat := fyne.NewPos(w/2, h/2)
	s.transform.Datum.DeviceDatum = devdat
	imdat := fyne.NewPos(float32(W)/2, float32(H)/2)
	s.transform.Datum.ImageDatum = image.Pt(int(imdat.X), int(imdat.Y))
	s.transform.Ticks = ticks
	s.transform.Sensitivity = sensitivity

}

func (s *SmallWidget) MouseUp(e *desktop.MouseEvent) {
	s.Fit()
	s.Refresh()

}
func (s *SmallWidget) MouseDown(e *desktop.MouseEvent) {}

func (s *SmallWidget) Scrolled(e *fyne.ScrollEvent) {
	s.transform.Zoom(e.Position, e.Scrolled.DY)
	s.Refresh()

}
