package adaptiveimagewidget

import (
	"errors"
	"image"
	"time"

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
//
// A goroutine monitors the container size periodically (at the update rate), to see if any change is required (and makes the change.)
//
//	updateRate : rate at which a check is made for updates to the container size (default 200 ms)
//	index      : the current layer of the pyramid which is being used in the image.Canvas
type AdaptiveImageWidget struct {
	widget.BaseWidget
	Image      canvas.Image
	Pyramid    []*image.Image
	index      int
	updateRate int `default:"200"`
	quality float32 `default:"1.0"`
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
	ci := *canvas.NewImageFromImage(*pyr[index])
	ci.FillMode = canvas.ImageFillContain
	ci.ScaleMode=canvas.ImageScaleFastest
	w := &AdaptiveImageWidget{
		Image:   ci,
		Pyramid: pyr,
		index:   index,
	}
	w.ExtendBaseWidget(w)

	var xratio, yratio, ratio float32
	var maxres float32 = .95
	var minres float32 = maxres / 2.05

	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(w.updateRate))

			size := w.Size()
			xratio = size.Width / float32((*w.Pyramid[w.index]).Bounds().Dx())
			yratio = size.Height / float32((*w.Pyramid[w.index]).Bounds().Dy())
			ratio = min(xratio, yratio)
			if ratio < maxres && ratio > minres {
				continue
			}
			if ratio > maxres {
				w.increaseResolution()
			} else {
				w.reduceResolution()
			}
		}
	}()
	return w, nil
}

func (item *AdaptiveImageWidget) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(&item.Image)
	ren := widget.NewSimpleRenderer(c)
	return ren
}

func (item *AdaptiveImageWidget) SetUpdateRate(milliseconds int) error {
	if milliseconds < 1 || milliseconds > 10000 {
		return errors.New("update rate should be between 1 and 10000")
	}
	item.updateRate = milliseconds
	return nil
}

func (item *AdaptiveImageWidget) GetUpdateRate() int {
	return item.updateRate
}


// increase the resolution
func (m *AdaptiveImageWidget) increaseResolution() error {
	if m.index == 0 {
		return errors.New("already at maximum resolution")
	}
	m.index -= 1
	m.updateResolution()
	return nil
}

// reduce the resolution
func (m *AdaptiveImageWidget) reduceResolution() error {
	if m.index == len(m.Pyramid)-1 {
		return errors.New("already at minimum resolution")
	}
	m.index += 1
	m.updateResolution()
	return nil
}

// implement the desired resolution
func (m *AdaptiveImageWidget) updateResolution() {
	ci := canvas.NewImageFromImage(*m.Pyramid[m.index])
	ci.FillMode = canvas.ImageFillContain
	ci.ScaleMode=canvas.ImageScaleFastest
	m.Image = *ci
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

	var layers []*image.Image = []*image.Image{img}

	for h > minsize {
		lastlayer := layers[len(layers)-1]

		b := (*lastlayer).Bounds()
		newlayer := image.Image(imaging.Resize(*lastlayer, b.Dx()/2, b.Dy()/2, imaging.Gaussian))
		layers = append(layers, &newlayer)
		h /= 2
	}
	return layers, nil
}
