package fynewidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// EditableLabel is a widget that displays a label that can be edited when clicked.
// It is a combination of a label and an entry widget, where the label is shown by
// default and the entry is shown when the label is clicked.
type EditableLabel struct {
	widget.BaseWidget
	edit                    *EntryWithFocus
	label                   *widget.Label
	frame                   *fyne.Container
	scrollEdit, scrollLabel *container.Scroll
}

func NewEditableLabel(text string) *EditableLabel {
	el := &EditableLabel{}
	el.edit = NewEntryWithFocus()
	el.edit.SetText(text)
	el.edit.OnSubmitted = func(s string) { el.ViewMode() }
	el.edit.OnFocusLost = func(text string) { el.ViewMode() }
	el.edit.OnFocusGained = func() { el.EditMode() }

	el.label = widget.NewLabel(text)
	el.label.Wrapping=fyne.TextWrapBreak
	el.scrollEdit = container.NewVScroll(el.edit)
	el.scrollLabel = container.NewVScroll(el.label)
	el.scrollEdit.Hide()
	el.scrollLabel.Show()

	el.frame = container.NewStack(el.scrollEdit, el.scrollLabel)
	el.ExtendBaseWidget(el)
	return el
}

func (l *EditableLabel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(l.frame)
}

// EditMode switches the widget to edit mode, showing the entry field and hiding the label.
func (l *EditableLabel) EditMode() {
	l.scrollLabel.Hide()
	l.scrollEdit.Show()
	canvas := fyne.CurrentApp().Driver().CanvasForObject(l.edit)
	if canvas != nil {
		canvas.Focus(l.edit)
	}
}

// ViewMode switches the widget to view mode, showing the label and hiding the entry field.
func (l *EditableLabel) ViewMode() {
	l.label.Text = l.edit.Text
	l.scrollLabel.Show()
	l.scrollEdit.Hide()
}

func (l *EditableLabel) MouseDown(ev *desktop.MouseEvent) {
}

// MouseUp switches the widget to edit mode when the mouse is released over the label.
func (l *EditableLabel) MouseUp(ev *desktop.MouseEvent) {
	l.EditMode()
	canvas := fyne.CurrentApp().Driver().CanvasForObject(l.edit)
	if canvas != nil {
		canvas.Focus(l.edit)
	}
}

// EntryWithFocus is a custom entry widget that provides callbacks for focus gained and lost events.
type EntryWithFocus struct {
	widget.Entry
	OnFocusLost   func(text string)
	OnFocusGained func()
}

func NewEntryWithFocus() *EntryWithFocus {
	e := &EntryWithFocus{}
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapWord
	e.ExtendBaseWidget(e)
	return e
}

func (e *EntryWithFocus) FocusLost() {
	e.Entry.FocusLost()

	if e.OnFocusLost != nil {
		e.OnFocusLost(e.Text)
	}
}

func (e *EntryWithFocus) FocusGained() {
	e.Entry.FocusGained()

	if e.OnFocusGained != nil {
		e.OnFocusGained()
	}
}
