package fynewidgets

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// A dial clock with a spinning seconds register that can also be used as a stopwatch
//
//   - By default, it's a clock.
//   - Left click to start a timer (stopwatch)
//   - Left click again to stop the timer
//   - Move the mouse over the timer or clock to pause the display
//   - Right click to reset the timer and return to clock mode (even if the timer is running)
type ClockRed struct {
	widget.BaseWidget
	chapters             []*canvas.Line
	display              *canvas.Text
	litColor, unlitColor color.Color
	pinger               *time.Ticker
	paused               bool
	size                 float32
	timerStart           time.Time
	timerMode            bool
}

func NewClockRed(size float32) *ClockRed {
	lit := color.RGBA{255, 0, 0, 255}
	unlit := color.RGBA{128, 0, 0, 255}
	highlight := color.RGBA{255, 255, 0, 255}
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantLight {
		lit = color.RGBA{255, 0, 0, 255}
		unlit = color.RGBA{255, 200, 200, 255}
	}

	chapters := make([]*canvas.Line, 60)
	for i := range chapters {
		chapters[i] = canvas.NewLine(lit)
		chapters[i].StrokeWidth = 3
	}

	display := canvas.NewText("00:00:00", lit)
	display.SetMinSize(fyne.NewSize(size, size))
	display.TextSize = size / 8
	display.TextStyle.Bold = true
	display.Alignment = fyne.TextAlignCenter
	nanosecs := 1000000000 / len(chapters)
	tick := time.NewTicker(time.Nanosecond * time.Duration(nanosecs))
	r := &ClockRed{
		chapters:   chapters,
		display:    display,
		pinger:     tick,
		litColor:   lit,
		unlitColor: unlit,
		size:       size,
	}
	go func() {
		var intervalTick int
		for x := range r.pinger.C {
			if r.paused {
				continue
			}
			if r.timerMode {
				dt := x.Sub(r.timerStart)
				totalSeconds := int64(dt.Seconds())
				hours := totalSeconds / 3600
				minutes := (totalSeconds % 3600) / 60
				seconds := totalSeconds % 60
				r.display.Text = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
				intervalTick = int(float64(dt.Nanoseconds()%1000000000) / float64(nanosecs))
			} else {
				display.Text = x.Format("15:04:05.0")
				intervalTick = x.Nanosecond() / nanosecs

			}
			// if intervalTick==0{r.display.Color=highlight}else{r.display.Color=r.litColor}
			if intervalTick == 0 {
				r.display.TextStyle.Bold = false
			} else {
				r.display.TextStyle.Bold = true
			}
			for i, chapter := range r.chapters {
				if i == intervalTick {
					chapter.StrokeColor = highlight
				} else {
					chapter.StrokeColor = r.unlitColor
				}
			}
			fyne.Do(r.Refresh)
		}
	}()

	r.ExtendBaseWidget(r)
	return r
}

func (r *ClockRed) CreateRenderer() fyne.WidgetRenderer {
	o := []fyne.CanvasObject{r.display}
	for _, chapter := range r.chapters {
		o = append(o, chapter)
	}
	c := container.New(ClockRedLayout{r}, o...)
	return widget.NewSimpleRenderer(c)
}

func (r *ClockRed) MouseIn(e *desktop.MouseEvent)  { r.paused = true }
func (r *ClockRed) MouseOut()                      { r.paused = false }
func (r *ClockRed) MouseMoved(*desktop.MouseEvent) {}
func (r *ClockRed) MouseDown(e *desktop.MouseEvent) {
	switch e.Button {
	case desktop.MouseButtonPrimary:
		r.timerMode = !r.timerMode
		if r.timerMode {
			r.timerStart = time.Now()
			r.paused = false
		}else{
			r.paused=true
		}
	case desktop.MouseButtonSecondary:
		r.timerMode = false
	}
}
func (r *ClockRed) MouseUp(*desktop.MouseEvent) {}

type ClockRedLayout struct {
	r *ClockRed
}

func (r ClockRedLayout) Layout(os []fyne.CanvasObject, sz fyne.Size) {
	w, h := sz.Width, sz.Height
	radius := min(w, h) / 2
	cx, cy := w/2, h/2
	r.r.display.Move(fyne.NewPos(0, 0))
	r.r.display.Resize(fyne.NewSize(w, h))
	var phi float64
	dphi := math.Pi * 2 / float64(len(r.r.chapters))
	for i, ch := range r.r.chapters {
		phi = dphi*float64(i) - math.Pi/2
		x1 := cx + radius*.9*float32(math.Cos(phi))
		y1 := cy + radius*.9*float32(math.Sin(phi))
		if i%5 == 0 {
			x1 = cx + radius*.85*float32(math.Cos(phi))
			y1 = cy + radius*.85*float32(math.Sin(phi))
		}
		x2 := cx + radius*float32(math.Cos(phi))
		y2 := cy + radius*float32(math.Sin(phi))
		ch.Position1 = fyne.NewPos(x1, y1)
		ch.Position2 = fyne.NewPos(x2, y2)
	}
}

func (r ClockRedLayout) MinSize(os []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(r.r.size, r.r.size)
}
