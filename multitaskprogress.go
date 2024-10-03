package fynewidgets

import (
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	eventbus "github.com/dtomasi/go-event-bus/v3"
)

// message format for communicating task progress to the widget through a channel
type TaskProgressMessage struct {
	Task  int
	Value float64
}

// holds graphical elements and a reference to the underlying widget.
type progrenderer struct {
	fyne.WidgetRenderer
	Prog       *MultiTaskProgress
	Rectangles []fyne.CanvasObject
}

// the rectangles to be drawn and filled to represent each task
func (p *progrenderer) Objects() []fyne.CanvasObject {

	return p.Rectangles
}

// set of vertical bars spaced by the theme padding. No outer padding
func (p *progrenderer) Layout(size fyne.Size) {
	if len(p.Prog.Names) == 0 {
		return
	}
	objects := p.Rectangles
	vals := p.Prog.Progress
	for i, o := range objects {
		if o, ok := o.(*canvas.Rectangle); ok {
			o.Resize(size)
		}
		pad := theme.Padding()

		w := (size.Width - pad*float32(len(objects)/2-1)) / float32(len(objects)) * 2
		j := i / 2
		if j*2 == i {
			o.Resize(fyne.NewSize(w, size.Height))
			o.Move(fyne.NewPos((w+theme.Padding())*float32(j), 0))

		} else {
			x, _ := vals[j].Get()
			o.Resize(fyne.NewSize(w, size.Height*float32(x)))
			o.Move(fyne.NewPos((w+theme.Padding())*float32(j), size.Height-size.Height*float32(x)))
		}
	}
}

// uses theme padding and a minimum bar width and height (package variable)
func (p *progrenderer) MinSize() fyne.Size {

	j := len(p.Rectangles) / 2
	d := theme.Padding()

	size := fyne.NewSize((d+barwidth)*float32(j)-d, barheight)

	return size
}

// just lays out the widget at the current size
func (p *progrenderer) Refresh() {
	p.Layout(p.Prog.Size())
}

func (p *progrenderer) Destroy() {

}

// tracks progress of multiple tasks as a set of columns laid out horizontally which fill up vertically
type MultiTaskProgress struct {
	widget.BaseWidget

	Names    []string
	Progress []binding.Float
	bus      *eventbus.EventBus
}

func NewMultiTaskProgress(names []string, bus *eventbus.EventBus) *MultiTaskProgress {

	p := &MultiTaskProgress{Names: names, bus: bus, Progress: make([]binding.Float, len(names))}
	p.ExtendBaseWidget(p)

	for i := range p.Progress {
		p.Progress[i] = binding.NewFloat()
	}

	echan:=p.bus.Subscribe("progress:multi")
	go func() {
		for message:= range echan{
			m:=message.Data.(*TaskProgressMessage)
			p.SetProgress(m.Task, m.Value)
			message.Done()
		}

	}()
	
	return p

}

// the graphical elements are stored in the renderer, and can be recreated at any time, as their data comes from the widget
func (p *MultiTaskProgress) CreateRenderer() fyne.WidgetRenderer {

	r := &progrenderer{Prog: p}
	r.Rectangles = make([]fyne.CanvasObject, 0)

	for i := 0; i < len(p.Names); i++ {
		var r1, r2 fyne.CanvasObject
		r1 = canvas.NewRectangle(theme.Color(theme.ColorNameFocus))
		r.Rectangles = append(r.Rectangles, r1)
		r2 = canvas.NewRectangle(theme.Color(theme.ColorNameWarning))
		r.Rectangles = append(r.Rectangles, r2)
	}

	return r
}

// sets the progress of an individual task
func (p *MultiTaskProgress) SetProgress(i int, v float64) {
	p.Progress[i].Set(v)

	if !p.stillGoing() {
		p.Hide()
	}
	p.Refresh()
}

func (p *MultiTaskProgress) stillGoing() bool {
	stillgoing := false
	for i := range p.Names {
		k, _ := p.Progress[i].Get()
		if k >= 1 {
			stillgoing = stillgoing || false
		} else {
			stillgoing = stillgoing || true
		}
	}
	return stillgoing
}

// sets the progress of all tasks to zero. Does not delete any tasks
func (p *MultiTaskProgress) Reset() {

	p.Progress = make([]binding.Float, len(p.Names))
	p.Refresh()
}

func (p *MultiTaskProgress) MouseDown(e *desktop.MouseEvent) {

}

// for long-running tasks, it may be helpful to show a more detailed popup of the tasks
func (p *MultiTaskProgress) MouseUp(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonPrimary {
		r := canvas.NewRectangle(color.Transparent)
		r.SetMinSize(fyne.NewSize(300, 3))
		go func() {
			log.Println("Up with primary button")

			form := widget.NewForm()
			for i := range p.Names {
				form.Append(p.Names[i], container.NewStack(r, widget.NewProgressBarWithData(p.Progress[i])))

			}
			w := form.Size().Width
			form.Resize(fyne.NewSize(w*2, form.Size().Height))
			view := container.NewVBox(form)
			
			popup := widget.NewModalPopUp(view, fyne.CurrentApp().Driver().AllWindows()[0].Canvas())
			popup.Canvas.SetOnTypedKey(func(e *fyne.KeyEvent){
				popup.Hide()
			})
			popup.Show()
			for p.stillGoing() {
				time.Sleep(100 * time.Millisecond)
			}
			popup.Hide()
			popup = nil

		}()
	}
}
