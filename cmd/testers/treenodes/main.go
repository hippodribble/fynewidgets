package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/hippodribble/fynewidgets/tree"
)

func main() {
	ap := app.NewWithID("test")
	w := ap.NewWindow("Tree Tests")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(800, 800))
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {

	thistree := tree.NewTree("root")

	for i := range 3 {
		child,_:=thistree.Root().AddChild(fmt.Sprintf("child %02d", i+1))
		for j := range 3 {
			grandchild,_:=child.AddChild(fmt.Sprintf("grandchild %02d", j+1))
			for k := range 3 {
				grandchild.AddChild(fmt.Sprintf("great-grandchild %02d", k+1))
			}
		}	
	}

	fmt.Println("Tree depth is",thistree.MaxDepth())
	nodeview := tree.NewTreeVis(thistree)

	return nodeview
}
