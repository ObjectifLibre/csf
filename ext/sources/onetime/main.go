// Package onetime sends an event just one time at startup. Usefull for
// debuging purposes.
package onetime

import (
	"time"
	"github.com/ObjectifLibre/csf/eventsources"
)

func init() {
	eventsources.RegisterEventSource("onetime", &dummyEventSourceImplementation{})
}

var _ eventsources.EventSourceInterface = dummyEventSourceImplementation{}

type dummyEventSourceImplementation struct {}

func (dummy dummyEventSourceImplementation) Events() map[string][]eventsources.ArgType {
	Events := map[string][]eventsources.ArgType{
		"onetime": {{T: "timestamp string", N: "timestamp"}},
	}
	return Events
}

func (dummy dummyEventSourceImplementation) Setup(ch chan eventsources.EventData, cfg []byte) error {
	event := eventsources.EventData{
		Name: "onetime",
		Data: map[string]interface{}{
			"timestamp": time.Now().String(),
		},
	}
	ch <- event
	return nil
}
