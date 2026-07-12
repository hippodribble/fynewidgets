package fynewidgets

import (
	"image/color"
	"math"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/pkg/errors"
)

// PolygonalButtons are a closely-packed group of buttons laid out in a polygon.
// - labels are best specified as "xx/blah blah" - only "xx" will be shown on screen, due to the limited space.
type PolygonalButtons struct {
	widget.BaseWidget
	shapes                  []fyne.CanvasObject
	diameter, innerdiameter int
	buttons                 []*widget.Button
	busy                    bool
}

func NewPolygonalButtons(buttons []*widget.Button, diameter, innerdiameter int) (*PolygonalButtons, error) {
	if len(buttons) == 0 {
		return nil, errors.New("no buttons provided")
	}

	poly := canvas.NewPolygon(uint(len(buttons)), orange)
	poly.CornerRadius = canvas.RadiusMaximum
	shapes := []fyne.CanvasObject{poly}
	for range len(buttons) {
		l := canvas.NewLine(theme.Color(theme.ColorNameBackground))
		l.StrokeWidth = 2
		shapes = append(shapes, l)
	}
	for i := range len(buttons) {
		t := buttons[i].Text
		var tag string
		spl := strings.Split(t, "/")
		if len(spl) > 1 {
			tag = spl[0]
		} else {
			tag = string(buttons[i].Text[:2])
		}
		text := canvas.NewText(tag, theme.Color(theme.ColorNameBackground))
		text.TextSize = float32(diameter)/6
		text.Alignment = fyne.TextAlignCenter
		shapes = append(shapes, text)
	}
	innercircle := canvas.NewCircle(theme.Color(theme.ColorNameBackground))
	innercircle.Resize(fyne.NewSize(float32(innerdiameter), float32(innerdiameter)))
	shapes = append(shapes, innercircle)
	pb := &PolygonalButtons{shapes: shapes, diameter: diameter, innerdiameter: innerdiameter, buttons: buttons}
	pb.ExtendBaseWidget(pb)
	return pb, nil
}

type PolyLayout struct {
	buttons *PolygonalButtons
}

func (p PolyLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(float32(p.buttons.diameter)+5, float32(p.buttons.diameter)+5)
}

func (p PolyLayout) Layout(objects []fyne.CanvasObject, sz fyne.Size) {
	nlines := 0
	nlabels := 0
	r := p.buttons.diameter / 2
	cx, cy := sz.Width/2, sz.Height/2
	for i, o := range objects {
		if q, ok := o.(*canvas.Polygon); ok {
			q.Move(fyne.NewPos(cx-float32(r), cy-float32(r)))
			q.Resize(fyne.NewSize(float32(2*r), float32(2*r)))
			// fmt.Println(q.Position(), q.Size())
		}
		if q, ok := o.(*canvas.Line); ok {
			angle := float64(i-1)/float64(len(p.buttons.buttons))*math.Pi*2 - math.Pi/2
			angle += math.Pi / float64(len(p.buttons.buttons))
			q.Position1 = fyne.NewPos(cx, cy)
			x1 := float64(cx) + math.Cos(angle)*float64(r)
			y1 := float64(cy) + math.Sin(angle)*float64(r)
			q.Position2 = fyne.NewPos(float32(x1), float32(y1))
			nlines++
		}
		if q, ok := o.(*canvas.Text); ok {
			angle := float64(nlabels)/float64(len(p.buttons.buttons))*math.Pi*2 - math.Pi/2
			dx := float64(r+18) / 1.66 * math.Cos(angle)
			dy := float64(r+18)/1.66*math.Sin(angle) - 18
			q.Move(fyne.NewPos(cx, cy).AddXY(float32(dx), float32(dy)))
			nlabels++
		}
		if q, ok := o.(*canvas.Circle); ok {
			q.Move(fyne.NewPos(cx, cy).SubtractXY(float32(p.buttons.innerdiameter)/2, float32(p.buttons.innerdiameter)/2))
			nlabels++
		}
	}
}

func (p *PolygonalButtons) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.New(PolyLayout{p}, p.shapes...))
}

// runs the function and changes the label colour to provide feedback
func (p *PolygonalButtons) MouseDown(e *desktop.MouseEvent) {
	if p.busy {
		return
	}

	sector := p.getSector(e)
	if sector < 0 {
		return
	}

	p.busy = true

	go func() {

		go func() {
			p.buttons[sector].OnTapped()
			p.busy = false
		}()

		nSectors := len(p.buttons)
		indexOfLabel := sector + 1 + nSectors

		for range 2 {
			fyne.DoAndWait(func() {
				p.shapes[indexOfLabel].(*canvas.Text).Color = color.RGBA{0, 0, 128, 255}
				p.shapes[indexOfLabel].(*canvas.Text).TextStyle.Bold = true
				p.shapes[indexOfLabel].Refresh()
			})

			time.Sleep(time.Millisecond * 50)

			fyne.DoAndWait(func() {
				p.shapes[indexOfLabel].(*canvas.Text).Color = theme.Color(theme.ColorNameBackground)
				p.shapes[indexOfLabel].(*canvas.Text).TextStyle.Bold = false
				p.shapes[sector+nSectors+1].Refresh()
			})

			time.Sleep(time.Millisecond * 50)
		}

	}()
}

func (p *PolygonalButtons) MouseUp(e *desktop.MouseEvent)    {}
func (p *PolygonalButtons) MouseIn(e *desktop.MouseEvent)    {}
func (p *PolygonalButtons) MouseMoved(e *desktop.MouseEvent) {}
func (p *PolygonalButtons) MouseOut()                        {}

// which sector of the polygon had the mouse click?
func (p *PolygonalButtons) getSector(e *desktop.MouseEvent) int {
	N := float64(len(p.buttons))
	cx, cy := p.Size().Width/2, p.Size().Height/2
	x, y := e.Position.X, e.Position.Y
	dx, dy := x-cx, y-cy
	r := math.Sqrt(float64(dx*dx + dy*dy))
	if r > float64(p.diameter/2) {
		return -1
	}
	f := r / float64(p.diameter)
	if f < .2 {
		return -1
	}

	angle := math.Atan2(float64(dy), float64(dx))
	angle += math.Pi / 2
	dphi := math.Pi / float64(len(p.buttons))
	angle = math.Mod(angle+dphi+math.Pi*2, 2*math.Pi)
	N *= angle / 2 / math.Pi
	sector := int(N)
	return sector
}
