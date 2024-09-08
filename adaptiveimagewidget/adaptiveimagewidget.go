package adaptiveimagewidget

import (
	"errors"
	"image"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
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
	Image        canvas.Image
	Pyramid      []*image.Image
	currentlayer int
}

// creates a new ImageWidget
//
//	img      the image to be displayed
//	minsize  the smallest dimension of the smallest image in the pyramid (it will be at least 32 pixels)
func NewImageWidget(img image.Image, minsize int) (*AdaptiveImageWidget, error) {
	if minsize < 32 {
		minsize = 32
	}
	pyr, err := makePyramid(&img, minsize)
	if err != nil {
		return nil, errors.Join(err)
	}
	index := len(pyr) - 1
	ci := canvas.NewImageFromImage(*pyr[index])
	ci.FillMode = canvas.ImageFillContain
	ci.ScaleMode = canvas.ImageScaleFastest
	w := &AdaptiveImageWidget{
		Image:        *ci,
		Pyramid:      pyr,
		currentlayer: index,
	}
	w.ExtendBaseWidget(w)
	return w, nil
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

func (m *AdaptiveImageWidget)Refresh(){
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
	m.Image.Image = *m.Pyramid[m.currentlayer]
	m.Image.Refresh()
	m.Refresh()
}

// makes the Gaussian pyramid of images - it uses github.com/disintegration/imaging for resizing
func makePyramid(img *image.Image, minsize int) ([]*image.Image, error) {
	h := (*img).Bounds().Dx()
	if (*img).Bounds().Dy() < h {
		h = (*img).Bounds().Dy()
	}
	if h < minsize {
		return nil, errors.New("image is already smaller than the required minimum")
	}

	var pyramid []*image.Image = []*image.Image{img}

	for h > minsize {
		lastlayer := pyramid[len(pyramid)-1]
		b := (*lastlayer).Bounds()
		newlayer := image.Image(imaging.Resize(*lastlayer, b.Dx()/2, b.Dy()/2, imaging.Gaussian))
		pyramid = append(pyramid, &newlayer)
		h /= 2
	}
	return pyramid, nil
}
