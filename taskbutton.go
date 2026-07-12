package fynewidgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// A TaskButton is a button that uses channels to show the progress of long-running tasks
// and pipe the output. It uses a ProgressFunc to calculate the progress and output.
//
// The basic process is:
//  * create the ProgressFunc (using NewProgressFunc) with channels for progress and output
//  * create the TaskButton with NewTaskButton(Label,ProgressFunc)
type TaskButton struct {
	widget.BaseWidget
	b, t               *canvas.Rectangle
	tf                 *canvas.Text
	Min, Max, Value, f float64
	OnClicked          *ProgressFunc
	busy               bool
	Text               string
}

func NewTaskButton(text string, f *ProgressFunc) *TaskButton {
	var bulletcolor color.Color
	r, g, b, a := theme.Color(theme.ColorNameBackground).RGBA()
	r /= 256
	g /= 256
	b /= 256
	a /= 256
	r = 255
	bulletcolor = color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	t := &TaskButton{
		Max:       1,
		b:         canvas.NewRectangle(theme.Color(theme.ColorNameBackground)),
		t:         canvas.NewRectangle(bulletcolor),
		tf:        canvas.NewText(text, theme.Color(theme.ColorNameForeground)),
		OnClicked: f,
		f:         0.025,
	}
	t.b.CornerRadius = canvas.RadiusMaximum
	t.t.CornerRadius = canvas.RadiusMaximum
	t.tf.Alignment = fyne.TextAlignCenter
	t.tf.TextStyle.Bold = true
	t.ExtendBaseWidget(t)
	return t
}

func (t *TaskButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.New(&progLayout{t}, t.b, t.t, t.tf))
}

type progLayout struct {
	t *TaskButton
}

func (l *progLayout) Layout(os []fyne.CanvasObject, sz fyne.Size) {
	os[0].Resize(sz)
	w := sz.Width * float32(l.t.f)
	os[1].Resize(fyne.NewSize(w, sz.Height))
	os[2].Move(fyne.NewPos(sz.Width/2, 0))
}

func (l *progLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 20)
}

func (t *TaskButton) MouseIn(evt *desktop.MouseEvent)    {}
func (t *TaskButton) MouseMoved(evt *desktop.MouseEvent) {}
func (t *TaskButton) MouseOut()                          {}
func (t *TaskButton) MouseDown(evt *desktop.MouseEvent)  {}
func (t *TaskButton) MouseUp(evt *desktop.MouseEvent) {
	if t.busy {
		return
	}
	t.busy = true
	go t.Execute()
}

func (t *TaskButton) SetValue(f float64) {
	t.Value = f
	t.f = (f - t.Min) / (t.Max - t.Min)
	fyne.Do(func() { t.Refresh() })
}

func (t *TaskButton) Execute() {

	t.OnClicked.progress = make(chan float64)
	go func() {
		for v := range t.OnClicked.progress {
			t.SetValue(v)
		}
	}()

	t.OnClicked.progress <- 0.0
	t.OnClicked.Execute()
	close(t.OnClicked.progress)
	t.busy = false
}
