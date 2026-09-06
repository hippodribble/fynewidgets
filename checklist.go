package fynewidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Status struct {
	Text     string
	Selected bool
}

// A Checklist takes a set of titles and maintains an equivalent slice of booleans indicating selection status for each item
type CheckList struct {
	widget.BaseWidget
	title  *widget.Label
	checks []*widget.Check
}

func NewCheckList(title string, items []*Status) *CheckList {
	all := make([]*widget.Check, len(items))
	for i := range all {
		all[i] = widget.NewCheck(items[i].Text, func(b bool) {})
	}
	l := &CheckList{
		title:  widget.NewLabel(title),
		checks: all,
	}
	l.ExtendBaseWidget(l)
	return l
}

func (l *CheckList) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewVBox()
	for _, k := range l.checks {
		c.Add(k)
	}
	b := container.NewBorder(
		l.title,
		nil, nil, nil,
		c,
	)
	return widget.NewSimpleRenderer(b)
}

func (l *CheckList) Selected() []bool {
	b := make([]bool, len(l.checks))
	for i := range l.checks {
		b[i] = l.checks[i].Checked
	}
	return b
}
