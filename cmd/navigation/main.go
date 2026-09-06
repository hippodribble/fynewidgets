package main

import (
	"fmt"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
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

// func main() {
// 	myApp := app.New()
// 	myWindow := myApp.NewWindow("Clip Container")

// 	big := canvas.NewRectangle(color.NRGBA{R: 0x33, G: 0x99, B: 0xcc, A: 0xff})
// 	big.CornerRadius=20
// 	// big.SetMinSize(fyne.NewSize(400, 400)) // much bigger than the window

// 	clip := container.NewClip(big)

// 	myWindow.SetContent(clip)
// 	myWindow.Resize(fyne.NewSize(150, 150)) // only this much will be drawn
// 	myWindow.ShowAndRun()
// }

// func main() {
// 	myApp := app.New()
// 	w := myApp.NewWindow("Gradient")

// 	// gradient := canvas.NewHorizontalGradient(color.White, color.Black)
// 	gradient := canvas.NewRadialGradient(color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255})
// 	w.SetContent(gradient)

// 	w.Resize(fyne.NewSize(100, 100))
// 	w.ShowAndRun()
// }

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("FileIcon Widget")

	txt := widget.NewFileIcon(storage.NewFileURI("./readme.txt"))
	jpg := widget.NewFileIcon(storage.NewFileURI("./photo.jpg"))
	png := widget.NewFileIcon(storage.NewFileURI("./photo.png"))
	pdf := widget.NewFileIcon(storage.NewFileURI("./photo.pdf"))
	zoid := fynewidgets.NewURIButton(storage.NewFileURI("./readme.html"), func() { fmt.Println("Clicked") })
	h1 := container.NewHBox(txt, jpg, zoid, png, pdf)

	items := []*fynewidgets.Status{}
	for i := range 3 {
		st := &fynewidgets.Status{Text: fmt.Sprintf("Item %d", i)}
		items = append(items, st)
	}

	chklist := fynewidgets.NewCheckList(
		"Pick Something!",
		items,
	)

	bl := container.NewBorder(nil,
		container.NewBorder(nil, nil, nil, widget.NewButton("OK", func() { fmt.Println(chklist.Selected()) })),
		nil, nil, chklist,
	)
	channel := make(chan any)

	go func() {
		for item := range channel {
			if x, ok := item.(*fynewidgets.Status); ok {
				fmt.Printf("%s was set to %v\n", x.Text, x.Selected)
				continue
			}
			if x, ok := item.([]bool); ok {
				fmt.Println("Status of all items:", x)
				continue
			}
		}
	}()

	chklist2 := fynewidgets.NewCheckListChannelled(items, "Channelled List", channel)

	redclock := fynewidgets.NewClockRed(150)

	myWindow.SetContent(container.NewVBox(h1, bl, chklist2, redclock))
	myWindow.ShowAndRun()
}
