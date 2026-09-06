package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets/dataflow"
)

var ch = make(chan *dataflow.Parameter)
var surface *dataflow.FlowDiagram

func main() {
	ap := app.New()
	w := ap.NewWindow("Data Flow v0.1")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(400, 400))
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	p1 := dataflow.NewFloatParameter("test float input", true).AsFormItem()
	p1.HintText="Needs to be a float"
	p2 := dataflow.NewFloatParameter("test float output", false).AsFormItem()
	p2.HintText="Needs to be a float"
	f1:=dataflow.FloatParameterFromFormItem(p1)
	fmt.Println(f1.Name)
	form:=widget.NewForm(p1,p2)
	return  container.NewStack(form)
}

func testparam() {

}
