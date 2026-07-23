package gps

import (
	"strings"
)

type GPSRecord string

func (g GPSRecord) Fields() []string {
	x := strings.Split(string(g), "*")[0]

	fields := strings.Split(x, ",")
	return fields
}

func (g GPSRecord) Type() string {
	if len(g) < 5 {
		return ""
	}
	return string(g)[3:6]
}

func (g GPSRecord) Origin() string {
	if len(g) < 5 {
		return ""
	}
	return string(g)[1:3]
}

type GPSHistory struct {
	Records  []GPSRecord
	Capacity int
}

func NewGPSHistory(capacity int) *GPSHistory {
	h := &GPSHistory{
		Records:  make([]GPSRecord, capacity),
		Capacity: capacity,
	}
	return h
}

func (h *GPSHistory) Add(r GPSRecord) {
	h.Records = append(h.Records[1:], r)
}
