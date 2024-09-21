package comparatorwidget

import (
	"log"
	"math"
	"sort"

	"fyne.io/fyne/v2"
	"github.com/hippodribble/fynewidgets"
	"github.com/hippodribble/fynewidgets/panzoomwidget"
)

const spacing float32 = 20 // in virtual quasi-pixels)

type PictureGridLayout struct {
	W, H float32
}

func (g *PictureGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	// go through the objects and get the largest width and height
	N := len(objects)
	rows, cols := fynewidgets.BestRowsColumns(N) // can be altered if necessary
	w := size.Width
	h := size.Height
	if w*h == 0 {
		log.Println("Screen has zero area")
		return
	}
	g.W = (w-spacing)/float32(cols) - spacing // the width of each canvas image
	g.H = (h-spacing)/float32(rows) - spacing // the height of each canvas image

	log.Printf("Grid: %d rows x %d columns, screen: %.1f x %.1f, thumbnails: %.1f x %.1f", rows, cols, w, h, g.W, g.H)
	for i, o := range objects {
		if o == nil {
			log.Println(i, "is nil")
			continue
		}
		if math.IsNaN(float64(o.Size().Height)) {
			log.Println(i, "has NaN height")
			continue
		}
		if o.Size().Height == 0 {
			log.Println(i, "has zero size")
			// continue
		}
		scalar := g.W / o.Size().Width
		rownum := i / cols
		colnum := i % cols
		log.Println(i, "before resize is", o.Size())
		o.Resize(fyne.NewSize(o.Size().Width*scalar, o.Size().Height*scalar))
		log.Println(i, "after resize is", o.Size())
		o.Move(fyne.NewPos((g.W+spacing)*float32(colnum)+spacing, (g.H+spacing)*float32(rownum)+spacing))
	}
}

func (g *PictureGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(20, 20)
	}
	if objects[0] == nil {
		return fyne.NewSize(20, 20)
	}

	N := len(objects)
	rows, cols := fynewidgets.BestRowsColumns(N) // can be altered if necessary
	sampleminsize := objects[0].MinSize()
	w := sampleminsize.Width*float32(cols) + spacing*float32(cols+1)
	h := sampleminsize.Height*float32(rows) + spacing*float32(rows+1)
	return fyne.NewSize(w, h)
}

// tries to fit a group of images to a grid,
type TightGridLayout struct {
	spacing float32
	cols    int
}

func NewTightGridLayout(spacing float32) *TightGridLayout {
	l := new(TightGridLayout)
	l.spacing = spacing
	return l
}

func (t *TightGridLayout) Layout(objects []fyne.CanvasObject, screensize fyne.Size) {
	s := t.spacing
	if len(objects) == 1 {
		rescale := (screensize.Width - 2*s) / objects[0].Size().Width
		objects[0].Resize(fyne.NewSize(screensize.Width-2*s, objects[0].Size().Height*rescale))
		objects[0].Move(fyne.NewPos(s, s))
		return
	}

	var oW, oH, wNew float32
	var c int
	if t.cols == 0 {

		for _, o := range objects {
			if o.Size().Height <= 0 || o.Size().Width <= 0 {
				return
			}
			oW += o.Size().Width
			oH += o.Size().Height
			log.Println("Found image with WxH of", o.Size().Width, o.Size().Height)
		}
		N := len(objects)                         // number of images
		oW /= float32(N)                          // mean width
		oH /= float32(N)                          // mean height
		R := screensize.Width / screensize.Height // screen aspect ratio
		// oR := oW / oH                                       // mean aspect ratio
		// oG := float32(math.Sqrt(float64(oH) * float64(oW))) // geometric mean of width and height
		log.Printf("No column count found - computing best grid for %d images of size %.1f x %.1f with padding %.1f and aspect ratio %.3f\n", N, oW, oH, s, R)

		_, c = scanRowColCount(N, oW, oH, s, R)

		// get optimum width to fill screen, given a fixed spacing
		wNew = screensize.Width - s*(float32(c)+1)
		wNew /= float32(c)
		t.cols = c

	}

	for i, o := range objects {
		row := i / t.cols
		col := i % t.cols
		x0 := s*(float32(col)+1) + float32(col)*oW
		y0 := s*(float32(row)+1) + float32(row)*oH
		scalar := wNew / o.Size().Width
		o.Resize(fyne.NewSize(o.Size().Width*scalar, o.Size().Height*scalar))
		o.Move(fyne.NewPos(x0, y0))
	}
}

func scanRowColCount(N int, w, h, s, R float32) (int, int) {
	// trial and error to test layouts
	bestrows, bestcols := -1, -1
	bestAspectRatio := 1000.0
	for ncols := 1; ncols < N; ncols++ {
		nrows := N/ncols + 1
		if nrows*ncols < N {
			nrows++
		}
		if nrows*ncols-N >= ncols {
			nrows--
		}
		//  test the layout, using the desired spacing and mean width and height
		Wnew := s*(float32(N)+1) + w*float32(N)
		Hnew := s*(float32(N)+1) + h*float32(N)
		Rnew := Wnew / Hnew
		relativeAspectRatio := math.Abs(float64(Rnew) - float64(R))
		if relativeAspectRatio < bestAspectRatio {
			bestrows = nrows
			bestcols = ncols
			bestAspectRatio = relativeAspectRatio
		}
	}

	// now we have the best aspect ratio

	log.Printf("%d objects: %d rows x %d columns (aspect=%0.2f)\n", N, bestrows, bestcols, bestAspectRatio)
	return bestrows, bestcols
}

func (t *TightGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {

	return fyne.NewSize(100, 100)
}

type LayoutB struct {
	cols int     // desired number of columns
	pad  float32 // pixel padding
	busy bool
}

func NewBLayout(ncols int, pad float32) *LayoutB {
	return &LayoutB{cols: ncols, pad: pad}
}

func (b *LayoutB) Layout(objects []fyne.CanvasObject, screensize fyne.Size) {
	ncols := b.cols
	if ncols == 0 {
		return
	}
	if b.busy {
		return
	}
	b.busy = true
	defer func() { b.busy = false }()
	if ncols == 0 {
		ncols = 1
	}
	log.Println("resize interior", ncols)
	w := screensize.Width - (float32(ncols)+1)*b.pad
	w /= float32(ncols)
	rows := len(objects) / ncols
	if rows*ncols < len(objects) {
		rows += 1
	}
	h := screensize.Height - (float32(rows)+1)*b.pad
	h /= float32(rows)
	for i := 0; i < len(objects); i++ {
		objects[i].Resize(fyne.NewSize(w, h))
		if o, ok := objects[i].(*panzoomwidget.PanZoomWidget); ok {
			o.Resize(fyne.NewSize(w, h))
			log.Println("resize interior", ncols)
		} else {
			continue
		}
		r := i / ncols
		c := i % ncols
		x := (w+b.pad)*float32(c) + b.pad
		y := (h+b.pad)*float32(r) + b.pad
		objects[i].Move(fyne.NewPos(x, y))
	}
}

func (b *LayoutB) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(10, 10)
	}
	ncols := b.cols
	if ncols == 0 {
		ncols = 1
	}
	rows := len(objects) / ncols
	if rows*ncols < len(objects) {
		rows++
	}
	log.Println(rows, "rows", ncols, "cols", len(objects), "images")
	rowwidths := make([]float64, rows)
	columnheights := make([]float64, ncols)
	for i := 0; i < len(objects); i++ {
		r := i / ncols
		c := i % ncols
		// log.Println(r, c)
		rowwidths[r] += float64(objects[i].Size().Width)
		columnheights[c] += float64(objects[i].Size().Height)
	}
	sort.Float64s(rowwidths)
	sort.Float64s(columnheights)
	maxwidth := rowwidths[len(rowwidths)-1]
	maxheight := columnheights[len(columnheights)-1]
	W := float32(maxwidth + float64(b.pad)*(float64(ncols)+1))
	H := float32(maxheight + float64(b.pad)*(float64(rows)+1))
	return fyne.NewSize(W, H)
}

func (b *LayoutB) SetColumns(n int) {
	b.cols = n
}

func (b *LayoutB) SetPadding(n float32) {
	b.pad = n
}
