package lazywidget

import (
	"image"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type SingleImageLayout struct {
	VerticalPosition, HorizontalPosition float32
}

func (l *SingleImageLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	// s := objects[0].Size()
	objects[0].Resize(size)
	objects[0].Refresh()
}

func (l *SingleImageLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return objects[0].MinSize()
}

var CURSOR_COLOR color.Color = color.NRGBA{255, 255, 0, 64}

type LazyWidget struct {
	widget.BaseWidget

	cv              *canvas.Image
	URI             fyne.URI
	mousepos        fyne.Position
	mousePosChannel chan fyne.Position
	composite       *fyne.Container
}

func NewLazyWidget(uri fyne.URI) *LazyWidget {
	w := LazyWidget{URI: uri, mousePosChannel: make(chan fyne.Position)}
	placeholder := image.NewNRGBA(image.Rect(0, 0, 50, 50))
	w.cv = canvas.NewImageFromImage(placeholder)
	w.cv.FillMode = canvas.ImageFillContain
	w.cv.SetMinSize(fyne.NewSize(200, 200))
	w.cv.Move(fyne.NewPos(0, 0))

	w.composite = container.New(&SingleImageLayout{}, w.cv)

	go func() {
		f, _ := os.Open(uri.Path())
		defer f.Close()
		imgLoaded, _, _ := image.Decode(f)
		w.cv.Image = imgLoaded
		// w.cv = canvas.NewImageFromFile(uri.Path())
		w.cv.Resize(fyne.NewSize(200, 200))
		w.Refresh()
	}()
	w.ExtendBaseWidget(&w)
	return &w
}

func (w *LazyWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.composite)
}

func (w *LazyWidget) GetMouseChannel() chan fyne.Position {
	return w.mousePosChannel
}

func (w *LazyWidget) SetMouseChannel(channel chan fyne.Position) {
	w.mousePosChannel = channel
}

func (w *LazyWidget) MouseIn(e *desktop.MouseEvent) {}
func (w *LazyWidget) MouseOut()                     {}
func (w *LazyWidget) MouseMoved(e *desktop.MouseEvent) {
	// log.Println("Moved")
	w.mousepos = e.Position

}

func(w *LazyWidget)Refresh(){
	w.Resize(w.Size())
}
