package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
)

func main() {
	ap := app.New()
	w := ap.NewWindow("EditableLabel Test")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(600, 600))
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	c := container.NewGridWithColumns(3)
	for range 7 {
		l := fynewidgets.NewEditableLabel("This is an editable label")
		c.Add(l)
	}

	name := fynewidgets.NewEntryWithFocus()
	name.SetPlaceHolder("Enter a name")

	name.OnFocusLost = func(text string) {
		fmt.Println("Entry lost focus; value:", text)

		if text == "" {
			name.SetValidationError(fmt.Errorf("name is required"))
		} else {
			name.SetValidationError(nil)
		}
	}
	c.Add(name)
	c.Add(widget.NewLabel("Click Here"))

	return c
}
