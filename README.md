# Some Widgets

## Adaptive Image Widget

Because large images can take time to render on a device, especially when the image has more pixels than the display device, it makes sense to display an image with as few pixels as possible.

The AdaptiveImageWidget addresses this by replacing the image.Image in a standard canvas.Image with a Gaussian puramid of images. This pyramid contains several layers, each at half the resolution of the previous layer, to some defined minimum size.

### Dynamic Resizing
The widget monitors its size periodically in a goroutine  to see if it is larger or smaller than the embedded image. This causes the appropriate resolution image to be selected from the pyramid and displayed.

This results in a more responsive display under resizing.

### Memory Requirements 
memory usage for the pyramid is around 1.5 times that of the underlying image.Image, including the full-resolution image, assuming size * ( 1 + 1/4 + 1/16 + ... )

### Customisation
the delay between updates can be set, in milliseconds. Images aren't resized very much, so this can be quite small, as it usually doesn't need to do anything. Default should be OK.

### Minimal Fyne Implementation
Below is a small fyne app that will use the AdaptiveImageWidget.

1. go mod tidy should import the [adaptiveimagewidget](github.com/hippodribble/fynewidgets/adaptiveimagewidget) package
1. copy the code below into main.go and go run go run ./main.go
1. The window will open. It is deliberately not very big, and is resizeable
1. Load a LARGE image file (like 100 megapixels etc) using the Canvas.Image button
1. The image will load. This might be slow - go's image library will convert the file to an image.Image - this takes as long as it takes
1. Maximise the screen, or resize it manually. It is probably not as reponsive as we would like.
1. Load the same image with the Adaptive button. Again, try to resize. It should be more responsive.



```

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


```

