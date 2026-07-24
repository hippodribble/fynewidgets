package sparkline

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type XY struct {
	X, Y float64
}

type XYs []XY

type SparkBar struct {
	widget.BaseWidget
	Image                  *canvas.Image
	xys                    XYs
	w, h, index, n         int
	xmin, xmax, ymin, ymax float64
	im                     *image.RGBA
	color                  color.Color
	bars                   []*canvas.Rectangle
	gap                    float32
}

func NewSparkBar(w, h, n int, color color.Color) *SparkBar {
	b := &SparkBar{
		xys:   make(XYs, n),
		w:     w,
		h:     h,
		im:    image.NewRGBA(image.Rect(0, 0, w, h)),
		color: color,
		bars:  make([]*canvas.Rectangle, n),
		gap:   .33333,
	}
	for i := range b.bars {
		b.bars[i] = canvas.NewRectangle(color)
	}
	// b.updateImage()
	b.Image = canvas.NewImageFromImage(b.im)
	b.Image.FillMode = canvas.ImageFillStretch
	// b.Image.SetMinSize(fyne.NewSize(float32(w), float32(h)))
	b.ExtendBaseWidget(b)
	return b
}

func (b *SparkBar) CreateRenderer() fyne.WidgetRenderer {
	// return widget.NewSimpleRenderer(b.Image)
	c := container.New(&SparkBarLayout{b})
	for i := range b.bars {
		c.Add(b.bars[i])
	}
	return widget.NewSimpleRenderer(c)
}

func (b *SparkBar) AddPoint(xy XY) {
	if len(b.xys) < b.n {
		b.xys[b.index] = xy
		b.index = min(b.index+1, b.n-1)
	} else {
		b.xys = append(b.xys[1:], xy)
	}
	// fmt.Println(len(b.xys), b.xys[len(b.xys)-1])
	// b.updateImage()
	b.Refresh()
}

func (b *SparkBar) updateRange() {
	xlo := 1.0e+20
	xhi := -1.0e+20
	ylo := 1.0e+20
	yhi := -1.0e+20
	for i := range b.xys {
		xlo = min(xlo, b.xys[i].X)
		ylo = min(ylo, b.xys[i].Y)
		xhi = max(xhi, b.xys[i].X)
		yhi = max(yhi, b.xys[i].Y)
	}
	b.xmin = xlo
	b.ymin = ylo
	b.xmax = xhi
	b.ymax = yhi
	// fmt.Println(xlo, xhi, ylo, yhi)
}

// func (b *SparkBar) updateImage() {
// 	if b.Image == nil {
// 		return
// 	}
// 	if len(b.bars) != len(b.xys) {
// 		log.Fatalln(len(b.bars), "bars", len(b.xys), "xys")
// 	}
// 	b.updateRange()
// 	var fy float64
// 	for i, d := range b.xys {
// 		y := d.Y
// 		// fx = (x - b.xmin) / (b.xmax - b.xmin)
// 		fy = (y - b.ymin) / (b.ymax - b.ymin)
// 		// fmt.Println(d.X, d.Y, fy)
// 		Y := int(float64(b.h) * fy)
// 		for j := range b.h {
// 			if j < Y {
// 				b.im.Set(i, b.h-1-j, b.color)
// 				// fmt.Printf("%3d %3d %5.1f %5.1f %5.1f %5.1f\n", i, j, fy,y, b.ymin,b.ymax)
// 			} else {
// 				b.im.Set(i, b.h-1-j, theme.Color(theme.ColorNameBackground))
// 			}
// 		}
// 	}
// 	b.Image.Refresh()
// 	// b.SaveImage("x.png")
// }

func (b *SparkBar) SaveImage(file string) {
	w, _ := os.Create(file)
	defer w.Close()
	png.Encode(w, b.im)
}

type SparkBarLayout struct {
	b *SparkBar
}

func (l SparkBarLayout) Layout(os []fyne.CanvasObject, sz fyne.Size) {
	l.b.updateRange()
	W, H := sz.Width, sz.Height
	w := W / float32(len(l.b.bars))

	thisbar := 0
	// fmt.Println(len(os), "Objects to lay out", len(l.b.bars), "bars")
	for _, o := range os {
		if o == nil {
			thisbar++
			continue
		}
		// fmt.Printf("%T\n", o)
		if r, ok := o.(*canvas.Rectangle); ok {
			if r == nil {
				thisbar++
				continue
			}
			// fmt.Println(r.Size())
			y := l.b.xys[thisbar].Y
			fy := (y - l.b.ymin) / (l.b.ymax - l.b.ymin)
			Y1 := H * float32((1 - fy))
			Y2 := H
			X1 := w * float32(thisbar)
			X2 := X1 + w
			l.b.bars[thisbar].Move(fyne.NewPos(X1, Y1))
			l.b.bars[thisbar].Resize(fyne.NewSize((X2-X1)*(1-l.b.gap), Y2-Y1))
		}

		thisbar++
	}
}

func (l SparkBarLayout) MinSize(os []fyne.CanvasObject) fyne.Size { return fyne.NewSize(100, 100) }

// for i, d := range b.xys {
// 	y := d.Y
// 	// fx = (x - b.xmin) / (b.xmax - b.xmin)
// 	fy = (y - b.ymin) / (b.ymax - b.ymin)
// 	// fmt.Println(d.X, d.Y, fy)
// 	Y := int(float64(b.h) * fy)
// 	for j := range b.h {
// 		if j < Y {
// 			b.im.Set(i, b.h-1-j, b.color)
// 			// fmt.Printf("%3d %3d %5.1f %5.1f %5.1f %5.1f\n", i, j, fy,y, b.ymin,b.ymax)
// 		} else {
// 			b.im.Set(i, b.h-1-j, theme.Color(theme.ColorNameBackground))
// 		}
// 	}
// }
