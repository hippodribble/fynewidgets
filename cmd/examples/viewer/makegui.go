package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
)

func makegui() {
	topButtons := container.NewHBox(
		widget.NewButton("File", openfile),
		widget.NewButton("Folder", openfolder),
	)
	toptools := widget.NewToolbar(
		widget.NewToolbarAction(theme.FileImageIcon(), openfile),
		widget.NewToolbarAction(theme.FolderOpenIcon(), openfolder),
	)
	top := container.NewHBox(topButtons, widget.NewSeparator(), toptools)

	bottom := fynewidgets.NewEventLabel(&bus)

	mainwindow.SetContent(
		container.NewBorder(top, bottom, nil, nil, widget.NewLabel("Welcome")),
	)
	mainwindow.CenterOnScreen()
}

func openfile() {
	for i := 0; i < 100; i++ {
		bus.Publish("status:show", fmt.Sprintf("Opening file %d", i))
		time.Sleep(time.Millisecond*10)
	}
}
func openfolder() { bus.Publish("status:show", "Opening folder...") }
