package listablethumbnail

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// A grid of thumbnail images from a slice of URIs
type ThumbnailGrid struct {
	widget.BaseWidget
	grid *fyne.Container
	uris []fyne.URI
	Thumbnails []*Thumbnail
	allnone binding.Bool
}

func NewThumbnailGrid(uris []fyne.URI, w, h int, maxlabellength int,channel chan interface{}) (*ThumbnailGrid, error) {

	if uris == nil {
		return nil, errors.New("nil URI list")
	}

	if len(uris) == 0 {
		return nil, errors.New("emptyURI list")
	}

	g := &ThumbnailGrid{uris: uris, Thumbnails: make([]*Thumbnail,0)}

	for _, uri := range uris {
		t, err := NewThumbNail(uri, w, h, 1000, maxlabellength)
		if err != nil {
			continue
		}
		g.Thumbnails = append(g.Thumbnails, t)
		t.SetChannel(channel)
	}

	ncols := 1
	if len(g.Thumbnails) > 10 {
		ncols = 2
	}
	g.grid = container.NewGridWithColumns(ncols)

	for _, t := range g.Thumbnails {
		g.grid.Add(t)
	}

	g.allnone=binding.NewBool()

	g.ExtendBaseWidget(g)
	return g, nil
}

func (g *ThumbnailGrid) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(g.grid)
	return widget.NewSimpleRenderer(c)
}

func (g *ThumbnailGrid)SelectedFiles()[]fyne.URI{
	selectedlist:=make([]fyne.URI,0)
	for _,t:=range g.Thumbnails{
		if t.IsSelected(){
			selectedlist = append(selectedlist, t.URI)
		}
	}
	return selectedlist
}

func (g *ThumbnailGrid)ToggleSelection(){
	b,_:=g.allnone.Get()
	g.allnone.Set(!b)
	for _,w:=range g.Thumbnails{
		w.SetSelected(!b)
	}
}