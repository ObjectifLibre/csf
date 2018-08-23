// Package dummysource implements a dummy event source for reference and debuging
// purposes.
package dummysource

import (
	"fmt"
	"time"
	"math/rand"
	"strconv"
	"github.com/ObjectifLibre/csf/eventsources"
)

func init() {
	eventsources.RegisterEventSource("dummy", &dummyEventSourceImplementation{})
}

var _ eventsources.EventSourceInterface = dummyEventSourceImplementation{}

type dummyEventSourceImplementation struct {}

// Events returns 2 dummy events filled with random data.
func (dummy dummyEventSourceImplementation) Events() map[string][]eventsources.ArgType {
	Events := map[string][]eventsources.ArgType{
		"dummy_event": {{T: "int", N: "dummy_id"}, {T: "string", N: "dummy_date"}, {T: "string", N: "dummy_host"}},
		"dummy_event_2": {{T: "int", N: "dummy_id"}, {T: "string", N: "dummy_date"}, {T: "string", N: "dummy_string"}},
	}
	return Events
}

// randomDummyEventGenerator generates random event every 2 seconds.
func randomDummyEventGenerator(ch chan eventsources.EventData) {
	for {
		time.Sleep(time.Duration((rand.Int() % 10000)) * time.Millisecond)
		fmt.Println("Generating new event...")
		if rand.Int() % 2 == 0 {
			event := eventsources.EventData{
				Name: "dummy_event",
				Data: map[string]interface{}{
					"dummy_id": string(strconv.Itoa(rand.Int())),
					"dummy_date": time.Now().UTC().Format(time.RFC3339),
					"dummy_host": string(strconv.Itoa(rand.Int())),
				},
			}
			ch <- event
		} else {
			event := eventsources.EventData{
				Name: "dummy_event_2",
				Data: map[string]interface{}{
					"dummy_id": string(strconv.Itoa(rand.Int())),
					"dummy_date": time.Now().UTC().Format(time.RFC3339),
					"dummy_string": string(strconv.Itoa(rand.Int())),
				},
			}
			ch <- event
		}
	}
}

// Setup only prints the configuration it receives and launches the
// randomDummyEventGenerator.
func (dummy dummyEventSourceImplementation) Setup(ch chan eventsources.EventData, cfg []byte) error {
	fmt.Println("Dummy event generator set up with config: " + string(cfg))
	go randomDummyEventGenerator(ch)
	return nil
}
