package cmd



import (
	"image"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets/adaptiveimagewidget"
)

var status *widget.Label
var progress *widget.ProgressBarInfinite
var stack *fyne.Container

func main() {
	ap := app.New()
	w := ap.NewWindow("Check ImageWidget")
	w.SetContent(gui())
	// w.SetFullScreen(true)
	w.Resize(fyne.NewSize(800, 800))
	w.ShowAndRun()

}

func gui() *fyne.Container {
	bCanvas := widget.NewButton("Canvas.Image", openImageNormally)
	bAdaptive := widget.NewButton("AdaptiveImageWidget", openImage)
	top := container.NewHBox(bCanvas,bAdaptive)

	stack = container.NewStack()

	status = widget.NewLabel("Load a large image")
	progress = widget.NewProgressBarInfinite()
	progress.Hide()
	r := canvas.NewRectangle(color.Transparent)
	r.Resize(fyne.NewSize(200, 10))
	bottom :=  container.NewBorder(nil,nil,nil,container.NewStack(r,progress),status)

	b := container.NewBorder(top, bottom, nil, nil, stack)
	return b
}

func openImage() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		stack.RemoveAll()
		progress.Show()
		defer progress.Hide()
		if err != nil {
			status.SetText("File dialog error")
			return
		}
		uri := uc.URI()
		status.SetText("Loading file " + uri.Path())
		progress.Start()
		f, err := os.Open(uri.Path())
		if err != nil {
			status.SetText("Error opening file")
			return
		}
		im, format, err := image.Decode(f)
		if err != nil {
			status.SetText("File is not an image")
			return
		}

		progress.Stop()
		status.SetText("Displaying " + format)
		progress.Start()
		widget, err := adaptiveimagewidget.NewImageWidget(im, 200)
		if err != nil {
			status.SetText("Error creating widget: " + err.Error())
			return
		}
		widget.SetUpdateRate(200)
		stack.Add(widget)
		stack.Refresh()
		progress.Stop()
		status.SetText("Done")
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.Show()
}

func openImageNormally() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		stack.RemoveAll()
		progress.Show()
		defer progress.Hide()
		if err != nil {
			status.SetText("File dialog error")
			return
		}
		uri := uc.URI()
		status.SetText("Loading file " + uri.Path())
		progress.Start()
		f, err := os.Open(uri.Path())
		if err != nil {
			status.SetText("Error opening file")
			return
		}
		im, format, err := image.Decode(f)
		if err != nil {
			status.SetText("File is not an image")
			return
		}

		progress.Stop()
		status.SetText("Displaying " + format)
		progress.Start()
		widget:=canvas.NewImageFromImage(im)
		widget.FillMode=canvas.ImageFillContain
		widget.ScaleMode=canvas.ImageScaleFastest
		stack.Add(widget)
		stack.Refresh()
		progress.Stop()
		status.SetText("Done")
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.Show()
}
