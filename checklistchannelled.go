package fynewidgets

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// A channeled checklist uses a channel to communicate payloads.
// Payloads can signify individual changes to element selection,
// or indicate the status of all elements in the list at one.
// Selection of an item is represented by a positive index of the element.
// Deselection is represented by a negative index of the element.
// Status of the entire list is passed as a []bool
type CheckListChannelled struct {
	widget.BaseWidget
	channel chan any
	header  *widget.Label
	items   []*Status
	checks  []*widget.Check
	submit  *widget.Button
}

func NewCheckListChannelled(items []*Status, title string, channel chan any) *CheckListChannelled {

	clc := &CheckListChannelled{
		channel: channel,
		header:  widget.NewLabel(title),
		items:   items,
	}

	clc.header.TextStyle.Bold = true
	clc.checks = make([]*widget.Check, len(items))
	for i := range items {
		clc.checks[i] = widget.NewCheck(items[i].Text, func(b bool) {
			fmt.Println(i)
			clc.items[i].Selected=b
			channel <- clc.items[i]
		})
	}
	clc.submit = widget.NewButton("Submit", func() {
		bools := make([]bool, len(items))
		for i := range clc.checks {
			bools[i] = clc.checks[i].Checked
		}
		clc.channel <- bools

	})

	clc.ExtendBaseWidget(clc)
	return clc

}

func (c *CheckListChannelled) CreateRenderer() fyne.WidgetRenderer {
	list := container.NewVBox(c.header)
	bottom := container.NewBorder(nil, nil, nil, c.submit, nil)
	for _, k := range c.checks {
		list.Add(k)
	}
	con := container.NewBorder(
		nil,
		bottom,
		nil, nil,
		list,
	)
	return widget.NewSimpleRenderer(con)
}
