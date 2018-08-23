// Package openstack connects to Rabbitmq via AMQP and listen for compute
// related notifications.
package openstack

import (
	"encoding/json"
	"strings"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"

	"github.com/ObjectifLibre/csf/eventsources"
)

func init() {
	eventsources.RegisterEventSource("openstack", &openStackImplementation{})
}

var _ eventsources.EventSourceInterface = openStackImplementation{}

type openstackConfig struct {
	RabbitUrl string `yaml:"rabbit_url"`
}

var cfg openstackConfig

type openStackImplementation struct {}

func (dummy openStackImplementation) Events() map[string][]eventsources.ArgType {
	Events := map[string][]eventsources.ArgType{
		"compute_event": {{T: "instance object", N: "instance"}},
	}
	return Events
}


func listenForEvents(events chan eventsources.EventData) error {
	conn, err := amqp.Dial(cfg.RabbitUrl)
	if err != nil {
		return fmt.Errorf("Could not dial Rabbit: %s", err)
	}
//	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Could not get channel: %s", err)
	}

//	defer ch.Close()

	msgs, err := ch.Consume(
		"versioned_notifications.info", // queue
		"",     // consumer
		false,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("Could not consume channel: %s", err)
	}

	startTime := time.Now()

	go func(events chan eventsources.EventData) {
		for d := range msgs {
			event := map[string]interface{}{}
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.WithFields(log.Fields{"err": err,
					"eventsource": "openstack"}).Warn("Received bad notification")
				continue
			}
			eventjson := []byte(event["oslo.message"].(string))
			if err := json.Unmarshal(eventjson, &event); err != nil {
				log.WithFields(log.Fields{"err": err,
					"eventsource": "openstack"}).Warn("Received bad notification")
				continue
			}
			eventType := event["event_type"].(string)
			if !strings.HasPrefix(eventType, "instance.") {
				log.WithFields(log.Fields{"event_type": eventType,
					"eventsource": "openstack"}).Debug("Dropped notification")
				continue
			} else if timestampstr, ok := event["timestamp"].(string); !ok {
				log.WithFields(log.Fields{"err": "Could not parse timestamp, not a string",
					"eventsource": "openstack"}).Warn("Received bad notification")
				continue
			} else if timestamp, err := time.Parse("2006-01-02 15:04:05.999999", timestampstr); err != nil {
				log.WithFields(log.Fields{"err": err,
					"eventsource": "openstack"}).Warn("Received bad timestamp in notification")
				continue
			} else if delta := startTime.Sub(timestamp); delta > time.Minute * 5 {
				log.WithFields(log.Fields{"timestamp": timestamp,
					"eventsource": "openstack"}).Debug("Dropped old notification")
				continue
			}
			eventData := eventsources.EventData{
				Name: "compute_event",
				Data: event,
			}
			events <- eventData
		}
	}(events)
	return nil
}

func (dummy openStackImplementation) Setup(ch chan eventsources.EventData, rawcfg []byte) error {
	if err := yaml.Unmarshal(rawcfg, &cfg); err != nil {
		return fmt.Errorf("Bad config: %s", err)
	}
	if err := listenForEvents(ch); err != nil {
		return err
	}
	return nil
}
