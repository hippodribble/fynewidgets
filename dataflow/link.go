package dataflow

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Link struct {
	widget.BaseWidget
	lines    []*canvas.Line
	points   []fyne.Position
	colour   color.Color
	from, to *Parameter
}

func NewLink(p1, p2 *Parameter) *Link {
	l := &Link{from: p1, to: p2, lines: []*canvas.Line{}}
	l.ExtendBaseWidget(l)
	return l
}

func (l *Link) CreateRenderer() fyne.WidgetRenderer {
	con := container.New(&LinkLayout{l})
	return widget.NewSimpleRenderer(con)
}

type LinkLayout struct {
	link *Link
}

func (l LinkLayout) MinSize(objects []fyne.CanvasObject) fyne.Size { return fyne.NewSize(10, 10) }
func (l LinkLayout) Layout(objects []fyne.CanvasObject, sz fyne.Size) {
	fmt.Println("layout")
	// for _, o := range objects {
	// 	if link, ok := o.(*Link); ok {
	// 		link.points[0] = link.from.square.Position()
	// 		link.points[5] = link.to.square.Position()
	// 		link.points[1] = link.from.square.Position().AddXY(MINLINKDISTANCE, 0)
	// 		link.points[4] = link.to.square.Position().AddXY(-MINLINKDISTANCE, 0)
	// 		if link.points[4].X > link.points[1].X {
	// 			fmt.Println("Right side")
	// 			// input is east of output
	// 			link.points[2].X = link.points[1].X + link.points[4].X
	// 			link.points[2].Y = link.points[1].Y
	// 			link.points[3].X = link.points[2].X
	// 			link.points[3].Y = link.points[4].Y
	// 		} else {
	// 			fmt.Println("Left side")
	// 			// input is west of output
	// 			link.points[2].X = link.points[1].X
	// 			link.points[3].X = link.points[4].X
	// 			link.points[2].Y = (link.points[1].Y + link.points[4].Y) / 2
	// 			link.points[3].Y = (link.points[1].Y + link.points[4].Y) / 2
	// 		}
	// 	}
	// }
}
