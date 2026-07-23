package gps

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Fix struct {
	latitude, longitude    float64
	altitude, geoid_height float64
	time                   time.Time
	nsats                  int
	HDOP, VDOP, PDOP       float64
	mode                   string
	speedKt, speedKmh, CMG float64
	north                  string
}

func (f Fix) DMS(v float64) string {
	if v == 0 {
		return ""
	}
	D := math.Floor(v / 100)
	m := v - 100*D
	M := math.Floor(m)
	s := 60 * (m - M)
	return fmt.Sprintf("%3.0f° %02.0f' %05.2f\"", D, M, s)
}

func (f Fix) String() string {
	s := fmt.Sprintf("Lat %s, Long %s\n", f.DMS(f.latitude), f.DMS(f.longitude))
	s += fmt.Sprintf("Alt %.1f, Geoid Height %.1f\n", f.altitude, f.geoid_height)
	s += fmt.Sprintf("Num satellites %d\n", f.nsats)
	s += fmt.Sprintf("Time %v\n", f.time)
	s += fmt.Sprintf("HDOP %v, PDOP %.1f, VDOP %.1f\n", f.HDOP, f.PDOP, f.VDOP)
	s += fmt.Sprintf("Speed %.1f kt, %.1f km/h\n", f.speedKt, f.speedKmh)
	s += fmt.Sprintf("North Mode %v\n", f.north)
	s += fmt.Sprintf("Fix Mode %v\n", f.mode)
	return s
}

type FixDisplay struct {
	widget.BaseWidget
	tLat, tLon, tAlt, tGeoid, tTime, tNsats, tHDOP, tVDOP, tPDOP, tMode, tSpeedKt, tSpeedKmh, tCMG *canvas.Text
	fix                                                                                                    Fix
	im                                                                                                     *IntegrityMonitor
	latlong, alttime, timensatsDOP, CMGSpeed                                                               *fyne.Container
}

func NewFixDisplay() *FixDisplay {
	fd := &FixDisplay{
		fix: Fix{},
		im:  NewIntegrityMonitor(),

		tLat:   canvas.NewText("Latitude 00° 00' 00.00\"", theme.Color(theme.ColorNameForeground)),
		tLon:   canvas.NewText("Longitude 000° 00' 00.00\"", theme.Color(theme.ColorNameForeground)),
		tAlt:   canvas.NewText("Altitude 0000.0 m", theme.Color(theme.ColorNameForeground)),
		tGeoid: canvas.NewText("Geoid Height 0000.0 m\"", theme.Color(theme.ColorNameForeground)),

		tTime:  canvas.NewText(time.Now().Format("15:04:05 UTC"), theme.Color(theme.ColorNameForeground)),
		tNsats: canvas.NewText("000 Satellites in View", theme.Color(theme.ColorNameForeground)),

		tHDOP: canvas.NewText("HDOP 00.0", theme.Color(theme.ColorNameForeground)),
		tVDOP: canvas.NewText("VDOP 00.0", theme.Color(theme.ColorNameForeground)),
		tPDOP: canvas.NewText("PDOP 00.0", theme.Color(theme.ColorNameForeground)),

		tCMG:   canvas.NewText("CMG 000.0°", theme.Color(theme.ColorNameForeground)),
		tMode:  canvas.NewText("MODE N", theme.Color(theme.ColorNameForeground)),

		tSpeedKt:  canvas.NewText("Speed 0000.0 kt", theme.Color(theme.ColorNameForeground)),
		tSpeedKmh: canvas.NewText("Speed 0000.0 km/h", theme.Color(theme.ColorNameForeground)),
	}

	fd.latlong = container.NewVBox(fd.tLat, fd.tLon)
	fd.alttime = container.NewVBox(fd.tAlt, fd.tTime)
	fd.timensatsDOP = container.NewVBox(fd.tGeoid, fd.tNsats, fd.tHDOP, fd.tVDOP, fd.tPDOP)
	fd.CMGSpeed = container.NewVBox(fd.tCMG, fd.tMode,  fd.tSpeedKt, fd.tSpeedKmh)

	fd.SetActiveColours()

	fd.ExtendBaseWidget(fd)
	return fd
}

func (fd *FixDisplay) CreateRenderer() fyne.WidgetRenderer {

	c := container.NewBorder(
		nil, nil,
		fd.im,
		container.NewVBox(fd.latlong, fd.alttime),
		container.NewGridWithColumns(2, fd.timensatsDOP, fd.CMGSpeed),
	)
	return widget.NewSimpleRenderer(c)
}

func (fd *FixDisplay) Refresh() {
	fd.tLat.Text = fd.fix.DMS(fd.fix.latitude)
	fd.tLon.Text = fd.fix.DMS(fd.fix.longitude)
	fd.tTime.Text = fd.fix.time.Format("15:04:05 UTC")
	fd.tAlt.Text = fmt.Sprintf("Altitude %.1f m", fd.fix.altitude)
	fd.tGeoid.Text = fmt.Sprintf("Geoid Height %.1f m", fd.fix.geoid_height)
	fd.tHDOP.Text = fmt.Sprintf("HDOP %.1f", fd.fix.HDOP)
	fd.tVDOP.Text = fmt.Sprintf("VDOP %.1f", fd.fix.VDOP)
	fd.tPDOP.Text = fmt.Sprintf("PDOP %.1f", fd.fix.PDOP)
	fd.tNsats.Text = fmt.Sprintf("Satellites Used %2d", fd.fix.nsats)
	fd.tSpeedKmh.Text = fmt.Sprintf("Speed %6.1f km/h", fd.fix.speedKmh)
	fd.tSpeedKt.Text = fmt.Sprintf("Speed %6.1f kt", fd.fix.speedKt)
	fd.tMode.Text=fd.fix.mode
	fd.tCMG.Text=fmt.Sprintf("CMG %5.1f°",fd.fix.CMG)
}

func (fd *FixDisplay) SetVoidColours() {
	for _, k := range []*canvas.Text{
		fd.tLat, fd.tLon, fd.tAlt, fd.tGeoid, fd.tTime, fd.tNsats,
		fd.tHDOP, fd.tVDOP, fd.tPDOP, fd.tMode, fd.tSpeedKt, fd.tSpeedKmh, fd.tCMG} {
		k.Color = color.Gray{128}
	}
}

func (fd *FixDisplay) SetActiveColours() {
	for _, k := range []*canvas.Text{
		fd.tLat, fd.tLon, fd.tAlt, fd.tGeoid, fd.tTime, fd.tNsats,
		fd.tHDOP, fd.tVDOP, fd.tPDOP, fd.tMode, fd.tSpeedKt, fd.tSpeedKmh, fd.tCMG} {
		k.Color = theme.Color(theme.ColorNameForeground)
	}

	for _, k := range []*canvas.Text{
		fd.tLat, fd.tLon, fd.tAlt, fd.tGeoid, fd.tTime, fd.tNsats,
		fd.tHDOP, fd.tVDOP, fd.tPDOP, fd.tMode, fd.tSpeedKt, fd.tSpeedKmh, fd.tCMG} {
		k.TextStyle.Monospace = true
		k.TextSize = 18
	}

	for _, k := range []*canvas.Text{fd.tLat, fd.tLon, fd.tAlt, fd.tTime} {
		k.TextSize = 36
		k.Color = color.RGBA{255, 128, 0, 255}
		k.TextStyle.Bold = true
		k.Alignment = fyne.TextAlignTrailing
	}
}

type IntegrityMonitor struct {
	widget.BaseWidget
	base *canvas.Rectangle
}

func NewIntegrityMonitor() *IntegrityMonitor {
	i := &IntegrityMonitor{}
	i.base = canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	i.base.StrokeWidth = 1
	i.base.StrokeColor = theme.Color(theme.ColorNameForeground)
	i.base.SetMinSize(fyne.NewSize(100, 100))
	i.ExtendBaseWidget(i)
	return i
}

func (i *IntegrityMonitor) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(i.base)
	return widget.NewSimpleRenderer(c)
}

func (i *IntegrityMonitor) MinSize() fyne.Size { return fyne.NewSize(300, 300) }
