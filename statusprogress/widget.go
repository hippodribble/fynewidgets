package statusprogress

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Message struct {
	Text     string // message text
	Duration int    // time to live in seconds
}

type StatusProgressWidget struct {
	widget.BaseWidget
	progressbar         *widget.ProgressBar
	infiniteprogressbar *widget.ProgressBarInfinite
	status              *widget.Label
	r                   *canvas.Rectangle
	commschannel        chan interface{}
	messagequeue        chan Message
	lastmessagetime     time.Time
}

func NewStatusProgress(channel chan interface{}) *StatusProgressWidget {
	w := &StatusProgressWidget{commschannel: channel}
	w.status = widget.NewLabel("")
	w.status.TextStyle.Bold = true

	w.progressbar = widget.NewProgressBar()
	w.infiniteprogressbar = widget.NewProgressBarInfinite()
	w.progressbar.Hide()
	w.infiniteprogressbar.Hide()

	w.messagequeue = make(chan Message)
	w.r = canvas.NewRectangle(color.Transparent)
	w.r.SetMinSize(fyne.NewSize(100, 1))
	// w.commschannel = make(chan interface{})
	go w.processMessages()
	go w.listen()
	// log.Println("running...")

	w.ExtendBaseWidget(w)
	return w
}

func (w *StatusProgressWidget) CreateRenderer() fyne.WidgetRenderer {
	log.Println("Renderer created")
	c := container.NewBorder(nil, nil, nil, container.NewStack(w.r, w.progressbar, w.infiniteprogressbar), w.status)
	return widget.NewSimpleRenderer(c)
}

func (w *StatusProgressWidget) listen() {
	for {
		select {
		case payload := <-w.commschannel:
			// log.Println("Message:", payload)
			if p, ok := payload.(Message); ok {
				// log.Println("format is of message type")
				w.status.SetText(p.Text)
				w.lastmessagetime = time.Now().Add(time.Duration(p.Duration) * time.Second)
				continue
			}
			if p,ok:=payload.(string);ok{
				w.status.SetText(p)
				continue
			}
			if p, ok := payload.(float64); ok {
				if p > 0 && p <= 1 {
					w.progressbar.Show()
					w.progressbar.SetValue(p)
				}
				if p == 0 {
					w.progressbar.Hide()
					w.progressbar.SetValue(p)
				}
				if p < 0 && p >= -1 {
					w.progressbar.Show()
					w.progressbar.SetValue(w.progressbar.Value - p)
				}
				if p > 1 {
					log.Println("Start infinite")
					w.infiniteprogressbar.Show()
					w.infiniteprogressbar.Start()
				}
				if p < -1 {
					log.Println("Stop infinite")
					w.infiniteprogressbar.Stop()
					w.infiniteprogressbar.Hide()
				}
				continue
			}

			log.Println("Unknown payload type:", payload)
			w.status.SetText(fmt.Sprintf("%v",payload))
		}
	}
}

func (w *StatusProgressWidget) processMessages() {
	go func() {
		for {
			time.Sleep(time.Millisecond * 300)
			// log.Println("Clock")
			if time.Now().After(w.lastmessagetime) {
				w.status.SetText("Ready...")
				w.lastmessagetime = time.Now().Add(time.Duration(60) * time.Second)
			}
		}
	}()
}

func (w *StatusProgressWidget) CurrentMessage() string {
	return w.status.Text
}

func (w *StatusProgressWidget) SetMessageDirectly(s string) {
	w.status.SetText(s)
}