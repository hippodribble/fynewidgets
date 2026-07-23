package gps

import (
	"bufio"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
	"go.bug.st/serial"
)

// const MAXLENGTH = 5

type GPSMonitor struct {
	widget.BaseWidget
	listHistory              *widget.List
	history, filteredHistory GPSHistory
	chanIn                   chan GPSRecord
	radioSource              *widget.RadioGroup
	radioFilter              *widget.RadioGroup
	done                     bool
	status                   *widget.Label
	gsvs                     []*GSV
	tickClear                *time.Ticker
	radar                    *GPSRadar
	fixdisplay               *FixDisplay
	light                    *Light
	TTL                      float64
	historyFilter            string
	fixHistory               []*Fix
}

func NewGPSMonitor() *GPSMonitor {

	Nsats := 96

	mon := &GPSMonitor{

		radioSource:     widget.NewRadioGroup([]string{"NET", "USB", "FILE", "DUMMY"}, func(s string) {}),
		radioFilter:     widget.NewRadioGroup([]string{"GSV", "GLL", "GGA", "RMC", "GSA", "VTG", "ZDA"}, func(s string) {}),
		chanIn:          make(chan GPSRecord),
		status:          widget.NewLabel("ready"),
		history:         *NewGPSHistory(500),
		filteredHistory: *NewGPSHistory(500),
		gsvs:            make([]*GSV, Nsats+1),
		tickClear:       time.NewTicker(time.Second * 2),
		light:           NewLight(color.RGBA{255, 128, 0, 255}, 20),
		TTL:             2.0,
		fixdisplay:      NewFixDisplay(),
		fixHistory:      []*Fix{},
	}

	for i := range Nsats + 1 {
		prn := fmt.Sprintf("%02d", i+1)
		mon.gsvs[i] = NewGSV(prn)
		mon.gsvs[i].data.PRN = prn
	}

	// make slice from map for the radar display - could probably do this the radar code

	mon.radar = NewGPSRadar(mon.gsvs, mon.TTL)

	mon.radioSource.OnChanged = mon.changeSource
	mon.radioSource.Horizontal = false

	mon.radioFilter.OnChanged = func(s string) { mon.historyFilter = s }
	mon.radioFilter.Horizontal = false

	mon.listHistory = widget.NewList(
		func() int { return len(mon.filteredHistory.Records) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel(strings.Repeat("X", 80))
			lbl.TextStyle.Bold = true
			return lbl
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			lbl := co.(*widget.Label)
			lbl.SetText(string(mon.filteredHistory.Records[lii]))
			mon.listHistory.SetItemHeight(lii, 25)
		},
	)

	go mon.listen()
	go mon.clearDeadSVs()

	mon.ExtendBaseWidget(mon)
	return mon
}

func (m *GPSMonitor) CreateRenderer() fyne.WidgetRenderer {

	// rLeft := canvas.NewRectangle(color.Transparent)
	// rLeft.StrokeWidth = 1
	// rLeft.StrokeColor = theme.Color(theme.ColorNameForeground)

	rRadar := canvas.NewRectangle(color.Transparent)
	rRadar.StrokeWidth = .5
	rRadar.StrokeColor = theme.Color(theme.ColorNameForeground)

	rArray := canvas.NewRectangle(color.Transparent)
	rArray.StrokeWidth = .5
	rArray.StrokeColor = theme.Color(theme.ColorNameForeground)

	rHistory := canvas.NewRectangle(color.Transparent)
	rHistory.StrokeWidth = .5
	rHistory.StrokeColor = theme.Color(theme.ColorNameForeground)

	array := container.NewAdaptiveGrid(16)

	for i, gsv := range m.gsvs {
		if i == 0 {
			continue
		}
		array.Add(gsv)
	}

	radar := m.radar
	c := container.NewBorder(
		nil,
		container.NewBorder(nil, nil, m.light, nil, m.status),
		container.NewStack(container.NewVBox(m.radioSource, &widget.Separator{}, &widget.Separator{}, &widget.Separator{}, m.radioFilter)),
		nil,
		container.NewBorder(
			container.NewStack(rArray, container.NewPadded(array)),
			nil,
			nil, nil,
			container.NewBorder(
				m.fixdisplay,
				nil, nil, nil, container.NewGridWithColumns(
					2,
					container.NewStack(rHistory, container.NewPadded(m.listHistory)),
					container.NewStack(rRadar, radar),
				),
			),
		),
	)

	return widget.NewSimpleRenderer(c)
}

func (m *GPSMonitor) filterHistory() {
	m.filteredHistory = *NewGPSHistory(500)
	// fmt.Println(m.filteredHistory.Records)
	for _, rec := range m.history.Records {
		if rec.Type() == m.historyFilter || m.historyFilter == "" {
			m.filteredHistory.Add(rec)
		}
	}
}

func (m *GPSMonitor) listen() {
	for record := range m.chanIn {
		m.light.Flash(50)
		m.history.Add(record)
		m.filterHistory()
		fyne.Do(m.listHistory.Refresh)
		fyne.Do(func() { m.listHistory.ScrollToBottom() })
		fields := record.Fields()

		var tempdata GSVData
		switch record.Type() {
		case "GSV":
			if record.Origin() != "GP" {
				continue
			}
			n := (len(fields) - 4) / 4
			for i := range n {
				prn := fields[4+4*i]
				index, err := stringToInt(prn)
				if err != nil {
					continue
				}
				(&tempdata).FromRecord(record, i)
				if tempdata.SNR == 0 {
					continue
				}
				m.gsvs[index].data.FromRecord(record, i)
				m.gsvs[index].formatInfo()
			}
			m.radar.sats = m.gsvs
			fyne.DoAndWait(func() { m.radar.Refresh() })
		case "TXT":
		case "RMC":
			if fields[2] == "V" {
				m.fixdisplay.SetVoidColours()
			} else {
				m.fixdisplay.SetActiveColours()
			}
			lat, err := stringToFloat(fields[3])
			if err != nil {
				continue
			}
			lon, err := stringToFloat(fields[5])
			if err != nil {
				continue
			}
			m.fixdisplay.fix.latitude, m.fixdisplay.fix.longitude = lat, lon

			t, err := time.Parse("150405.00", fields[1])
			if err == nil {
				m.fixdisplay.fix.time = t
			}
			fyne.DoAndWait(m.fixdisplay.Refresh)
		case "GGA":
			nsats, err := stringToInt(fields[7])
			if err == nil {
				m.fixdisplay.fix.nsats = nsats
			}
			hdop, err := stringToFloat(fields[8])
			if err == nil {
				m.fixdisplay.fix.HDOP = hdop
			}
			alti, err := stringToFloat(fields[9])
			if err == nil {
				m.fixdisplay.fix.altitude = alti
			}
			geoid, err := stringToFloat(fields[11])
			if err == nil {
				m.fixdisplay.fix.geoid_height = geoid
			}
		case "GSA":
			pdop, err := stringToFloat(fields[15])
			if err == nil {
				m.fixdisplay.fix.PDOP = pdop
			}
			hdop, err := stringToFloat(fields[16])
			if err == nil {
				m.fixdisplay.fix.HDOP = hdop
			}
			vdop, err := stringToFloat(fields[17])
			if err == nil {
				m.fixdisplay.fix.VDOP = vdop
			}
		case "VTG":
			cmg, err := stringToFloat(fields[1])
			if err == nil {
				m.fixdisplay.fix.CMG = cmg
			}
			spkt, err := stringToFloat(fields[5])
			if err == nil {
				m.fixdisplay.fix.speedKt = spkt
			}
			spkmh, err := stringToFloat(fields[7])
			if err == nil {
				m.fixdisplay.fix.speedKmh = spkmh
			}
			var s string
			switch fields[9] {
			case "A":
				s = "Autonomous Mode"
			case "D":
				s = "Differential Mode"
			case "E":
				s = "Dead Reckoning Mode"
			case "M":
				s = "Manual Imput"
			case "N":
				s = "Data Not Valid"
			}
			m.fixdisplay.fix.mode = s
		}
	}
}

// func (m *GPSMonitor) DMS(s string) (string, error) {
// 	if len(s) == 0 {
// 		return "00°00'00\"", errors.New("no data")
// 	}
// 	fields := strings.Split(s, ".")
// 	if len(fields[0]) == 4 {
// 		d, err := stringToInt(fields[0][:2])
// 		if err != nil {
// 			return "00°00'00\"", err
// 		}
// 		m, err := stringToFloat(s[2:])
// 		if err != nil {
// 			return "00°00'00\"", err
// 		}
// 		s := (m - math.Floor(m)) * 60
// 		return fmt.Sprintf("%02d°%02d'%02.1f\"", d, int(m), s), nil
// 	} else if len(fields[0]) == 5 {
// 		d, err := stringToInt(fields[0][:3])
// 		if err != nil {
// 			return "00°00'00\"", err
// 		}
// 		m, err := stringToFloat(s[3:])
// 		if err != nil {
// 			return "00°00'00\"", err
// 		}
// 		s := (m - math.Floor(m)) * 60
// 		return fmt.Sprintf("%02d°%02d'%02.1f\"", d, int(m), s), nil
// 	}
// 	return "00°00'00\"", errors.New("no data")

// }

func (m *GPSMonitor) useUSB() {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	newports := []string{}
	for _, p := range ports {
		if strings.Contains(p, "tty.usbmodem") {
			newports = append(newports, p)
		}
	}
	ports = newports
	if len(ports) == 0 {
		m.status.SetText("No serial ports found")
		return
	}

	for _, p := range ports {
		if !strings.Contains(p, "tty.usbmodem") {
			continue
		}
	}
	m.status.SetText(fmt.Sprintf("Opening port: %s", ports[0]))
	m.done = false

	go func() {
		mode := serial.Mode{
			BaudRate: 115200,
			DataBits: 8,
			Parity:   serial.NoParity,
			StopBits: serial.OneStopBit,
		}
		port, err := serial.Open(ports[0], &mode)
		if err != nil {
			log.Fatalln(err)
		}
		defer func() {
			port.Close()
			m.status.SetText(fmt.Sprintf("port %s closed", ports[0]))
		}()
		reader := bufio.NewReader(port)
		sc := bufio.NewScanner(reader)
		for sc.Scan() && !m.done {
			m.chanIn <- GPSRecord(sc.Text())
		}
		fmt.Println(sc.Err(), m.done)
	}()

}

func (m *GPSMonitor) useNetwork() {
}

func (m *GPSMonitor) useFile() {
	m.done = false
	f, err := zenity.SelectFile(
		zenity.FileFilters{zenity.FileFilter{
			Name:     "NMEA Files",
			Patterns: []string{"*.nmea", "*.txt"},
			CaseFold: true,
		}},
	)
	if err != nil {
		log.Fatalln(err)
	}
	reader, err := os.Open(f)
	if err != nil {
		return
	}
	sc := bufio.NewScanner(reader)
	go func() {
		for sc.Scan() && !m.done {
			m.chanIn <- GPSRecord(sc.Text())
			time.Sleep(time.Millisecond * 10)
		}
		fmt.Println(sc.Err())
	}()
}

func (m *GPSMonitor) useDummy() {
	m.done = false
	m.status.SetText("Using Dummy Data")
	records := strings.Split(smallString, "\n")
	index := 0
	go func() {
		for !m.done {
			if len(records[index]) < 10 {
				index = 0
			}
			m.chanIn <- GPSRecord(records[index])
			index++
			if index == len(records)-1 {
				index = 0
			}
			time.Sleep(time.Millisecond * 200)
		}
	}()
}

func (m *GPSMonitor) changeSource(sourceName string) {
	m.done = true
	if sourceName == "" {
		m.status.SetText("OFF")
		return
	}
	m.status.SetText(fmt.Sprintf("%s MODE", sourceName))
	// time.Sleep(time.Millisecond * 100)
	switch sourceName {
	case "NET":
		m.useNetwork()
	case "FILE":
		m.useFile()
	case "USB":
		m.useUSB()
	case "DUMMY":
		m.useDummy()
	}
}

func (m *GPSMonitor) clearDeadSVs() {
	for range m.tickClear.C {
		for _, v := range m.gsvs {
			if time.Since(v.data.lastupdate).Seconds() > m.TTL {
				v.clear()
				v.Refresh()
			}
		}
	}
}
