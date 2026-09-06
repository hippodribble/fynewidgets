package fynewidgets

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// An InstrumentPanel is a set of orange "lights" with labelled names and an ON/OFF switch
//   - the level of each light can be set from 0-255
//   - lights can be turned "On" and "Off" - these set the light to 255 and 0 respectively.
type InstrumentPanel struct {
	widget.BaseWidget
	lamps    []*InstrumentLamp
	c        *fyne.Container
	nPerRow  int
	lampsize fyne.Size
	channel  chan *InstrumentLamp
}

// Constructor for InstrumentPanel
//   - names are used on the individual lamps
//   - size - base size of a single lamp
//   - nPerRow - number of lamps on each row
func NewInstrumentPanel(names []string, size fyne.Size, nPerRow int) *InstrumentPanel {
	lamps := []*InstrumentLamp{}
	for _, name := range names {
		lamp := NewInstrumentLamp(name, size)
		lamps = append(lamps, lamp)
	}

	p := &InstrumentPanel{nPerRow: nPerRow, lamps: lamps, lampsize: size}

	return p
}

func (p *InstrumentPanel) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewGridWithColumns(p.nPerRow)
	for _, l := range p.lamps {
		c.Add(l)
	}
	return widget.NewSimpleRenderer(c)
}

func (p *InstrumentPanel) MinSize() fyne.Size {
	ncols := p.nPerRow
	nrows := len(p.lamps) / ncols
	if nrows*ncols < len(p.lamps) {
		nrows++
	}
	W := ncols * int(p.lampsize.Width)
	H := nrows * int(p.lampsize.Height)
	// fmt.Println(W, H, nrows, ncols, nrows*ncols)
	return fyne.NewSize(float32(W)+10, float32(H)*1.5)
}

func (p *InstrumentPanel) SetChannel(c chan *InstrumentLamp) {
	p.channel = c
	for _, lamp := range p.lamps {
		lamp.SetChannel(c)
	}
}

// turns all lamps on
func (p *InstrumentPanel) AllOn() {
	for _, lamp := range p.lamps {
		lamp.SwitchOn()
	}
}
func (p *InstrumentPanel) On(s string) {
	for _, lamp := range p.lamps {
		if lamp.Name == s {
			lamp.SwitchOn()
		}
	}
}

// turns all lamps off
func (p *InstrumentPanel) AllOff() {
	for _, lamp := range p.lamps {
		lamp.SwitchOff()
	}
}
func (p *InstrumentPanel) Off(s string) {
	for _, lamp := range p.lamps {
		if lamp.Name == s {
			lamp.SwitchOff()
		}
	}
}

func (p *InstrumentPanel) SetLevel(s string, i uint8) {}

// An InstrumentLamp just shows a binary status for a named instrument.
// - Names are just short labels for the lamp.
// - The lamp can be clicked to toggle the status between ON and OFF.
// - The lamp can have a channel that can be used to send itself to subscribers when the ON/OFF status changes.
// - Generally, all lamps in a panel would share a channel
type InstrumentLamp struct {
	widget.BaseWidget
	outerWhite, lowerWhite, spacer *canvas.Rectangle
	toplabel, Status          *canvas.Text
	Name                           string
	lampsize                       fyne.Size
	On                             bool
	Channel                        chan *InstrumentLamp
	level                          uint8
}

func NewInstrumentLamp(name string, size fyne.Size) *InstrumentLamp {
	darkmildgreen := color.RGBA{32, 64, 32, 255}

	outerWhite := canvas.NewRectangle(darkmildgreen)
	outerWhite.StrokeColor = color.White
	outerWhite.StrokeWidth = size.Height / 20
	outerWhite.CornerRadius=20
	lowerWhite := canvas.NewRectangle(color.Transparent)
	lowerWhite.StrokeColor = color.White
	lowerWhite.StrokeWidth = size.Height / 25

	toplabel := canvas.NewText(name, color.White)
	toplabel.TextSize = size.Height / 4
	bottomlabel := canvas.NewText("", color.White)
	bottomlabel.TextSize = size.Height / 4
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(size.Height/25, size.Height/25))

	l := &InstrumentLamp{
		Name:        name,
		outerWhite:  outerWhite,
		lowerWhite:  lowerWhite,
		toplabel:    toplabel,
		Status: bottomlabel,
		spacer:      spacer,
		lampsize:    size,
		level:       255,
	}

	l.ExtendBaseWidget(l)

	return l
}

func (l *InstrumentLamp) CreateRenderer() fyne.WidgetRenderer {
	if l.outerWhite == nil {
		log.Fatalln("outer white error")
	}
	if l.lowerWhite == nil {
		log.Fatalln("lower white error")
	}
	if l.toplabel == nil {
		log.Fatalln("top label error")
	}
	if l.Status == nil {
		log.Fatalln("bottom label error")
	}
	l.toplabel.Alignment = fyne.TextAlignCenter
	l.Status.Alignment = fyne.TextAlignCenter
	l.toplabel.TextStyle.Bold = true
	l.Status.TextStyle.Bold = true
	c := container.NewStack(l.outerWhite,
		container.NewBorder(l.spacer, l.spacer, l.spacer, l.spacer,
			container.NewGridWithColumns(1,
				l.toplabel,
				container.NewBorder(l.spacer, l.spacer, l.spacer, l.spacer,
					container.NewStack(l.lowerWhite, l.Status),
				),
			),
		),
	)

	return widget.NewSimpleRenderer(c)
}

func (l *InstrumentLamp) MinSize() fyne.Size { return l.lampsize }

func (l *InstrumentLamp) MouseDown(e *desktop.MouseEvent)  {}
func (l *InstrumentLamp) MouseMoved(e *desktop.MouseEvent) {}
func (l *InstrumentLamp) MouseIn(e *desktop.MouseEvent)    {}
func (l *InstrumentLamp) MouseOut()                        {}
func (l *InstrumentLamp) MouseUp(e *desktop.MouseEvent) {
	fmt.Println(l.On)
	l.On = !l.On
	fmt.Println(l.On)
	if l.On {
		l.SwitchOn()
	} else {
		l.SwitchOff()
	}
}

func (l *InstrumentLamp) SwitchOn() {
	fmt.Println("Error")
	l.On = true
	l.outerWhite.FillColor = color.RGBA{l.level, 0, 0, 255}
	l.lowerWhite.FillColor = color.RGBA{l.level, 0, 0, 255}
	l.Status.Text = "ACK"
	l.Refresh()
	if l.Channel != nil {
		l.Channel <- l
	}
}

func (l *InstrumentLamp) SwitchOff() {
	fmt.Println("switching off")
	l.On = false
	l.outerWhite.FillColor = color.RGBA{32, 64, 32, 255}
	l.lowerWhite.FillColor = color.RGBA{32, 64, 32, 255}
	l.Status.Text = ""
	l.Refresh()
	if l.Channel != nil {
		l.Channel <- l
	}
}

func (l *InstrumentLamp) SetChannel(c chan *InstrumentLamp) {
	l.Channel = c
}
