package comparatorwidget

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	lazywidget "github.com/hippodribble/fynewidgets/lazyloadingimagewidget"
)

// A widget to display multiple images with the same transform. Generally, these will be different aspects of the same underlying image,
// such as RGB or HSV planes, weighting systems, etc
type ComparatorWidget struct {
	widget.BaseWidget
	canvaswidgets   []*lazywidget.LazyWidget // these are what is displayed, and need to be customised to show crosshairs etc.
	ncols           int                      // number of columns to use in the display
	commsChannel    chan interface{}         // talk to the app
	mouseposchannel chan fyne.Position
}

func NewComparatorWidget(uris []fyne.URI, cols int, commschannel chan interface{}) (*ComparatorWidget, error) {

	w := new(ComparatorWidget)
	w.ncols = 3
	if len(uris) == 0 {
		return w, errors.New("no image URIs to process")
	}

	w.mouseposchannel = make(chan fyne.Position)
	w.commsChannel = commschannel
	commschannel <- "Created the widget. Loading the images..."
	w.ncols = cols
	w.canvaswidgets = make([]*lazywidget.LazyWidget, len(uris))
	// log.Println("Making images...")
	constrictor := make(chan int, 4)
	for i, uri := range uris {
		constrictor <- 1
		w.canvaswidgets[i] = lazywidget.NewLazyWidget(uri)
		w.canvaswidgets[i].SetMouseChannel(w.mouseposchannel)
		<-constrictor
	}
	// log.Println("Done with all images")
	w.commsChannel <- "DONE"
	w.ExtendBaseWidget(w)
	return w, nil
}

func (c *ComparatorWidget) CreateRenderer() fyne.WidgetRenderer {
	g := container.NewGridWithColumns(c.ncols)
	// g := container.New(&PictureGridLayout{})
	// g := container.New(&TightGridLayout{spacing: 10})
	for _, w := range c.canvaswidgets {
		g.Add(w)
	}
	return widget.NewSimpleRenderer(g)
}

func (c *ComparatorWidget) TellApp(f interface{}) {
	if c.commsChannel == nil {
		return
	}
	c.commsChannel <- f
}

func (c *ComparatorWidget) MouseChannel() *chan fyne.Position {
	return &c.mouseposchannel
}
