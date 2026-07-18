package fynewidgets

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// a payload for a DialCircular
//   - represents continuous numeric data along with any text information to display
//   - use a channel of these to update a dial in real time
type CircularData struct {
	Value float64
	Text  string
}

// A DialInfinite is a continuous, circular dial good for showing things like time or compass direction
//   - does the whole modulo thing
type InfiniteDial struct {
	widget.BaseWidget
	indexmarks             []*canvas.Line
	textDisplay            *canvas.Text
	lowColor, accentColor  color.Color
	paused                 bool
	size                   float32
	updateChannel          chan CircularData
	major, minor           int
	majorStart, minorStart float32
	value, maxval          float64
	text                   string
}

// Returns a new DialInfinite
//   - minor - the number of ticks in a circle
//   - major - the number of minor ticks per major tick. Generally, minor/major should be an integer
//   - size - width/height of the square dial
//   - majorStart (0.3-0.99) - fraction of the dial radius at which major ticks start - 0.85 is good
//   - minorStart (0.5-0.99) - fraction of the dial radius at which minor ticks start - 0.90 is good
//   - lowColor - unemphasised colour
//   - accentColor - emphasised colour - colour of the tick that represents the current value of the dial
//   - maxval - the maximum value (eg 360) - modulus is applied to display out-of-range values
//   - text - inital text to display (warning - it can easily extend outside the dial and is not clipped)
//
// There is also a function to set an internal channel of CircularData so that data can be piped to the dial
func NewDialInfinite(minor, major int,
	size, majorStart, minorStart float32,
	lowColor, accentColor color.Color,
	maxval float64,
	text string,
) *InfiniteDial {

	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantLight {
		lowColor, accentColor = accentColor, lowColor
	}

	chapters := make([]*canvas.Line, minor)
	for i := range chapters {
		chapters[i] = canvas.NewLine(lowColor)
		chapters[i].StrokeWidth = 3
	}

	display := canvas.NewText(text, accentColor)
	display.SetMinSize(fyne.NewSize(size, size))
	display.TextSize = size / 8
	display.TextStyle.Bold = true
	display.Alignment = fyne.TextAlignCenter

	d := &InfiniteDial{
		textDisplay: display,
		indexmarks:  chapters,
		major:       major,
		majorStart:  majorStart,
		minor:       minor,
		minorStart:  minorStart,
		lowColor:    lowColor,
		accentColor: accentColor,
		size:        size,
		maxval:      maxval,
		value:       maxval / 2,
		text:        text,
	}
	d.ExtendBaseWidget(d)
	return d
}

func (d *InfiniteDial) CreateRenderer() fyne.WidgetRenderer {
	o := []fyne.CanvasObject{d.textDisplay}
	for _, chapter := range d.indexmarks {
		o = append(o, chapter)
	}
	c := container.New(InfiniteDialLayout{d}, o...)
	return widget.NewSimpleRenderer(c)
}

func (d *InfiniteDial) SetChannel(ch chan CircularData) {
	d.updateChannel = ch
	go func() {
		for update := range d.updateChannel {
			d.text = update.Text
			d.value = update.Value
			d.update()
		}
	}()
}

func (d *InfiniteDial) SetValue(v float64) {
	fmt.Println("SetValue")
	d.value = v
	d.update()
}

func (d *InfiniteDial) SetText(text string) {
	fmt.Println("SetText")
	d.text = text
	d.update()
}

func (d *InfiniteDial) SetCircularData(data CircularData) {
	d.text = data.Text
	d.value = data.Value
	d.update()
}

func (d *InfiniteDial) update() {
	x := math.Mod(d.value, d.maxval)
	if x < 0 {
		x += d.maxval
	}
	f := math.Mod(x/d.maxval, 1)
	// fmt.Println(f,math.Mod(f,1),x)
	nearest := int(f*float64(len(d.indexmarks)) + .5)
	// fmt.Println(nearest)
	for i := range d.indexmarks {
		if i == nearest {
			d.indexmarks[i].StrokeColor = d.accentColor
		} else {
			d.indexmarks[i].StrokeColor = d.lowColor
		}
	}
	d.textDisplay.Text = d.text
	fyne.Do(d.Refresh)
}

type InfiniteDialLayout struct{ d *InfiniteDial }

func (d InfiniteDialLayout) Layout(os []fyne.CanvasObject, sz fyne.Size) {
	// fmt.Println("Lay Out")
	if d.d == nil {
		fmt.Println("No d.d")
		return
	}
	if d.d.textDisplay == nil {
		fmt.Println("No text display")
		return
	}
	// fmt.Println(len(d.d.indexmarks),"markers set")
	w, h := sz.Width, sz.Height
	radius := min(w, h) / 2
	cx, cy := w/2, h/2
	d.d.textDisplay.Move(fyne.NewPos(0, 0))
	d.d.textDisplay.Resize(fyne.NewSize(w, h))
	var phi float64
	dphi := math.Pi * 2 / float64(len(d.d.indexmarks))
	for i, ch := range d.d.indexmarks {
		phi = dphi*float64(i) - math.Pi/2
		x1 := cx + radius*d.d.minorStart*float32(math.Cos(phi))
		y1 := cy + radius*d.d.minorStart*float32(math.Sin(phi))
		if i%d.d.major == 0 {
			x1 = cx + radius*d.d.majorStart*float32(math.Cos(phi))
			y1 = cy + radius*d.d.majorStart*float32(math.Sin(phi))
		}
		x2 := cx + radius*float32(math.Cos(phi))
		y2 := cy + radius*float32(math.Sin(phi))
		ch.Position1 = fyne.NewPos(x1, y1)
		ch.Position2 = fyne.NewPos(x2, y2)
		// fmt.Printf("Index %3d Pos: %5.1f,%5.1f Pos: %5.1f,%5.1f\n",i,x1,y1,x2,y2)
	}
}

func (d InfiniteDialLayout) MinSize(os []fyne.CanvasObject) fyne.Size {
	// fmt.Println("MinSize")
	return fyne.NewSize(50, 50)
}
