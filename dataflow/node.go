package dataflow

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
)

type TaskNode struct {
	fynewidgets.DraggableBaseWidget
	title                              *canvas.Text
	rTitle, rOuter, spacer, grabhandle *canvas.Rectangle
	body                               fyne.CanvasObject
	parameters                         []*Parameter
}

func NewTaskNode(name string, body fyne.CanvasObject) *TaskNode {
	title := canvas.NewText(name, color.White)
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter

	grab := canvas.NewRectangle(theme.Color(theme.ColorNameForeground))
	grab.SetMinSize(fyne.NewSize(300, 10))
	grab.CornerRadius = 20

	rout := canvas.NewRectangle(color.Gray{64})
	rout.CornerRadius = 20
	rout.SetMinSize(fyne.NewSize(300, 300))

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(10, 10))

	rTitle := canvas.NewRectangle(color.Gray{48})
	rTitle.TopLeftCornerRadius = 20
	rTitle.TopRightCornerRadius = 20
	rTitle.SetMinSize(fyne.NewSize(300, 20))
	t := &TaskNode{
		title:      title,
		body:       body,
		rTitle:     rTitle,
		spacer:     spacer,
		rOuter:     rout,
		grabhandle: grab,
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *TaskNode) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewStack(
			t.rOuter,
			container.NewBorder(
				container.NewStack(t.rTitle, t.title),
				t.grabhandle,
				nil,
				nil,
				t.body,
			)),
	)
}

func (t *TaskNode) MinSize() fyne.Size {
	// sz:=t.title.Size()
	// w,h:=sz.Width,sz.Height
	// w=max(w,t.body.Size().Width)
	// h+=t.body.Size().Height

	return fyne.NewSize(300, 300)
}
