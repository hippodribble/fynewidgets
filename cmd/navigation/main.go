package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// func main() {
// 	myApp := app.New()
// 	myWindow := myApp.NewWindow("Navigation Container")

// 	var nav *container.Navigation
// 	root := container.NewVBox(
// 		widget.NewLabel("The first screen"),
// 		widget.NewButton("Show detail", func() {
// 			detail := widget.NewLabel("The detail of the item")
// 			nav.PushWithTitle(detail, "Detail")
// 		}),
// 	)

// 	nav = container.NewNavigationWithTitle(root, "Home")

// 	myWindow.SetContent(nav)
// 	myWindow.Resize(fyne.NewSize(300, 200))
// 	myWindow.ShowAndRun()
// }

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Clip Container")

	big := canvas.NewRectangle(color.NRGBA{R: 0x33, G: 0x99, B: 0xcc, A: 0xff})
	big.SetMinSize(fyne.NewSize(400, 400)) // much bigger than the window

	clip := container.NewClip(big)

	myWindow.SetContent(clip)
	myWindow.Resize(fyne.NewSize(150, 150)) // only this much will be drawn
	myWindow.ShowAndRun()
}
