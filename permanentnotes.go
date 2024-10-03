package fynewidgets

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	eventbus "github.com/dtomasi/go-event-bus/v3"
)

type PermanentNotes struct {
	widget.BaseWidget
	Notes         []string
	filterednotes []string
	boundnotes    binding.StringList
	list          *widget.List
	lastfolder    fyne.ListableURI
	bus           *eventbus.EventBus
}

func NewPermanentNotes(oldnotes []string, bus *eventbus.EventBus) *PermanentNotes {

	notes := &PermanentNotes{Notes: oldnotes, bus: bus}
	notes.ExtendBaseWidget(notes)
	notes.filterednotes = make([]string, 0)
	notes.filterednotes = append(notes.filterednotes, notes.Notes...)

	slices.SortFunc(notes.filterednotes, func(a, b string) int {
		if a == b {
			return 0
		}
		if a < b {
			return 1
		}
		return -1
	})

	notes.boundnotes = binding.BindStringList(&notes.filterednotes)
	notes.list = widget.NewListWithData(notes.boundnotes,
		func() fyne.CanvasObject { return widget.NewLabel(strings.Repeat("X", 80)) },
		func(di binding.DataItem, co fyne.CanvasObject) {
			if co, ok := co.(*widget.Label); ok {
				co.Bind(di.(binding.String))
			} else {
				return
			}
		},
	)

	return notes

}



func (p *PermanentNotes) CreateRenderer() fyne.WidgetRenderer {

	filter := widget.NewEntry()
	filter.PlaceHolder = "Enter a search term to see matching notes, or enter a note and press enter to save it"

	filter.OnSubmitted = func(s string) {
		p.Add(s)
		p.filterednotes = []string{}
		p.filterednotes = append(p.filterednotes, p.Notes...)
		filter.SetText("")
	}

	filter.OnChanged = func(s string) {
		t := filter.Text
		p.filterednotes = []string{}
		for _, n := range p.Notes {
			if strings.Contains(strings.ToLower(n), strings.ToLower(t)) {
				p.filterednotes = append(p.filterednotes, n)
			}
		}
		p.boundnotes.Set(p.filterednotes)
	}

	c := container.NewBorder(container.NewBorder(nil, nil, widget.NewLabel("Search or Add"), nil, filter), nil, nil, nil, p.list)
	return widget.NewSimpleRenderer(c)
}

func (p *PermanentNotes) Add(note string) {

	if len(note) == 0 {
		return
	}

	p.Notes = append(p.Notes, fmt.Sprintf("%s - %s\n", time.Now().Format("20060102 150405"), note))
	p.filterednotes = []string{}
	p.filterednotes = append(p.filterednotes, p.Notes...)
	p.boundnotes.Set(p.filterednotes)

	p.bus.Publish("stringlist:notes", p.Notes)

}

func (p *PermanentNotes) Refresh() {

	p.BaseWidget.Refresh()

}

func (p *PermanentNotes) ExportNotes() {

	dlg := dialog.NewFolderOpen(func(uc fyne.ListableURI, err error) {

		if err != nil {
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		if uc == nil {
			return
		}

		name := fmt.Sprintf("notes-%s.txt", time.Now().Format("20060102150405"))
		name = strings.Replace(name, ":", "-", -1)

		fileURI, err := storage.Child(uc, name)
		if err != nil {
			dialog.NewError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		w, err := storage.Writer(fileURI)
		if err != nil {
			dialog.NewError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		defer w.Close()
		for _, note := range p.Notes {
			w.Write([]byte(note))
		}

		p.lastfolder = uc

	}, fyne.CurrentApp().Driver().AllWindows()[0])

	dlg.Resize(fyne.NewSize(800, 600))
	dlg.SetLocation(p.lastfolder)
	dlg.Show()

}
