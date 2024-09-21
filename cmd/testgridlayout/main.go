package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets/comparatorwidget"
	"github.com/hippodribble/fynewidgets/panzoomwidget"
	"github.com/hippodribble/fynewidgets/statusprogress"
)

var ch chan interface{} = make(chan interface{})
var top, bottom, centre *fyne.Container
var imagecount, columncount, padding binding.Float

func main() {
	ap := app.New()
	w := ap.NewWindow("Test Grid Layout")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(1200, 900))
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	imagecount, columncount, padding = binding.NewFloat(), binding.NewFloat(), binding.NewFloat()

	imagecount.AddListener(binding.NewDataListener(func() {
		N, _ := imagecount.Get()
		ch <- statusprogress.Message{Text: fmt.Sprintf("%d images total", int(N)), Duration: 1}
	}))

	columncount.AddListener(binding.NewDataListener(func() {
		N, _ := columncount.Get()
		ch <- statusprogress.Message{Text: fmt.Sprintf("%d Columns", int(N)), Duration: 1}
	}))

	padding.AddListener(binding.NewDataListener(func() {
		N, _ := padding.Get()
		ch <- statusprogress.Message{Text: fmt.Sprintf("%d-pixel padding", int(N)), Duration: 1}
	}))

	top = container.NewHBox()
	monitor := statusprogress.NewStatusProgress(ch)
	bottom = container.NewStack(monitor)
	centre = container.NewStack()
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(200, 5))
	top.Add(widget.NewButton("Lay Out", dolayout))
	slideN := widget.NewSliderWithData(1, 23, imagecount)
	slideN.Step = 1
	slideCols := widget.NewSliderWithData(2, 5, columncount)
	slideCols.Step = 1
	slidePad := widget.NewSliderWithData(5, 50, padding)
	slidePad.Step = 5
	top.Add(container.NewStack(r, slideN))
	top.Add(container.NewStack(r, slideCols))
	top.Add(container.NewStack(r, slidePad))
	return container.NewBorder(top, bottom, nil, nil, centre)
}

func dolayout() {
	N, _ := imagecount.Get()
	objects := make([]fyne.CanvasObject, 0)
	var img image.Image
	for i := 0; i < int(N); i++ {
		img = image.NewNRGBA(image.Rect(0, 0, 500, 250))
		zz:=img.(draw.Image)
		r, g, b := uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256))
		c := color.NRGBA{R: r, G: g, B: b, A: 255}
		draw.Draw(zz, img.Bounds(), image.NewUniform(c), image.Point{0, 0}, draw.Over)
		var yy image.Image=zz
		x, err := panzoomwidget.NewPanZoomWidget(&yy, 100, 7)
		if err!=nil{
			log.Println(i,err)
			continue
		}
		log.Println("made panzoom")
		objects = append(objects, x)
	}

	layout := comparatorwidget.NewBLayout(4, 50)
	grid := container.New(layout, objects...)
	columncount.AddListener(binding.NewDataListener(func() {
		f, _ := columncount.Get()
		layout.SetColumns(int(f))
		grid.Refresh()
	}))
	padding.AddListener(binding.NewDataListener(func() {
		f, _ := padding.Get()
		layout.SetPadding(float32(f))
		grid.Refresh()
	}))

	centre.RemoveAll()
	centre.Add(grid)
	centre.Refresh()
}
