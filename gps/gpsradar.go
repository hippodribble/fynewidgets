package gps

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type CircleWithAngle struct {
	*canvas.Circle
	angle float32
}

type GPSRadar struct {
	widget.BaseWidget
	sats       []*GSV
	circles    []*canvas.Circle
	radials    []*canvas.Line
	squares    []*canvas.Rectangle
	labels     []*canvas.Text
	dphi       float32
	daz        float32
	squaresize float32
	TTL        float64
}

func NewGPSRadar(sats []*GSV, ttl float64) *GPSRadar {

	r := &GPSRadar{sats: sats, TTL: ttl}
	r.makeCircles(10)
	r.makeRadials(10)
	r.makeSquares(15)
	r.makeLabels()
	r.ExtendBaseWidget(r)
	return r
}

func (r *GPSRadar) CreateRenderer() fyne.WidgetRenderer {
	c := container.New(RadarLayout{r})
	for _, circle := range r.circles {
		c.Add(circle)
	}
	for _, radial := range r.radials {
		c.Add(radial)
	}

	for _, s := range r.squares {
		c.Add(s)
	}

	for _, s := range r.labels {
		c.Add(s)
	}

	return widget.NewSimpleRenderer(c)
}

// make circles from 0 to 90 derees at intervals
func (r *GPSRadar) makeCircles(degrees int) {
	if degrees < 5 || degrees > 45 {
		degrees = 30
	}
	r.circles = []*canvas.Circle{}
	for i := 0; i < 90; i += degrees {
		circle := canvas.NewCircle(color.Transparent)
		circle.StrokeColor = color.White
		circle.StrokeWidth = 0.5
		r.circles = append(r.circles, circle)
	}
	r.dphi = float32(degrees)

}

func (r *GPSRadar) makeRadials(degrees int) {
	r.radials = []*canvas.Line{}
	r.daz = float32(degrees)
	var angle float32
	for angle = 0; angle < 180; angle += float32(degrees) {
		l := canvas.NewLine(color.White)
		l.StrokeWidth = 0.25
		r.radials = append(r.radials, l)
	}
}

func (r *GPSRadar) makeSquares(width float32) {
	r.squares = []*canvas.Rectangle{}
	for range r.sats {
		square := canvas.NewRectangle(color.RGBA{255, 255, 0, 255})
		square.StrokeColor = color.RGBA{0, 0, 255, 255}
		square.StrokeWidth = 1
		r.squares = append(r.squares, square)
	}
	r.squaresize = width
}

func (r *GPSRadar) makeLabels() {
	r.labels = []*canvas.Text{}
	for i := range r.sats {
		l := canvas.NewText(fmt.Sprintf("%02d", i), theme.Color(theme.ColorNameForeground))
		r.labels = append(r.labels, l)
	}
}

func (r *GPSRadar) SetSVs(svs []*GSV) {
	r.sats = svs
}

type RadarLayout struct {
	r *GPSRadar
}

func (r RadarLayout) Layout(os []fyne.CanvasObject, sz fyne.Size) {

	cx, cy := sz.Width/2, sz.Height/2
	radius := min(cx, cy)*.9

	circlecount := 0
	radialcount := 0
	squarecount := 0
	labelcount := 0
	for _, o := range os {
		if c, ok := o.(*canvas.Circle); ok {
			circlecount++
			angle := r.r.dphi * float32(circlecount)
			rr := float64(radius * angle / 90)
			x0, y0 := cx-float32(rr), cy-float32(rr)
			c.Move(fyne.NewPos(x0, y0))
			c.Resize(fyne.NewSize(float32(rr*2), float32(rr*2)))
		}
		if l, ok := o.(*canvas.Line); ok {
			angle := r.r.daz * float32(radialcount) / 180 * 3.14159
			dx := float32(math.Cos(float64(angle))) * radius
			dy := float32(math.Sin(float64(angle))) * radius
			l.Position1 = fyne.NewPos(cx-dx, cy-dy)
			l.Position2 = fyne.NewPos(cx+dx, cy+dy)
			radialcount++
		}
		if sq, ok := o.(*canvas.Rectangle); ok {

			age := time.Since(r.r.sats[squarecount].data.lastupdate).Seconds()
			if age > r.r.TTL {
				sq.Hide()
			} else {
				prn,err:=stringToInt(r.r.sats[squarecount].data.PRN)
				if err!=nil{continue}
				if prn>64{
					sq.FillColor=color.RGBA{255,128,0,255}
				}else if prn>32{
					sq.FillColor=color.RGBA{255,0,0,255}
				}else{
					sq.FillColor=color.RGBA{0,255,255,255}
				}
				sat := r.r.sats[squarecount]
				az := 90 - float64(sat.data.Azimuth)
				az *= 3.14159 / 180
				el := float64(sat.data.Elevation)
				rr := float64(radius) * (90 - el) / 90 // radius
				x := float32(rr * math.Cos(az))
				y := float32(rr * math.Sin(az))
				x = cx + x - r.r.squaresize/2
				y = cy - y - r.r.squaresize/2
				sq.Move(fyne.NewPos(x, y))
				sq.Resize(fyne.NewSize(r.r.squaresize, r.r.squaresize))
				sq.Show()
			}
			squarecount++
		}
		if l, ok := o.(*canvas.Text); ok {

			age := time.Since(r.r.sats[labelcount].data.lastupdate).Seconds()
			if age > r.r.TTL {
				l.Hide()
			} else {
				sat := r.r.sats[labelcount]
				az := 90 - float64(sat.data.Azimuth)
				az *= 3.14159 / 180
				el := float64(sat.data.Elevation)
				rr := float64(radius) * (90 - el) / 90 // radius
				x := float32(rr * math.Cos(az))
				y := float32(rr * math.Sin(az))
				x = cx + x - r.r.squaresize/2+r.r.squaresize
				y = cy - y - r.r.squaresize/2+r.r.squaresize
				l.Move(fyne.NewPos(x, y))
				l.Resize(fyne.NewSize(r.r.squaresize, r.r.squaresize))
				l.Show()
			}
			labelcount++
		}
	}
}

func (r RadarLayout) MinSize(os []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(300, 300)
}
