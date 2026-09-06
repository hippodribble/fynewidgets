package gps

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net"
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

	Nsats := 224

	mon := &GPSMonitor{

		radioSource:     widget.NewRadioGroup([]string{"NET", "USB", "FILE", "DUMMY"}, func(s string) {}),
		radioFilter:     widget.NewRadioGroup([]string{"GSV", "GLL", "GGA", "RMC", "GSA", "VTG", "ZDA"}, func(s string) {}),
		chanIn:          make(chan GPSRecord, 10),
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

	// go mon.Strobe()
	go mon.listen()
	mon.light.Off()
	go mon.clearDeadSVs()

	mon.ExtendBaseWidget(mon)
	return mon
}

func (m *GPSMonitor) CreateRenderer() fyne.WidgetRenderer {

	rRadar := canvas.NewRectangle(color.Transparent)
	rRadar.StrokeWidth = .5
	rRadar.StrokeColor = theme.Color(theme.ColorNameForeground)

	rArray := canvas.NewRectangle(color.Transparent)
	rArray.StrokeWidth = .5
	rArray.StrokeColor = theme.Color(theme.ColorNameForeground)

	rHistory := canvas.NewRectangle(color.Transparent)
	rHistory.StrokeWidth = .5
	rHistory.StrokeColor = theme.Color(theme.ColorNameForeground)

	array := container.NewAdaptiveGrid(32)

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
	for _, rec := range m.history.Records {
		if rec.Type() == m.historyFilter || m.historyFilter == "" {
			m.filteredHistory.Add(rec)
		}
	}
}

func (m *GPSMonitor) listen() {
	for record := range m.chanIn {
		m.light.On()
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

func (m *GPSMonitor) useUSB() {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err, ". Quitting.")
	}
	newports := []string{}
	for _, p := range ports {
		if strings.Contains(p, "cu.usbmodem") {
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
	fmt.Printf("Opening port: %s\n", ports[0])
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
			log.Println(err)
			return
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
	m.done = true
	go func() {
		conn, err := net.Dial("tcp", "localhost:2947")
		if err != nil {
			log.Fatalln("No server")
		}
		defer conn.Close()
		// fmt.Println("Set gpsd watch mode as NMEA")
		// fmt.Fprintln(conn, `?WATCH={"enable":true,"json":true}`)
		fmt.Fprintln(conn, `?WATCH={"enable":true,"nmea":true}`)

		// m.readGPSDJSON(conn)
		m.readGPSDNMEAOverTCP2947(conn)
	}()
}

func (m *GPSMonitor) readGPSDJSONOverTCP2947(conn net.Conn) {
	m.done = false
	fmt.Println("reading connection. done is", m.done)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Println(line)

		if m.done {
			return
		}

		fyne.Do(func() { m.light.Flash(30) })

		// Generic envelope to detect class:
		var env struct {
			Class string `json:"class"`
		}
		if err := json.Unmarshal([]byte(line), &env); err != nil {
			log.Printf("bad json: %v (%s)", err, line)
			continue
		}
		// fmt.Printf("Class: %s\n", env.Class)

		switch env.Class {
		case "TPV":
			var tpv tpv
			if err := json.Unmarshal([]byte(line), &tpv); err == nil {
				// fmt.Printf("TPV: %+v\n", tpv)
				if tpv.Lat == 0 && tpv.Lon == 0 {
					continue
				}
				fmt.Printf("Time: %v Lat: %.6f Lon: %.6f\n", tpv.Time, tpv.Lat, tpv.Lon)
			} else {
				fmt.Println("JSON error")
			}
		case "SKY":
			var sky sky
			if err := json.Unmarshal([]byte(line), &sky); err == nil {
				// fmt.Printf("SKY: %+v\n", sky)
			}
			// handle satellite info
			// other message types...
			for _, sat := range sky.Satellites {
				if sat.Ss > 1000 {
					fmt.Printf("Satellite PRN: %3d, GNSSID: %3d, SIGID: %3d SNR %.1f dB Hz \n", sat.PRN, sat.Gnssid, sat.Sigid, sat.Ss)
				}
				// fmt.Printf("Satellite PRN: %d, Az: %.2f, El: %.2f, SNR: %.2f, Health: %d, GNSSID: %d, SIGID: %d\n",
				// 	sat.PRN, sat.Az, sat.El, sat.Ss, sat.Health, sat.Gnssid, sat.Sigid)
			}
		default:
			// fmt.Printf("Other class: %s\n", env.Class)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("scanner error: %v", err)
	}
}

func (m *GPSMonitor) readGPSDNMEAOverTCP2947(conn2 net.Conn) {

	_, err := conn2.Write([]byte(`?WATCH={"enable":true,"nmea":true}` + "\n"))
	if err != nil {
		log.Fatalln("Error requesting stream")
	}
	m.done = false
	scanner := bufio.NewScanner(conn2)
	for scanner.Scan() && !m.done {
		line := scanner.Text()
		// if !strings.Contains(line,"RMC"){continue}
		// fmt.Println(line)
		// fyne.Do(func() { m.light.Flash(20) })
		m.chanIn <- GPSRecord(line)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("scanner error: %v", err)
	}
}

func (m *GPSMonitor) useFile() {
	m.done = true
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
	m.done = true
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
			time.Sleep(time.Millisecond * 800)
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
				fyne.Do(func() {
					v.clear()
					v.Refresh()
					m.light.Off()
				})
			}
		}
	}
}
