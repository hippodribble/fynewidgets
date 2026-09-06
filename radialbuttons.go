package fynewidgets

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/colornames"
)

// RadialButtons are a set of concentric circles with arbitrary colours and a popup label
// - clicking on a ring will trigger the action associated with the smallest circle containing the click.
type RadialButtons struct {
	widget.BaseWidget
	shapes        []fyne.CanvasObject
	diameter      float32
	buttons       []*widget.Button
	outline, fill color.Color
}

func NewRadialButtons(buttons []*widget.Button, diameter float32) *RadialButtons {
	n := float32(len(buttons))
	outline := color.RGBA{128, 64, 0, 255}
	// R, G, B, A := orange.RGBA()
	// R, G, B, A = R/256, G/256, B/256, 16
	// fill := color.RGBA{R: uint8(R), G: uint8(G), B: uint8(B), A: uint8(A)}
	fill := theme.Color(theme.ColorNameBackground)
	shapes := []fyne.CanvasObject{}
	for i := range buttons {
		c := canvas.NewCircle(fill)
		c.StrokeColor = outline
		c.StrokeWidth = 2
		radius := (n - float32(i)) / n * diameter
		// fmt.Println(radius, diameter)
		c.Resize(fyne.NewSize(radius, radius))
		shapes = append(shapes, c)
	}
	r := &RadialButtons{diameter: diameter, buttons: buttons, shapes: shapes, outline: outline, fill: fill}

	r.ExtendBaseWidget(r)
	return r
}

type radialLayout struct {
	r *RadialButtons
}

func (l *radialLayout) MinSize(obs []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(l.r.diameter, l.r.diameter)
}

func (l radialLayout) Layout(obs []fyne.CanvasObject, sz fyne.Size) {
	n := float32(len(l.r.buttons))
	cx, cy := sz.Width/2, sz.Height/2
	diameter := l.r.diameter
	for i, o := range obs {
		// fmt.Println(i)
		if c, ok := o.(*canvas.Circle); ok {
			radius := (n - float32(i)) / n * diameter
			// fmt.Println(radius, diameter)
			c.Resize(fyne.NewSize(radius, radius))
			c.Move(fyne.NewPos(cx-radius/2, cy-radius/2))
			// fmt.Println(c.Size())
			c.Refresh()
		}
	}
}

func (r *RadialButtons) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.New(&radialLayout{r: r}, r.shapes...))
}

func (r *RadialButtons) MouseIn(evt *desktop.MouseEvent) {}
func (r *RadialButtons) MouseMoved(evt *desktop.MouseEvent) {
	ring := r.getRing(evt)
	if ring < 0 {
		for _, shp := range r.shapes {
			if k, ok := shp.(*canvas.Circle); ok {
				k.FillColor = r.fill
			}
		}
		r.Refresh()
		return
	}
	if ring >= len(r.buttons) {
		return
	}
	if ring >= len(r.buttons) {
		for i := range r.buttons {
			r.shapes[i].(*canvas.Circle).FillColor = r.fill
			r.Refresh()
		}
		return
	}
	for i := range r.buttons {
		if i == ring {
			r.shapes[i].(*canvas.Circle).FillColor = colornames.Orange
			r.Refresh()
		} else {
			r.shapes[i].(*canvas.Circle).FillColor = r.fill
			r.Refresh()
		}
	}
	

}
func (r *RadialButtons) MouseOut() {
	for _, shp := range r.shapes {
		if k, ok := shp.(*canvas.Circle); ok {
			k.FillColor = r.fill
		}
	}
	r.Refresh()
}

func (r *RadialButtons) MouseDown(evt *desktop.MouseEvent) {
	ring := r.getRing(evt)
	if ring < 0 {
		return
	}
	r.buttons[ring].OnTapped()
}

func (r *RadialButtons) MouseUp(evt *desktop.MouseEvent) {}

func (r *RadialButtons) getRing(evt *desktop.MouseEvent) int {
	cx, cy := r.Size().Width/2, r.Size().Height/2
	n := float32(len(r.buttons))
	x, y := evt.Position.X, evt.Position.Y
	dx, dy := x-cx, y-cy
	rad := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	ring := len(r.buttons) - 1 - int(2*rad/r.diameter*n)
	if ring >= len(r.buttons) {
		ring = -1
	}
	return ring
}
