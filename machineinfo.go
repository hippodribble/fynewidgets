package fynewidgets

import (
	"fmt"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	eventbus "github.com/dtomasi/go-event-bus/v3"
)

type MachineInfo struct {
	widget.BaseWidget
	// clock, memory, goroutines *widget.Label
	allinonelabel *widget.Label
	bus           *eventbus.EventBus
}

func NewMachineInfo(bus *eventbus.EventBus) *MachineInfo {
	m := &MachineInfo{bus: bus}

	malloc := runtime.MemStats{}
	runtime.ReadMemStats(&malloc)

	m.allinonelabel = widget.NewLabel(fmt.Sprintf("|Heap%5dMB|Stack%5dMB|%3d procs|%s", malloc.HeapInuse/1024/1024, malloc.StackInuse/1024/1024, runtime.NumGoroutine(), time.Now().Format("Mon 15:04")))
	m.allinonelabel.TextStyle.Bold = true
	m.allinonelabel.TextStyle.Monospace = true
	m.allinonelabel.Importance=widget.MediumImportance

	go func() {
		for range time.NewTicker(time.Second * 1).C {
			runtime.ReadMemStats(&malloc)
			m.allinonelabel.SetText(fmt.Sprintf("|Heap%5dMB|Stack%5dMB|%3d procs|%s", malloc.HeapInuse/1024/1024, malloc.StackInuse/1024/1024, runtime.NumGoroutine(), time.Now().Format("Mon 15:04")))
		}
	}()

	m.ExtendBaseWidget(m)
	return m
}

func (m *MachineInfo) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(m.allinonelabel)
}
