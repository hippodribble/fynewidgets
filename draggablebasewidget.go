package fynewidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type DragMode int

// Move is the norm
// Resize only happens when the bottom-right corner is clicked
const (
	Move DragMode = iota
	Resize
)

type DraggableBaseWidget struct {
	widget.BaseWidget
	dragstartmouse, dragfinishmouse fyne.Position
	dragstartwidget                 fyne.Position
	mode                            DragMode
	initialSize                     fyne.Size
}

func (d *DraggableBaseWidget) Dragged(ev *fyne.DragEvent) {
	if d.mode == Move {
		d.Move(d.dragstartwidget.Add(ev.AbsolutePosition.Subtract(d.dragstartmouse)))
	}
	if d.mode == Resize {
		change := ev.AbsolutePosition.Subtract(d.dragstartmouse)
		d.Resize(d.initialSize.Add(change))
	}
}

func (d *DraggableBaseWidget) DragEnd() {
}

func (d *DraggableBaseWidget) MouseDown(ev *desktop.MouseEvent) {
	d.dragstartmouse = ev.AbsolutePosition
	d.dragstartwidget = d.Position()
	if ev.Position.X/d.Size().Width > 0.9 && ev.Position.Y/d.Size().Height > 0.9 {
		// fmt.Println("Resize")
		d.mode = Resize
		d.initialSize = d.Size()
	} else {
		// fmt.Println("Move")
		d.mode = Move
	}
}
func (d *DraggableBaseWidget) MouseUp(ev *desktop.MouseEvent) {
	d.dragfinishmouse = ev.AbsolutePosition
}
