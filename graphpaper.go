package fynewidgets

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type GraphPaper struct {
	widget.BaseWidget
	minorSpace, majorSpace            int
	w, h                              int
	majorColor, minorColor, baseColor color.Color
	cv                                *canvas.Image
}

func NewGraphPaper(major, minor, w, h int, majorColor, minorColor, baseColor color.Color) (*GraphPaper, error) {

	g := &GraphPaper{minorSpace: minor, majorSpace: major, majorColor: majorColor, minorColor: minorColor, baseColor: baseColor,w: w,h: h}
	err := g.MakePaper()
	if err != nil {
		return nil, err
	}

	g.ExtendBaseWidget(g)
	return g, nil
}

func (g *GraphPaper) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(g.cv))
}

func (g *GraphPaper) MakePaper() error {
	im := image.NewRGBA(image.Rect(0, 0, g.w, g.h))

	for x := range g.w {
		for y := range g.h {
			im.Set(x,y,g.baseColor)
		}
	}

	for x := 0; x <= g.w; x += g.minorSpace { // vertical minor lines
		for y := range g.h {
			im.Set(x, y, g.minorColor)
		}
	}
	for x := 0; x <= g.w; x += g.majorSpace { // vertical major lines
		for y := range g.h {
			im.Set(x, y, g.majorColor)
		}
	}

	for x := range g.h {
		for y := 0; y <= g.h; y += g.minorSpace { // horizontal minor lines
			im.Set(x, y, g.minorColor)
		}
	}
	for x := range g.h {
		for y := 0; y <= g.h; y += g.majorSpace { // horizontal minor lines
			im.Set(x, y, g.majorColor)
		}
	}

	saveimage(im)

	g.cv = canvas.NewImageFromImage(im)
	g.cv.FillMode = canvas.ImageFillOriginal
	// g.cv.SetMinSize(fyne.NewSize(float32(g.w),float32(g.h)))
	return nil
}


func saveimage(im image.Image){
	w,err:=os.Create("testgraphpaperimage.png")
	if err!=nil{log.Fatalln(err)}
	defer w.Close()
	err=png.Encode(w,im)
	if err!=nil{log.Fatalln(err)}
}