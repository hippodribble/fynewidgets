package fynewidgets

import (
	"errors"
	"image"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const MINCOLUMNS int = 1
const MAXCOLUMNS int = 10

type SynchronisedImageGrid struct {
	widget.BaseWidget
	grid          *fyne.Container
	columns       int
	holder        *fyne.Container
	columnchannel chan int         // requests to change the number of columns in the grid are received on this channel
	infochannel   chan interface{} // status updates and progress meter changes are sent from here to the app
	datumchannel  chan Datum       // listens to changes in pan and zoom on one widget, and sends it to the others, to keep them synchronised
}

func NewSynchronisedImageGrid(numberofcolumns int, infochannel chan interface{}) (*SynchronisedImageGrid, error) {
	s := &SynchronisedImageGrid{}
	s.ExtendBaseWidget(s)

	s.infochannel = infochannel

	s.holder = container.NewStack()
	s.monitorDatumChanges()

	s.columns = 3

	s.grid = container.NewGridWithColumns(numberofcolumns)
	pl := MakeFillerImage(300, 200)
	for i := 0; i < 9; i++ {
		ci := canvas.NewImageFromImage(pl)
		ci.FillMode = canvas.ImageFillContain
		s.grid.Add(ci)
	}
	s.holder.Add(s.grid)
	return s, nil
}

func (s *SynchronisedImageGrid) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(nil, nil, nil, nil, s.holder)
	return widget.NewSimpleRenderer(c)
}

func (s *SynchronisedImageGrid) Items() ([]fyne.CanvasObject, error) {
	if len(s.grid.Objects) == 0 {
		return nil, errors.New("no items to get from grid")
	}
	return s.grid.Objects, nil
}

// for each item in the grid, returns the portion of the full image that is currently being displayed.
// Generally, this is useful when extracting a small portion of each of a group of similar images for further processing
func (s *SynchronisedImageGrid) Images() ([]image.Image, error) {
	if s.grid == nil {
		return nil, errors.New("no grid yet")
	}
	if len(s.grid.Objects) == 0 {
		return nil, errors.New("empty grid - no images to return")
	}
	images := make([]image.Image, len(s.grid.Objects))
	for i := range s.grid.Objects {
		if im, ok := s.grid.Objects[i].(*PanZoomCanvas); ok {
			im, err := im.CurrentImage()
			if err != nil {
				continue
			}
			images[i] = im
		} 
	}
	return images, nil
}

func (s *SynchronisedImageGrid) RemoveAll() error {
	if s.grid == nil {
		return errors.New("no grid to remove items from")
	}
	s.grid.RemoveAll()
	return nil
}

func (s *SynchronisedImageGrid) AddPanZoom(items ...*PanZoomCanvas) error {
	if s.grid == nil {
		return errors.New("no grid defined")
	}

	for _, pz := range items {
		pz.SetDatumChannel(s.datumchannel)
		s.grid.Add(pz)
	}
	return nil
}

func (s *SynchronisedImageGrid) SetColumnChannel(c chan int) {
	s.columnchannel = c
	go func() {
		for {
			select {
			case cols := <-s.columnchannel:

				s.columns = cols

				objects := s.grid.Objects
				s.holder.RemoveAll()
				g := container.NewGridWithColumns(cols)
				for i := 0; i < len(objects); i++ {
					g.Add(objects[i])
				}
				s.holder.Add(g)
				s.Refresh()
			}
		}
	}()
}

func (s *SynchronisedImageGrid) monitorDatumChanges() {

	s.datumchannel = make(chan Datum)
	go func() {
		for {
			select {
			case datum := <-s.datumchannel:
				if len(s.grid.Objects) == 0 {
					continue
				}
				wg := &sync.WaitGroup{}
				wg.Add(len(s.grid.Objects))
				for i := range s.grid.Objects {
					go func(i int, wg *sync.WaitGroup) {
						defer wg.Done()
						if im, ok := s.grid.Objects[i].(*PanZoomCanvas); ok {
							otherdatum := im.Datum()
							if otherdatum == nil {
								return
							}
							otherdatum.DeviceCoords = datum.DeviceCoords
							otherdatum.ImageCoords = datum.ImageCoords
							otherdatum.Scale = datum.Scale
							otherdatum.Ticks = datum.Ticks
							otherdatum.Sensitivity = datum.Sensitivity
							otherdatum.Pyramid.SetLevel(datum.Pyramid.Level())
							im.Refresh()
						}
					}(i, wg)
				}
				wg.Wait()
			}
		}
	}()

}

func (s *SynchronisedImageGrid) SetImages(uris []fyne.URI) {

	s.grid.RemoveAll()
	for i := range uris {
		im, err := NewPanZoomCanvasFromFile(uris[i], image.Pt(100, 100), s.infochannel)
		if err != nil {
			ci := canvas.NewImageFromImage(MakeFillerImage(100, 100))
			ci.FillMode = canvas.ImageFillContain
			s.grid.Add(ci)
			s.infochannel <- "failed image " + err.Error()
			continue

		}
		im.SetDatumChannel(s.datumchannel)
		s.grid.Add(im)
		s.infochannel <- "Loaded image from " + uris[i].Path()
	}
	s.Refresh()
}
