package panzoomwidget

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"math"

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

// PanZoomWidget - allows images to be displayed, panned and zoomed in a window by using image/draw operations from the original image.
//
//	Large files are accommodated via Guassian pyramids
type PanZoomWidget struct {
	widget.BaseWidget
	canvasImage       canvas.Image
	Pyramid           []*image.NRGBA
	Requestfullscreen binding.Bool
	messages          chan string
	wheelclicks       int // tracks mouse wheel clicks for zoom factor, ensuring zoom can always return to exactly 1.0
	transform         *fynewidgets.PyramidTransform
	detailedImage     *image.NRGBA
	Zoom              *imagedetailwidget.ImageDetailWidget
	dragging          bool
}

// returns a new PanZoomWidget
//
//	minsize - the smallest level of the Gaussian pyramid in pixels
func NewPanZoomWidget(img *image.Image, minsize int, scrollSensitivity int) (*PanZoomWidget, error) {

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

	bestlevel, localscale, globalscale := fynewidgets.PyramidScale(1, scrollSensitivity, len(pyr))
	ci := canvas.NewImageFromImage(pyr[bestlevel])
	ci.FillMode = canvas.ImageFillContain
	ci.ScaleMode = canvas.ImageScaleFastest
	t := fynewidgets.PyramidTransform{
		GlobalScale:  float32(globalscale),
		LayerScale:   float32(localscale),
		ImageCentre:  &image.Point{pyr[bestlevel].Bounds().Dx() / 2, pyr[bestlevel].Bounds().Dy() / 2},
		NumLayers:    len(pyr),
		CurrentLayer: bestlevel,
		Sensitivity:  scrollSensitivity,
	}
	p := &PanZoomWidget{
		canvasImage: *ci,
		Pyramid:     pyr,
		transform:   &t,
		messages:    make(chan string),
	}
	p.Requestfullscreen = binding.NewBool()
	p.Requestfullscreen.Set(false)
	p.ExtendBaseWidget(p)
	return p, nil
}

// Using the appropriate level of the Gaussian pyramid, fills the screen by scaling the most appropriate level of the pyramid
func (p *PanZoomWidget) CreateRenderer() fyne.WidgetRenderer {
	b := container.NewBorder(nil, nil, nil, nil, &p.canvasImage)
	return widget.NewSimpleRenderer(b)
}

// send messages to the main app on this channel - useful for status updates, etc
//
//	alternative to using the Listener interface
func (p *PanZoomWidget) GetMessageChannel() chan string {
	return p.messages
}

func (p *PanZoomWidget) Cursor() desktop.Cursor {
	return desktop.CrosshairCursor
}

// converts mouse scroll wheel movement to a change both scale and pyramid level
func (p *PanZoomWidget) Scrolled(e *fyne.ScrollEvent) {
	if e.Scrolled.DY > 0 {
		p.wheelclicks++
	} else {
		p.wheelclicks--
	}
	sc := math.Pow(2, float64(p.wheelclicks)/float64(p.transform.Sensitivity))
	p.transform.Rescale(e.Position, sc)
	realsize := p.Pyramid[0].Bounds()
	wReal := realsize.Dx()
	hReal := realsize.Dy()
	dp, _ := p.transform.FromDevice(e.Position)
	X := float32(dp.X)
	Y := float32(dp.Y)
	X = min(max(X, 0), float32(wReal))
	Y = min(max(Y, 0), float32(hReal))
	index := 1 << p.transform.CurrentLayer
	X *= float32(index)
	Y *= float32(index)

	if p.messages != nil {
		p.messages <- fmt.Sprintf("%.1f, %.1f scale %.3f", X, Y, p.transform.GlobalScale)
	}

	p.Refresh()
}

// shows the full image, centred on screen
func (p *PanZoomWidget) FitToScreen() {
	if len(p.Pyramid) == 0 {
		return
	}

	global := min(p.Size().Width/float32(p.Pyramid[0].Bounds().Dx()), p.Size().Height/float32(p.Pyramid[0].Bounds().Dy()))
	clicks := fynewidgets.FloatScaleToTicks(global, p.transform.Sensitivity)
	newglobal := float64(fynewidgets.TickScaleToFloatScale(clicks, p.transform.Sensitivity))

	lvl, scale, _ := fynewidgets.PyramidScale(float32(newglobal), p.transform.Sensitivity, len(p.Pyramid))

	p.transform.CurrentLayer = lvl
	im := p.Pyramid[lvl]
	b := im.Bounds()
	p.transform.ImageCentre = &image.Point{X: b.Dx() / 2, Y: b.Dy() / 2}
	p.transform.NumLayers = len(p.Pyramid)
	p.transform.DeviceCentre = &fyne.Position{X: p.Size().Width / 2, Y: p.Size().Height / 2}
	p.transform.GlobalScale = float32(newglobal)
	p.transform.LayerScale = float32(scale)
	p.wheelclicks = clicks
	p.Refresh()

}
func (p *PanZoomWidget) Refresh() {

	if p.transform.DeviceCentre == nil {
		p.FitToScreen()
		return
	}

	TL, err := p.transform.FromDevice(fyne.NewPos(0, 0))
	if err != nil {
		return
	}
	BR, err := p.transform.FromDevice(fyne.NewPos(p.Size().Width, p.Size().Height))
	if err != nil {
		return
	}
	rSource := image.Rectangle{*TL, *BR}
	rDest := rSource.Sub(*TL)
	nrgba := image.NewNRGBA(rDest)
	draw.Draw(nrgba, rDest, p.Pyramid[p.transform.CurrentLayer], *TL, draw.Over)
	p.canvasImage.Image = nrgba
	p.BaseWidget.Refresh()

}

func (m *PanZoomWidget) DetailedImage(w, h int) *imagedetailwidget.ImageDetailWidget {
	if m.detailedImage == nil {
		m.detailedImage = image.NewNRGBA(image.Rect(0, 0, w, h))
		draw.Draw(m.detailedImage, image.Rect(0, 0, w, h), m.Pyramid[0], image.Point{0, 0}, draw.Over)
	}
	m.Zoom = imagedetailwidget.NewImageDetailWidget(m.detailedImage)
	return m.Zoom
}

func (p *PanZoomWidget) Resize(size fyne.Size) {
	p.BaseWidget.Resize(size)
	p.FitToScreen()
}

func (p *PanZoomWidget) MouseUp(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonSecondary {
		p.FitToScreen()
		return
	}
	if e.Button == desktop.MouseButtonPrimary {
		if !p.dragging {
			return
		}
		p.dragging = false

		wantFullScreen, _ := p.Requestfullscreen.Get()
		p.Requestfullscreen.Set(!wantFullScreen)
		if wantFullScreen {
			p.FitToScreen()
		}
		return
	}
}
func (p *PanZoomWidget) MouseDown(e *desktop.MouseEvent) { p.dragging = true }
func (p *PanZoomWidget) MouseIn(e *desktop.MouseEvent)   {}
func (p *PanZoomWidget) MouseMoved(e *desktop.MouseEvent) {

	if p.detailedImage == nil {
		p.DetailedImage(20, 20)
	}

	hw := p.detailedImage.Bounds().Dx() / 2
	hh := p.detailedImage.Bounds().Dy() / 2
	realsize := p.Pyramid[0].Bounds()
	wReal := realsize.Dx()
	hReal := realsize.Dy()

	dp, err := p.transform.FromDevice(e.Position)
	if err != nil {
		return
	}

	X := float32(dp.X)
	Y := float32(dp.Y)
	X = min(max(X, 0), float32(wReal))
	Y = min(max(Y, 0), float32(hReal))
	index := 1 << p.transform.CurrentLayer
	X *= float32(index)
	Y *= float32(index)
	if p.messages != nil {
		p.messages <- fmt.Sprintf("%.1f, %.1f scale %.3f", X, Y, p.transform.GlobalScale)
	}
	draw.Draw(p.detailedImage, p.detailedImage.Rect, p.Pyramid[0], image.Pt(int(X), int(Y)).Sub(image.Pt(hw, hh)), draw.Src)
	p.Zoom.Refresh()
}
func (p *PanZoomWidget) MouseOut() {}
