package gps

import (
	"time"
)

type GSVData struct {
	PRN                     string
	Elevation, Azimuth, SNR int
	lastupdate              time.Time
}

func (d *GSVData) FromRecord(record GPSRecord, index int) {
	fields := record.Fields()
	// fmt.Println(index,len(fields),4+4*index)
	var err error
	d.PRN = fields[4+4*index]
	d.Elevation, err = stringToInt(fields[5+4*index])
	if err != nil {
		d.Elevation = 0
	}
	d.Azimuth, err = stringToInt(fields[6+4*index])
	if err != nil {
		d.Azimuth = 0
	}
	d.SNR, err = stringToInt(fields[7+4*index])
	if err != nil {
		d.SNR = 0
	}
	d.lastupdate = time.Now()
}
