package main

import (
	"fmt"
	"log"
	"math/rand"

	eventbus "github.com/dtomasi/go-event-bus/v3"
)

var bus *eventbus.EventBus = eventbus.NewEventBus()

func main() {
	talkers := []talker{}
	for i := 0; i < 3; i++ {
		newtalker := *newTalker(fmt.Sprintf("Unit %d", i+1), bus)
		talkchan := newtalker.bus.Subscribe("notify")
		go func() {
			for x := range talkchan {
				log.Println(newtalker.name, x.Topic, x.Data)
				x.Done()
			}
		}()
		talkers = append(talkers, newtalker)
	}
	for i := 0; i < 10; i++ {
		log.Println("Phase", i+1)

		rn := rand.Intn(3)
		talkers[rn].say("hello")
	}
}

type talker struct {
	name string
	bus  *eventbus.EventBus
}

func newTalker(name string, bus *eventbus.EventBus) *talker {
	return &talker{name: name, bus: bus}
}

func (t *talker) say(msg string) {
	bus.Publish("notify", t.name+" says: "+msg)
}
