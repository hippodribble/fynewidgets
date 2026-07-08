package fynewidgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ItemData struct {
	Title    string `json:"title"`
	Details  string `json:"details"`
	Category string `json:"category"`
}

func NewItemData(title, details, category string) *ItemData {
	return &ItemData{
		Title:    title,
		Details:  details,
		Category: category,
	}
}

type Item struct {
	widget.BaseWidget
	Container *fyne.Container
	data      ItemData
}

func NewItem(d *ItemData) *Item {
	lbl := widget.NewLabel(d.Details)
	lbl.Wrapping = fyne.TextWrapWord
	r := canvas.NewRectangle(color.Gray{225})
	// title := canvas.NewText(d.title, theme.Color(theme.ColorNameForeground))
	title := *NewColouredText(d.Title, theme.Color(theme.ColorNameForeground))
	r2 := canvas.NewRectangle(color.Gray{200})
	r2.StrokeColor = theme.Color(theme.ColorNameBackground)
	r2.StrokeWidth = 1
	r.StrokeColor = theme.Color(theme.ColorNameBackground)
	r.StrokeWidth = 1
	i := &Item{
		Container: container.NewBorder(container.NewStack(r2, &title), nil, nil, nil, container.NewStack(r, container.NewVScroll(lbl))),
		data:      *d,
	}

	i.ExtendBaseWidget(i)
	return i
}

func (i *Item) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(i.Container)
}

func (i *Item) MouseDown(evt *desktop.MouseEvent) { fmt.Println("Mouse Down") }
func (i *Item) MouseUp(evt *desktop.MouseEvent) {
	fmt.Println("Mouse Up")
	w := fyne.CurrentApp().NewWindow(i.data.Title)
	w.Resize(fyne.NewSize(500, 500))
	w.SetContent(NewItemEditor(i))
	w.Show()
}

func NewDefaultItem() *Item {
	d := NewItemData("Title", "Body Text", "Inbox")
	return NewItem(d)
}

// Desktop is a container whose layout is managed explicitly, using this
func NewDesktop() *fyne.Container {
	d := container.New(&DesktopLayout{})
	return d
}

type DesktopLayout struct{}

func (d *DesktopLayout) Layout(ws []fyne.CanvasObject, size fyne.Size) {
	catmap := make(map[*Item]string)
	cats := make(map[string]int)
	for i := range ws {
		if ii, ok := ws[i].(*Item); ok {
			cat := ii.data.Category
			fmt.Println(cat)
			catmap[ii] = cat
			cats[cat]++
		}
	}

	fmt.Println(len(catmap), len(cats))

	n := len(ws)
	if n == 0 {
		return
	}

	// simple: compute grid cols/rows (you can plug in any algorithm here)
	cols := int(fyne.Min(float32(n), 6)) // cap at 3 cols for example
	rows := (n + cols - 1) / cols

	cellW := size.Width / float32(cols)
	cellH := size.Height / float32(rows)

	for i, w := range ws {
		r := i / cols
		c := i % cols

		pos := fyne.NewPos(
			float32(c)*cellW,
			float32(r)*cellH,
		)
		w.Move(pos)
		// fmt.Println(pos)
		w.Resize(fyne.NewSize(cellW, cellH))
		// fmt.Println(cellW, cellH)
		w.Refresh()
	}
}

func (d *DesktopLayout) MinSize(os []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(1000, 1000)
}

type ColouredText *canvas.Text

func NewColouredText(text string, colour color.Color) ColouredText {
	t := canvas.NewText(text, colour)
	t.Alignment = fyne.TextAlignCenter
	t.TextStyle.Bold = true
	t.TextSize = 24
	ct := ColouredText(t)
	return ct
}

type ItemEditor struct {
	widget.BaseWidget
	eTitle, eDescription *widget.Entry
	item                 *Item
}

func NewItemEditor(item *Item) *ItemEditor {
	e := &ItemEditor{item: item}
	title := e.item.data.Title
	description := e.item.data.Details
	e.eTitle = widget.NewEntry()
	e.eTitle.SetText(title)
	e.eDescription = widget.NewMultiLineEntry()
	e.eDescription.SetText(description)
	e.eDescription.Wrapping = fyne.TextWrapWord

	e.ExtendBaseWidget(e)
	return e
}

func (e *ItemEditor) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(
		e.eTitle, nil, nil, nil, container.NewVScroll(e.eDescription),
	)
	return widget.NewSimpleRenderer(c)
}
