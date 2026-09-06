package gps

import "time"

type Satellite struct {
	PRN    int     `json:"PRN"`
	Az     float64 `json:"az"`
	El     float64 `json:"el"`
	Ss     float64 `json:"ss"`     // SNR (dBHz)
	Health int     `json:"health"` // 0=unknown, 1=OK, 2=unhealthy
	Gnssid int     `json:"gnssid"` // GNSS ID
	Sigid  int     `json:"sigid"`  // signal ID
}

type sky struct {
	Class      string      `json:"class"`
	Tag        string      `json:"tag"`
	Device     string      `json:"device"`
	Time       time.Time   `json:"time"`
	Xdop       float64     `json:"xdop"`
	Ydop       float64     `json:"ydop"`
	Vdop       float64     `json:"vdop"`
	Tdop       float64     `json:"tdop"`
	Hdop       float64     `json:"hdop"`
	Pdop       float64     `json:"pdop"`
	Gdop       float64     `json:"gdop"`
	Satellites []Satellite `json:"satellites"`
}

type tpv struct {
	Class string  `json:"class"`
	Mode  int     `json:"mode"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Time  string  `json:"time"`
	// add fields as needed
}
