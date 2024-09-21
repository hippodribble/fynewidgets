package panzoomwidgetb

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"github.com/hippodribble/fynewidgets"
)

const DEVICE_SENSITIVITY int = 7

type PZWidget struct {
	widget.BaseWidget
	Transform   *PyramidTransform
	Canvas      *canvas.Image
	Pyramid     []*image.NRGBA
	Description string
	Channel     chan interface{}
}

func NewPZWidget(img image.Image, description string, sensitivity int, channel chan interface{}) (*PZWidget, error) {

	if sensitivity <= 0 {
		return nil, errors.New("sensitivity should be a small positive integer, eg 7 - it represents mouse scroll ticks required to double or halve the scale of the image")
	}

	var imin *image.NRGBA
	if im, ok := img.(*image.NRGBA); ok {
		imin = im
	} else {
		imin = imaging.Clone(im)
	}

	pyr, err := fynewidgets.MakePyramid(imin, 100)
	if err != nil {
		return &PZWidget{}, err
	}

	ci := canvas.NewImageFromImage(pyr[len(pyr)-1])
	ci.FillMode = canvas.ImageFillContain

	midimagepoint := image.Pt(pyr[0].Bounds().Dx()/2, pyr[0].Bounds().Dy()/2)
	t := PyramidTransform{
		ImageDatum:  &midimagepoint,
		Scale:       1,
		Ticks:       0,
		Sensitivity: sensitivity,
		Level:       0,
		Height:      len(pyr),
		Pyramid:     pyr,
	}
	widget := &PZWidget{
		Pyramid:     pyr,
		Canvas:      ci,
		Description: description,
		Transform:   &t,
		Channel:     channel,
	}
	widget.ExtendBaseWidget(widget)
	return widget, nil

}

func (p *PZWidget) CreateRenderer() fyne.WidgetRenderer {

	c := container.NewBorder(nil, widget.NewLabel(p.Description), nil, nil, p.Canvas)
	return widget.NewSimpleRenderer(c)

}

// redraws as long as the transform is appropriate. If not, it displays the full image instead.
func (p *PZWidget) Refresh() {

	log.Println("Refresh")
	if p.Transform.DeviceDatum == nil {
		p.Transform.FitToScreen(p.Canvas.Size())
		// return
	}
	imout := image.NewNRGBA(image.Rect(0, 0, int(p.Canvas.Size().Width), int(p.Canvas.Size().Height)))
	log.Println("New canvas:", imout.Bounds())

	w := p.Canvas.Size().Width                                            // REDRAWING THE OUTPUT
	h := p.Canvas.Size().Height                                           // dimensions of canvas
	tl := fyne.NewPos(0, 0)                                               // top left in device
	TL := p.Transform.ToImage(tl)                                         // top left on image
	br := fyne.NewPos(w, h)                                               // bottom right in device
	BR := p.Transform.ToImage(br)                                         // image coordinates of canvas corners
	rSource := image.Rectangle{*TL, *BR}                                  // rectangle covering image coordinates of canvas corners
	rDest := rSource.Sub(*TL)                                             // move the rectangle so its TL is at 0,0 - the TL requested will be at the NW corner
	nrgba := image.NewNRGBA(rDest)                                        // create something the right size, ie, as big as the pictures that will be on the screen
	draw.Draw(nrgba, rDest, p.Pyramid[p.Transform.Level], *TL, draw.Over) // draw to the new image from the origin to the size of the requested image, the base image beginning at the requested top left.
	p.Canvas.Image = nrgba                                                // put it in the canvas.Image
	p.Canvas.Refresh()                                                    // refresh the canvas.Image

}
func (p *PZWidget) tellApp(s string) {
	if p.Channel == nil {
		return
	}
	p.Channel <- s
}

// ensures the image fits on the device by creating the appropriate transform

func (p *PZWidget) Scrolled(e *fyne.ScrollEvent) {
	tick := e.Scrolled.DY
	p.Transform.Zoom(tick, e.Position)
	p.Refresh()
}

func (p *PZWidget) MouseIn(e *desktop.MouseEvent) {}

func (p *PZWidget) MouseMoved(e *desktop.MouseEvent) {

	point := p.Transform.ToImage(e.Position)

	p.tellApp(fmt.Sprintf("Mouse %.1f,%.1f, Image %d,%d, Layer %d", e.Position.X, e.Position.Y, point.X, point.Y, p.Transform.Level))

}

func (p *PZWidget) MouseOut() {}

func (p *PZWidget) MouseDown(e *desktop.MouseEvent) {}

func (p *PZWidget) MouseUp(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonSecondary {
		p.Transform.FitToScreen(p.Canvas.Size())
		p.Refresh()
	}
}
