package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"github.com/hippodribble/fynewidgets/notes"
)

const NOTES_PREF_KEY = "notes-json"

var selectChannel chan *notes.NoteView
var deleteChannel chan uint64
var addChannel chan *notes.Note
var c *fyne.Container
var w fyne.Window
var pad float32 = 5
var s notes.Store

func main() {
	// netstuff()
	selectChannel = make(chan *notes.NoteView)
	deleteChannel = make(chan uint64)
	addChannel = make(chan *notes.Note)
	ap := app.NewWithID("com.github.hippodribble.fynewidgets.notes")
	w = ap.NewWindow("Test Button")
	w.Resize(fyne.NewSize(1200, 1200))
	w.SetFullScreen(true)
	fmt.Println(w.Canvas().Size())
	w.SetFullScreen(false)
	w.SetContent(gui())
	w.Canvas().SetOnTypedKey(handleKeys)
	handleKeys(&fyne.KeyEvent{Name: "L"})
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {

	c = container.NewWithoutLayout()

	go func() {
		for n := range selectChannel {
			found := -1
			// fmt.Println(&n, "from channel", n.Note.ID)
			for i, o := range c.Objects {
				if p, ok := o.(*notes.NoteView); ok {
					if p.Note.ID == n.Note.ID {
						found = i
					}
				}
			}
			if found < 0 {
				continue
			}
			x := []fyne.CanvasObject{}
			for i := range c.Objects {
				if i != found {
					x = append(x, c.Objects[i])
				}
			}
			x = append(x, c.Objects[found])
			c.Objects = x
			fyne.Do(func() { c.Refresh() })
		}
	}()

	go func() {
		for range addChannel {
			fyne.Do(func() {
				fmt.Println(len(s.Notes))
				views := s.CreateViews(selectChannel, deleteChannel)
				c.Objects = c.Objects[:0]
				for _, v := range views {
					c.Add(v)
				}
				SortByProject(c.Objects)
				SaveNotesPref(fyne.CurrentApp(), s.Notes, NOTES_PREF_KEY)
			})
		}
	}()

	go func() {
		for id := range deleteChannel {
			fmt.Println(id)
			s.RemoveNote(id)
			fyne.Do(func() {
				views := s.CreateViews(selectChannel, deleteChannel)
				c.Objects = c.Objects[:0]
				for _, v := range views {
					c.Add(v)
				}
				SortByProject(c.Objects)
			})
		}
	}()
	return c
}

func handleKeys(k *fyne.KeyEvent) {
	sz := w.Canvas().Size()
	switch k.Name {
	case "A", "=":
		s.ShowEditor(addChannel)
	case "F":
		fmt.Println("Find")
	case "L":
		loadStoreFromJSON()
		views := s.CreateViews(selectChannel, deleteChannel)
		c.Objects = c.Objects[:0]
		for _, v := range views {
			c.Add(v)
		}
		SortByProject(c.Objects)
	case "P":
		SortByProject(c.Objects)
	case "Q":
		fmt.Println("Quit")
		quitConfirm()
	case "R":
		fmt.Println("Revert")
	case "S":
		fmt.Println("Saving...")
		SaveNotesPref(fyne.CurrentApp(), s.Notes, NOTES_PREF_KEY)

	case "T":
		fmt.Print(len(c.Objects), len(s.Notes))
		s.CreateViews(selectChannel, deleteChannel)
		fmt.Print(len(c.Objects), len(s.Notes))
		// fmt.Println("Tile", len(c.Objects))
		cols := int(math.Sqrt(float64(len(c.Objects)))) + 1
		rows := len(c.Objects)/cols + 1
		W := sz.Width / float32(cols)
		H := sz.Height / float32(rows)
		for i, o := range c.Objects {
			row := i % cols
			col := i / cols
			o.Resize(fyne.NewSize(W-pad, H-pad))
			o.Move(fyne.NewPos(W*float32(row), H*float32(col)))
		}
		w.Canvas().Refresh(c)
	}
}

func SortByProject(obs []fyne.CanvasObject) {
	projects := make(map[string][]fyne.CanvasObject)
	for i := range obs {
		if v, ok := obs[i].(*notes.NoteView); ok {
			if projects[v.Note.Project] == nil {
				projects[v.Note.Project] = []fyne.CanvasObject{v}
			} else {
				projects[v.Note.Project] = append(projects[v.Note.Project], v)
			}
		}
	}

	keys := make([]string, 0, len(projects))
	for k := range projects {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	sz := w.Canvas().Size()

	maxn := 0
	for _, v := range projects {
		maxn = max(maxn, len(v))
	}

	H := (sz.Height) / float32(maxn)
	W := (sz.Width) / float32(len(projects))
	fmt.Println(W, H)
	var col float32
	for _, k := range keys {
		v := projects[k]
		// height depends on count of items in project.
		// spacing is fixed at 50
		var dy float32 = 100.0
		// fmt.Println(k, len(v))
		H := (sz.Height - 10) - float32(len(v)-1)*dy
		x := col * W
		nsz := fyne.NewSize(W-pad, H-pad)

		for i, o := range v {
			o.Resize(nsz)
			y := float32(i) * dy
			o.Move(fyne.NewPos(x, y))
		}
		col++
	}
	w.Canvas().Refresh(c)
}

func SaveNotesPref(a fyne.App, notes []*notes.Note, notesPrefKey string) error {
	data, err := json.Marshal(notes)
	if err != nil {
		return fmt.Errorf("marshal notes: %w", err)
	}
	a.Preferences().SetString(notesPrefKey, string(data)) // key/value stored on disk [web:38]
	return nil
}

func LoadNotesPref(a fyne.App, notesPrefKey string) ([]*notes.Note, error) {
	pref := a.Preferences()
	jsonStr := pref.String(notesPrefKey) // empty string if not set [web:38]

	if jsonStr == "" {
		return []*notes.Note{}, nil // no notes yet
	}

	var notes []*notes.Note
	if err := json.Unmarshal([]byte(jsonStr), &notes); err != nil {
		return nil, fmt.Errorf("unmarshal notes: %w", err)
	}

	return notes, nil
}

func loadStoreFromJSON() {

	loadednotes, err := LoadNotesPref(fyne.CurrentApp(), NOTES_PREF_KEY)
	if err != nil {
		log.Fatalln(err)
	}
	s = *notes.NewStore(loadednotes)
}

func quitConfirm() {
	dialog.ShowConfirm("Exit?", "Do you want to stop doing this\nand do something else instead?", func(b bool) {
		if b {
			fyne.CurrentApp().Quit()
		}
	}, w)
}