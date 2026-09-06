package dataflow

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var MINLINKDISTANCE float32 = 10
var lastclicked, secondlastclicked *Parameter

type FlowDiagram struct {
	widget.BaseWidget
	tasks   []*TaskNode
	links   []*Link
	objects []fyne.CanvasObject

	LinkChannel chan *Parameter
}

func NewFlowDiagram(tasks ...*TaskNode) *FlowDiagram {
	d := &FlowDiagram{tasks: tasks, links: []*Link{}, LinkChannel: make(chan *Parameter)}
	go d.listen()
	d.ExtendBaseWidget(d)
	return d
}

func (d *FlowDiagram) CreateRenderer() fyne.WidgetRenderer {
	fmt.Println("create renderer")
	c := container.New(&diagramlayout{d})
	for _, t := range d.tasks {
		c.Add(t)
	}
	for _, llist := range d.links {
		c.Add(llist)
		// for _, l := range llist.lines {
		// 	c.Add(l)
		// }
	}
	return widget.NewSimpleRenderer(c)
}

func (d *FlowDiagram) AddTaskNode(task *TaskNode) {
	d.tasks = append(d.tasks, task)
}

func (d *FlowDiagram) AddLink(p1, p2 *Parameter) {
	fmt.Println("Add link here")

	if (*p1).IsInput() || !(*p2).IsInput() {
		fmt.Println("reject - output not output or input not input!")
		return
	}
	d.links = append(d.links, NewLink(p1, p2))
	d.Refresh()
}

type diagramlayout struct {
	d *FlowDiagram
}

func (l diagramlayout) Layout(os []fyne.CanvasObject, sz fyne.Size) {
	fmt.Println("lay out diagram", len(os))
	for _, o := range os {
		fmt.Printf("%T\n", o)
		if t, ok := o.(*TaskNode); ok {
			wt := t.title.Size().Width
			wb := t.body.Size().Width
			fmt.Println("widths of title and body:", wt, wb)
			ht := t.title.Size().Height
			hb := t.body.Size().Height
			if hb < 0 {
				t.body.Resize(fyne.NewSize(t.body.Size().Width, -hb))
			}

			fmt.Println("heights of title and body:", ht, hb)
			wt = max(wt, wb) + 10

			h := t.body.Size().Height + t.title.Size().Height + 10
			h = max(h, 100)
			wt = max(wt, 100)

			fmt.Println(" updating node size", wt, h)
			t.Resize(fyne.NewSize(wt, h))
			t.Resize(fyne.NewSize(wt, h))
		}
		if link, ok := o.(*Link); ok {
			fmt.Println(" updating link")

			if link.points[4].X > link.points[1].X {
				// input is east of output
				fmt.Println(" to right")

			} else {
				// input is west of output
				fmt.Println(" to left")

			}
		}
	}
}

func (l diagramlayout) MinSize(os []fyne.CanvasObject) fyne.Size { return fyne.NewSize(700, 500) }

func (d *FlowDiagram) MouseUp(e *desktop.MouseEvent) {}
func (d *FlowDiagram) MouseDown(e *desktop.MouseEvent) {
	switch e.Button {
	case desktop.MouseButtonPrimary:
		fmt.Println("Left")
	case desktop.MouseButtonSecondary:
		fmt.Println("Right")
	}
}

func (d *FlowDiagram) listen() {
	for pv := range d.LinkChannel {
		if pv == nil {
			log.Fatalln("Empty node parameter clicked - should not be possible")
		}
		if lastclicked != nil {
			secondlastclicked = lastclicked
			lastclicked = pv
			fyne.Do(func() { d.AddLink(secondlastclicked, lastclicked) })
		}
		lastclicked = pv
	}
}
