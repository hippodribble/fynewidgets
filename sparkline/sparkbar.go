package sparkline

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
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
}

func NewSparkBar(w, h, n int, color color.Color) *SparkBar {
	b := &SparkBar{
		xys:   make(XYs, n),
		w:     w,
		h:     h,
		im:    image.NewRGBA(image.Rect(0, 0, w, h)),
		color: color,
	}
	b.updateImage()
	b.Image = canvas.NewImageFromImage(b.im)
	b.Image.FillMode = canvas.ImageFillContain
	b.Image.SetMinSize(fyne.NewSize(float32(w), float32(h)))
	b.ExtendBaseWidget(b)
	return b
}

func (b *SparkBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.Image)
}

func (b *SparkBar) AddPoint(xy XY) {
	if len(b.xys) < b.n {
		b.xys[b.index] = xy
		b.index = min(b.index+1, b.n-1)
	} else {
		b.xys = append(b.xys[1:], xy)
	}
	// fmt.Println(len(b.xys), b.xys)
	b.updateImage()
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

func (b *SparkBar) updateImage() {
	if b.Image == nil {
		return
	}
	b.updateRange()
	var fy float64
	for i, d := range b.xys {
		y := d.Y
		// fx = (x - b.xmin) / (b.xmax - b.xmin)
		fy = (y - b.ymin) / (b.ymax - b.ymin)
		// fmt.Println(d.X, d.Y, fy)
		Y := int(float64(b.h) * fy)
		for j := range b.h {
			if j < Y {
				b.im.Set(i, b.h-1-j, b.color)
				// fmt.Printf("%3d %3d %5.1f %5.1f %5.1f %5.1f\n", i, j, fy,y, b.ymin,b.ymax)
			} else {
				b.im.Set(i, b.h-1-j, theme.Color(theme.ColorNameBackground))
			}
		}
	}
	b.Image.Refresh()
	b.SaveImage("x.png")
}

func (b *SparkBar) SaveImage(file string) {
	w, _ := os.Create(file)
	defer w.Close()
	png.Encode(w, b.im)
}
