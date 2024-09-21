package adaptiveimagewidget

import (
	"errors"
	"fmt"
	"image"
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"github.com/hippodribble/fynewidgets"
	"github.com/hippodribble/fynewidgets/imagedetailwidget"
)

// AdaptiveImageWidget is a Fyne widget that contains a single canvas.Image
//
// Displaying larger images in a window requires that the graphics display subsystem filter and
// rescale the image for display. This can take significant time.
//
// The AdaptiveImageWidget seeks to solve this problem by adaptively setting the resolution of the image dynamically.
// Internally, it converts the image to a Gaussian pyramid, ie a set of increasingly lower-resolution images.
// One of these images is then displayed at all times.
// The displayed image is selected based on the size of the imagecanvas's container. Therefore:
//   - A small image in a large window is displayed at full resolution.
//   - A large image in a small window has its resolution reduced.
type AdaptiveImageWidget struct {
	widget.BaseWidget
	Image             canvas.Image
	Pyramid           []*image.NRGBA
	Zoom              *imagedetailwidget.ImageDetailWidget
	Requestfullscreen binding.Bool
	currentlayer      int
	detailedImage     *image.NRGBA
	dragging          bool
	messages          chan string
	desktop.Cursorable
}

// creates a new ImageWidget
//
//	img      the image to be displayed
//	minsize  the smallest dimension of the smallest image in the pyramid (it will be at least 32 pixels)
func NewImageWidget(img *image.Image, minsize int) (*AdaptiveImageWidget, error) {
	if minsize < 32 {
		minsize = 32
	}
	var nrgba *image.NRGBA
	a := *img
	if f, ok := a.(*image.NRGBA); ok {
		nrgba = f
	} else {
		nrgba = imaging.Clone(a)
	}
	pyr, err := fynewidgets.MakePyramid(nrgba, minsize)
	if err != nil {
		return nil, errors.Join(err)
	}
	index := len(pyr) - 1
	ci := canvas.NewImageFromImage(pyr[index])
	ci.FillMode = canvas.ImageFillContain
	ci.ScaleMode = canvas.ImageScaleFastest
	w := &AdaptiveImageWidget{
		Image:        *ci,
		Pyramid:      pyr,
		currentlayer: index,
	}
	w.Requestfullscreen = binding.NewBool()
	w.Requestfullscreen.Set(false)
	w.messages=make(chan string)
	w.ExtendBaseWidget(w)
	return w, nil
}

func (m *AdaptiveImageWidget) GetMessageChannel() chan string {
	return m.messages
}

func (m *AdaptiveImageWidget) Resize(size fyne.Size) {

	var xratio, yratio, ratio float32
	var maxres float32 = .95
	var minres float32 = maxres / 2.05 // >2 prevents resolution from oscillating

	m.BaseWidget.Resize(size) // needs to be called before changing the widget's image, otherwise the image gets overwritten by the resize

	xratio = size.Width / float32((*m.Pyramid[m.currentlayer]).Bounds().Dx())
	yratio = size.Height / float32((*m.Pyramid[m.currentlayer]).Bounds().Dy())
	ratio = min(xratio, yratio)
	if ratio < maxres && ratio > minres {
		return
	}

	if ratio > maxres {
		m.increaseResolution()
	} else {
		m.reduceResolution()
	}

}

func (m *AdaptiveImageWidget) Refresh() {
	m.Resize(m.Size()) // without this, maximising the screen will fail to select the appropriate resolution from the pyramid
}

func (m *AdaptiveImageWidget) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(&m.Image)
	ren := widget.NewSimpleRenderer(c)
	return ren
}

// increase the resolution
func (m *AdaptiveImageWidget) increaseResolution() error {
	if m.currentlayer == 0 {
		return errors.New("already at maximum resolution")
	}
	m.currentlayer -= 1
	m.updateResolution()
	return nil
}

// reduce the resolution
func (m *AdaptiveImageWidget) reduceResolution() error {
	if m.currentlayer == len(m.Pyramid)-1 {
		return errors.New("already at minimum resolution")
	}
	m.currentlayer += 1
	m.updateResolution()
	return nil
}

// implement the desired resolution
func (m *AdaptiveImageWidget) updateResolution() {
	m.Image.Image = m.Pyramid[m.currentlayer]
	m.Image.Refresh()
	m.Refresh()
}

func (m *AdaptiveImageWidget) DetailedImage(w, h int) *imagedetailwidget.ImageDetailWidget {
	if m.detailedImage == nil {
		m.detailedImage = image.NewNRGBA(image.Rect(0, 0, w, h))
		draw.Draw(m.detailedImage, image.Rect(0, 0, w, h), m.Pyramid[0], image.Point{0, 0}, draw.Over)
		// log.Println("made image size ", m.detailedImage.Bounds())
	}
	m.Zoom = imagedetailwidget.NewImageDetailWidget(m.detailedImage)
	return m.Zoom
}

func (m *AdaptiveImageWidget) MouseDown(e *desktop.MouseEvent) {
	m.dragging = true
}
func (m *AdaptiveImageWidget) MouseUp(e *desktop.MouseEvent) {
	if !m.dragging {
		return
	}
	m.dragging = false
	z, _ := m.Requestfullscreen.Get()
	m.Requestfullscreen.Set(!z)
}

func (m *AdaptiveImageWidget) MouseIn(e *desktop.MouseEvent) {}

func (m *AdaptiveImageWidget) MouseMoved(e *desktop.MouseEvent) {
	if m.detailedImage == nil {
		return
	}

	hw := m.detailedImage.Bounds().Dx() / 2
	hh := m.detailedImage.Bounds().Dy() / 2

	x, y := e.Position.X, e.Position.Y
	realsize := m.Image.Image.Bounds()
	wReal := realsize.Dx()
	hReal := realsize.Dy()
	wCanvas := m.Image.Size().Width
	hCanvas := m.Image.Size().Height
	xscale := float32(wReal) / wCanvas
	yscale := float32(hReal) / hCanvas
	// scale is real over canvas, so generally positive for large images S= R/C  => R = SC
	scale := max(xscale, yscale)
	// get actual pixel locations from the mouse point by scaling etc
	X, Y := (x-wCanvas/2)*scale+float32(wReal)/2, (y-hCanvas/2)*scale+float32(hReal)/2
	X = min(max(X, 0), float32(wReal))
	Y = min(max(Y, 0), float32(hReal))
	index := 1 << m.currentlayer
	X *= float32(index)
	Y *= float32(index)
	m.messages<-fmt.Sprintf("%.1f, %.1f",X,Y)
	draw.Draw(m.detailedImage, m.detailedImage.Rect, m.Pyramid[0], image.Pt(int(X), int(Y)).Sub(image.Pt(hw, hh)), draw.Src)
	m.Zoom.Refresh()
}

// to satisfy the interface
func (m *AdaptiveImageWidget) MouseOut() {}

func (m *AdaptiveImageWidget) Cursor() desktop.Cursor {
	return &fynewidgets.BoxCursor{}
}