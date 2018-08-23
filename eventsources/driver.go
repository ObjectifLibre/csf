// Package actions provides types and interfaces necessary to write event sources
package eventsources

import (
	"sync"
	log "github.com/sirupsen/logrus"
	"github.com/ObjectifLibre/csf/configprovider"
)

// ArgType describes the different arguments and their types that
// an action needs as input. It is only used for humans so you can set arbitrary
// strings to describe the argument.
type ArgType struct {
	T string `json:"type"`
	N string `json:"name"`
}

// EventData represents an event. It is sent from a eventsource.
// Name is the name of the event and Data is an arbitrary map containing
// the event data.
type EventData struct {
	Name string
	Data map[string]interface{}
}

// EventSourceInterface is the interface event sources must implement.
// Setup() is called once when CSF starts and is used to configure and or
// init modules. Events() returns a map where Keys are events name and values
// are lists of parameters of the event.
type EventSourceInterface interface {
	Setup(ch chan EventData, config []byte) error
	Events() map[string][]ArgType
}

var sources = make(map[string]EventSourceInterface)
var sourcesM sync.RWMutex

// RegisterEventSource registers the event source. Must be called from the
// init() of the event sources.
func RegisterEventSource(name string, ds EventSourceInterface) {
	sourcesM.Lock()
	defer sourcesM.Unlock()
	sources[name] = ds
}

// SetupEventSources gets the configuration and calls the Setup() function
// of the event sources in dslist.
func SetupEventSources(ch chan EventData, cfp configprovider.ConfigProviderInterface, dslist []string) {
	sourcesM.RLock()
	defer sourcesM.RUnlock()
	log.Debug("Starting event sources setup")
	for _, name := range(dslist) {
		source, ok := sources[name]
		if !ok {
			log.WithFields(log.Fields{"eventsource": name}).Error("No such eventsource")
			continue
		}
		log.WithFields(log.Fields{"eventsource": name}).Debug("Eventsource setup started")
		go func(eventsource EventSourceInterface, dsname string) {
			config, err := cfp.GetEventSourceConfig(dsname)
			if err != nil {
				log.WithFields(log.Fields{"err": err, "eventsource": dsname}).Warn("Could not get config, using empty conf")
				config = make([]byte, 0)
			}
			if err = eventsource.Setup(ch, config); err != nil {
				log.WithFields(log.Fields{"err": err, "eventsource": dsname}).Error("Could not config eventsource")
				return
			}
			log.WithFields(log.Fields{"eventsource": dsname}).Debug("Event source set up")
		}(source, name)
	}
}

// GetEventSourcesList returns the list of all event sources.
func GetEventSourcesList() (sources []string)  {
	sources = make([]string, len(sources))
	for _, source := range(sources) {
		sources = append(sources, source)
	}
	return
}


// EventDesc describes an event. Used for the json API.
type EventDesc struct {
	Args []ArgType `json:"data"`
	Name string `json:"name"`
}

// Events is a list of events. Used for the json API.
type Events struct{
	Events []EventDesc `json:"events"`
}

// GetAllEvents returns a list of all events.
func GetAllEvents() (events Events) {
	for _, source := range(sources) {
		for key, value := range(source.Events()) {
			eventdesc := EventDesc{Args: value, Name: key}
			events.Events = append(events.Events, eventdesc)
		}
	}
	return
}
