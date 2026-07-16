package notes

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type SubmitMultiLineEntry struct {
	widget.Entry
	OnSubmit func(text string) // callback when editing is finished.
}

func NewSubmitMultiLineEntry() *SubmitMultiLineEntry {
	e := &SubmitMultiLineEntry{}
	e.MultiLine = true // or use widget.NewMultiLineEntry and embed that. [web:4][web:43]
	e.ExtendBaseWidget(e)
	return e
}

func (e *SubmitMultiLineEntry) TypedKey(key *fyne.KeyEvent) {
	var desktopDrv desktop.Driver
	if key.Name == fyne.KeyReturn { // Enter pressed. [web:39]

		if drv, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok { // get desktop driver. [web:46]
			desktopDrv = drv
		}

		if desktopDrv != nil {
			mods := desktopDrv.CurrentKeyModifiers() // which modifiers are held now. [web:46][web:49]
			if (mods & fyne.KeyModifierShift) != 0 {
				// fmt.Println("Shift Enter?")
				if e.OnSubmit != nil {
					e.OnSubmit(e.Text)
				}
				return
			} else {
				// fmt.Println("Enter?")
				e.Entry.TypedKey(key)
			}
		}

		// Optionally remove focus or swap back to a Label.
		// e.Disable() // or e.FocusLost(), or notify parent.
		return
	}
	if key.Name == fyne.KeyEscape {
		if e.OnSubmit != nil {
			e.OnSubmit(e.Text)
		}
		return
	}

	// For other keys, delegate back to the original Entry behaviour. [web:39]
	e.Entry.TypedKey(key)
}

func (e *SubmitMultiLineEntry)MinSize()fyne.Size{
	return fyne.NewSize(100,200)
}