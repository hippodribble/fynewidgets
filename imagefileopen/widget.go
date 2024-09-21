package imagefileopen

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
)

type ImageFileFinder struct {
	*widget.Button
	window fyne.Window
	URI    fyne.URI
	Image  image.Image
}

func NewImageFileFinder(w *fyne.Window) *ImageFileFinder {
	component := &ImageFileFinder{}
	component.Button=widget.NewButton("TESTING",func() {})
	// component.Button = widget.NewButtonWithIcon("Open Image File", theme.FileIcon(),func() { })
	component.window = *w

	component.ExtendBaseWidget(component)
	return component

}

func (iff *ImageFileFinder) CreateRenderer() fyne.WidgetRenderer {

	return widget.NewSimpleRenderer(iff)

}

func (iff *ImageFileFinder) Openfile() {
	dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			return
		}
		if uc == nil {
			return
		}
		uri := uc.URI()
		iff.URI = uri
		img, err := iff.loadNRGBA(uri)
		if err != nil {
			// iff.Button.SetIcon(theme.CancelIcon())
			return
		}
		iff.Image = img
		// iff.Button.SetIcon(theme.ConfirmIcon())

	}, iff.window)
	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".png", ".gif"}))
	dlg.Show()

}

func (iff *ImageFileFinder) loadImage(uri fyne.URI) (*image.Image, error) {

	img, err := imaging.Open(uri.Path())
	if err != nil {
		return nil, errors.Wrap(err, "LoadImage")
	}

	return &img, nil

}

func (iff *ImageFileFinder) loadNRGBA(uri fyne.URI) (*image.NRGBA, error) {

	img, err := iff.loadImage(uri)
	if err != nil {
		return nil, errors.Wrap(err, "LoadNRGBA")
	}
	return imaging.Clone(*img), nil

}
