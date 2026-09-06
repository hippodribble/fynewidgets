package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	ap := app.New()
	w := ap.NewWindow("Time Picker Test")
	w.SetContent(gui())
	w.Resize(fyne.NewSize(800, 800))
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {
	e := newTimeEntry()
	status := widget.NewLabel("ready...")
	return container.NewBorder(e, status, nil, nil,nil)
}

func requiredTimeValidator(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("time is required; use HH:MM")
	}

	if _, err := time.Parse("15:04", s); err != nil {
		return errors.New("enter a valid 24-hour time, e.g. 09:30")
	}
	return nil
}

func newTimeEntry() *widget.Entry {
	e := widget.NewEntry()
	e.PlaceHolder = "HH:MM (e.g. 09:30)"
	e.Validator = requiredTimeValidator
	e.AlwaysShowValidationError = true
	e.OnSubmitted = func(s string) {

		if err := e.Validate(); err != nil {
			return // show a dialog, keep submit disabled, etc.
		}

		parsedTime, err := time.Parse("15:04", strings.TrimSpace(e.Text))
		if err != nil {
			return // defensive: should match validator outcome
		}

		// parsedTime has date zero-value components and UTC location.
		// Use parsedTime.Hour() and parsedTime.Minute().
		hour, minute := parsedTime.Hour(), parsedTime.Minute()
		fmt.Printf("%02d:%02d", hour, minute)

	}
	return e
}
