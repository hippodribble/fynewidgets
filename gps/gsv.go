package gps

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// GSV shows satellite status
type GSV struct {
	widget.BaseWidget
	c                             *fyne.Container
	record                        string
	data                          *GSVData
	textPRN, textSNR, textELEVAZI *canvas.Text
	r                             *canvas.Rectangle
	bkground                      color.RGBA
	amin                          uint8
}

func NewGSV(ID string) *GSV {
	textPRN := canvas.NewText(ID, theme.Color(theme.ColorNameForeground))
	textSNR := canvas.NewText(" ", theme.Color(theme.ColorNameForeground))
	textELEVAZI := canvas.NewText(" ", theme.Color(theme.ColorNameForeground))
	textELEVAZI.TextStyle.Monospace = true
	textPRN.TextStyle.Monospace = true
	textSNR.TextStyle.Monospace = true
	textELEVAZI.TextStyle.Bold = true
	textPRN.TextStyle.Bold = true
	textSNR.TextStyle.Bold = true

	textPRN.Color = color.RGBA{255, 128, 0, 255}
	textELEVAZI.Color = color.RGBA{255, 255, 0, 255}
	textSNR.Color = color.RGBA{0, 128, 255, 255}

	textELEVAZI.TextSize = 12
	textPRN.TextSize = 12
	textSNR.TextSize = 12

	textELEVAZI.Alignment = fyne.TextAlignCenter
	textSNR.Alignment = fyne.TextAlignCenter
	textPRN.Alignment = fyne.TextAlignCenter
	R, G, B, _ := theme.Color(theme.ColorNameForeground).RGBA()

	bkground := color.RGBA{uint8(R), uint8(G), uint8(B), 0}
	// bkground.A=uint8(0)
	// fmt.Println(bkground)

	d := GSVData{PRN: ID}
	r := &GSV{
		textPRN:     textPRN,
		textSNR:     textSNR,
		textELEVAZI: textELEVAZI,
		data:        &d,
		r:           canvas.NewRectangle(bkground),
		bkground:    bkground,
		amin:        16,
	}
	r.ExtendBaseWidget(r)
	return r
}

func (g *GSV) CreateRenderer() fyne.WidgetRenderer {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(5, 5))
	c := container.NewVBox(
		spacer,
		g.textPRN, g.textSNR,
	)
	// rect := canvas.NewRectangle(color.Gray{32})
	return widget.NewSimpleRenderer(container.NewStack(g.r, c))
}

func (g *GSV) MinSize() fyne.Size {
	return fyne.NewSize(40, 40)
}

func (g *GSV) formatInfo() {
	g.textELEVAZI.Text = fmt.Sprintf("%2d/%3d", g.data.Elevation, g.data.Azimuth)
	g.textPRN.Text = g.data.PRN
	g.textSNR.Text = fmt.Sprintf("%2d", g.data.SNR)
	g.data.lastupdate = time.Now()
	if g.data == nil {
		g.bkground.A = g.amin
		g.r.FillColor = g.bkground
	} else {
		colorfraction := float64(g.data.SNR) / 99.0 * 128.0
		g.bkground.A = uint8(colorfraction)
		g.r.FillColor = g.bkground

	}
}

func (g *GSV) clear() {
	g.textELEVAZI.Text = ""
	g.textPRN.Text = ""
	g.textSNR.Text = ""
	g.bkground.A = g.amin
	g.r.FillColor = g.bkground
}
