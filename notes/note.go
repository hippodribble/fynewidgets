package notes

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
)

// Stores basic Note data including title, content, updated time, and project, as well as a unique ID to speed up sorting, searcjing, etc.
// - there is a Label view to see the data
// - right-click changes the contect to editing, and the Label is replaced with an Entry widget to edit the content. Typing Escape will finish up
type Note struct {
	Title   string    `json:"title"`
	Content string    `json:"content,omitempty"`
	Updated time.Time `json:"updated"`
	Project string    `json:"project"`
	ID      uint64    `json:"id"`
	Tags    []string  `json:"tags,omitempty"`
}

func NewNote(title, content, project string) *Note {
	return &Note{
		Title:   title,
		Content: content,
		Updated: time.Now(),
		Project: project,
		ID:      rand.Uint64(),
	}
}

type NoteView struct {
	fynewidgets.DraggableBaseWidget
	Note         *Note
	title        *canvas.Text
	contentLabel *widget.Label
	contentEntry *SubmitMultiLineEntry
	status       *canvas.Text
	footer       *fyne.Container
	r            *canvas.Rectangle
	colour       color.Color
	chSelect     chan *NoteView
	chDelete     chan uint64
	boundContent binding.String
}

func NewNoteView(note *Note, c color.Color, ch chan *NoteView, ch2 chan uint64) *NoteView {
	d := binding.BindString(&note.Content)
	v := &NoteView{Note: note, colour: c, boundContent: d}
	v.chSelect = ch
	v.chDelete = ch2
	textcolour := theme.Color(theme.ColorNameForeground)
	v.title = canvas.NewText(note.Title, textcolour)
	v.title.TextSize = 18
	v.title.TextStyle.Bold = true
	v.title.TextStyle.Italic = true
	// v.title.Alignment = fyne.TextAlignCenter
	v.contentLabel = widget.NewLabelWithData(v.boundContent)
	v.contentLabel.Wrapping = fyne.TextWrapWord
	v.contentLabel.TextStyle.Italic=true
	v.contentEntry = NewSubmitMultiLineEntry()
	v.contentEntry.Bind(v.boundContent)
	v.contentEntry.Wrapping = fyne.TextWrapWord
	v.contentEntry.Hide()
	v.contentEntry.OnSubmit = func(s string) {
		fmt.Println(s)
		v.contentLabel.Refresh()
		v.contentEntry.Hide()
		v.contentLabel.Show()
	}
	v.status = canvas.NewText("", textcolour)
	v.status.TextSize = 9
	clock := canvas.NewText(note.Updated.Format("02-01-2006 15:04"), textcolour)
	clock.TextSize = 9
	clock.TextStyle.Italic = true
	v.footer = container.NewBorder(nil, nil, nil, clock, v.status)
	v.r = canvas.NewRectangle(v.colour)
	v.r.StrokeColor = textcolour
	v.r.StrokeWidth = 0.5
	v.r.CornerRadius = 5
	v.ExtendBaseWidget(v)
	return v
}

func (v *NoteView) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewBorder(v.title, v.footer, nil, nil, container.NewStack(container.NewVScroll(v.contentLabel), v.contentEntry))
	return widget.NewSimpleRenderer(container.NewStack(v.r, c))
}

func (v *NoteView) MouseDown(e *desktop.MouseEvent) {
	switch e.Button {
	case desktop.MouseButtonPrimary:
		v.chSelect <- v
		v.DraggableBaseWidget.MouseDown(e)
	case desktop.MouseButtonSecondary:
		v.chDelete <- v.Note.ID
	}
}

func (v *NoteView) DoubleTapped(e *fyne.PointEvent) {
	fmt.Println(e.Position)
	v.contentEntry.Show()
	v.contentLabel.Hide()
}

func SaveNotes(filename string, notes []Note) error {
	// Marshal to pretty JSON; use json.Marshal for compact output [web:3][web:33]
	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal notes: %w", err)
	}

	// 0644 is a typical permission mask for data files [web:35][web:37]
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("write file %s: %w", filename, err)
	}

	return nil
}

func LoadNotes(filename string) ([]Note, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filename, err)
	}

	var notes []Note
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, fmt.Errorf("unmarshal notes: %w", err)
	}

	return notes, nil
}

type FormEditor struct {
	widget.Form
	Note  *Note
	title binding.String
}

// a Store contains a slice of Note elements and enables access to tags, etc
type Store struct {
	Notes []*Note
}

func NewStore(notes []*Note) *Store {
	s := &Store{Notes: notes}
	return s
}

func (s *Store) Projects() []string {
	projects := make(map[string]int)
	for _, n := range s.Notes {
		projects[n.Project] = 1
	}
	keys := make([]string, 0, len(projects))
	for k := range projects {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func (s *Store) Tags() []string {
	tags := make(map[string]int)
	for _, n := range s.Notes {
		for _, v := range n.Tags {
			tags[v] = 1
		}
	}
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func (s *Store) AddNote(n *Note) {
	s.Notes = append(s.Notes, n)
}

func (s *Store) RemoveNote(ID uint64) {
	newnotes := []*Note{}
	for _, n := range s.Notes {
		if n.ID != ID {
			newnotes = append(newnotes, n)
		}
	}
	s.Notes = newnotes
}

func (s *Store) ShowEditor(chNewNote chan *Note) {
	boundProject := binding.NewString()
	boundTag := binding.NewString()
	boundTitle := binding.NewString()
	boundContent := binding.NewString()

	wTitle := widget.NewEntryWithData(boundTitle)
	wContent := widget.NewEntryWithData(boundContent)
	wContent.MultiLine=true
	wProjects := widget.NewSelectEntry(s.Projects())
	wProjects.Bind(boundProject)
	wTags := widget.NewSelectEntry(s.Tags())
	wTags.Bind(boundTag)

	formItems := []*widget.FormItem{
		widget.NewFormItem("Title", wTitle),
		widget.NewFormItem("Content", wContent),
		widget.NewFormItem("Project", wProjects),
		widget.NewFormItem("Tag", wTags),
	}

	dlg := dialog.NewForm("New Note", "OK", "Cancel", formItems, func(b bool) {
		if !b {
			return
		}
		fmt.Println("Confirm")
		proj, err := boundProject.Get()
		if err != nil {
			return
		}
		cont, err := boundContent.Get()
		if err != nil {
			return
		}
		title, err := boundTitle.Get()
		if err != nil {
			return
		}
		tag, err := boundTag.Get()
		if err != nil {
			return
		}
		n := NewNote(title, cont, proj)
		n.Tags = append(n.Tags, tag)
		fmt.Println(len(s.Notes))
		s.Notes = append(s.Notes, n)
		chNewNote <- n
		fmt.Println("made new note")
		fmt.Println(len(s.Notes))
	}, fyne.CurrentApp().Driver().AllWindows()[0])
	dlg.Resize(fyne.NewSize(800, 800))
	dlg.Show()
}

func (s *Store) CreateViews(selectChannel chan *NoteView, deleteChannel chan uint64) []*NoteView {
	projmap := make(map[string][]*Note)
	for _, n := range s.Notes {
		if projmap[n.Project] == nil {
			projmap[n.Project] = []*Note{}
		}
		projmap[n.Project] = append(projmap[n.Project], n)
	}
	colours := makeColours(len(projmap))
	views := []*NoteView{}
	projcount := 0
	for _, p := range projmap {
		for _, n := range p {
			views = append(views, NewNoteView(n, colours[projcount], selectChannel, deleteChannel))
		}
		projcount++
	}
	return views
}

func hsvColor(h, s, v float64) color.Color {
	// h: 0–360, s: 0–1, v: 0–1
	// simple HSV→RGB conversion
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

func makeColours(N int) color.Palette {
	p := color.Palette{}
	// p = append(p, color.RGBA{64, 64, 32, 255})
	// p = append(p, color.RGBA{64, 32, 64, 255})
	// p = append(p, color.RGBA{32, 64, 64, 255})
	// p = append(p, color.RGBA{64, 32, 32, 255})
	// p = append(p, color.RGBA{32, 64, 32, 255})
	// p = append(p, color.RGBA{32, 32, 64, 255})
	var s, v float64
	fmt.Println("Dark Mode:", isDarkMode())
	if isDarkMode() {
		v = 0.25
		s = 0.9
	} else {
		v = 0.9
		s = 0.1
	}
	// s = 0.5
	dh := 360.0 / float64(N)
	for h := 0.0; h < 360.0; h += dh {
		p = append(p, hsvColor(h, s, v))
	}
	return p
}

func isDarkMode() bool {
	return fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantDark
}
