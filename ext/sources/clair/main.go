// Package clair listen for clair notifications and fetch vulnerabities details
// from clair to generate events. An external clair server is requiered.
package clair

import (
	"fmt"
	"encoding/json"
	"net/http"
	"net/url"
	log "github.com/sirupsen/logrus"
	"github.com/ObjectifLibre/csf/eventsources"
	"gopkg.in/yaml.v2"
)

func init() {
	eventsources.RegisterEventSource("Clair", &clairEventSourceImplementation{})
}

var _ eventsources.EventSourceInterface = clairEventSourceImplementation{}

type clairEventSourceImplementation struct {}

func (clair clairEventSourceImplementation) Events() map[string][]eventsources.ArgType {
	Events := map[string][]eventsources.ArgType{
		"new_vuln": {{T: "Object containing new vuln details", N: "new"}, {T: "Object containing old vuln details", N: "old"}},
	}
	return Events
}

type clairNotificationHandler struct {
}


type clairConfig struct {
	ch chan eventsources.EventData
	path *url.URL
}

type clairConfigFromYaml struct {
	Notifurl string `yaml:"notification_endpoint"`
	Webhook string `yaml:"webhook_listen_addr"`
}

var config = clairConfig{}

func getVulnPage(notification string, page string) (map[string]interface{}, error) {
	values := url.Values{}
	values.Set("limit", "1")
	if len(page) > 0 {
		values.Set("page", page)
	}
	ntf_url := config.path.String() + "/" + notification + "?" + values.Encode()
	resp, err := http.Get(ntf_url)
	if err != nil {
		log.WithFields(log.Fields{"url": ntf_url,
			"err": err}).Warn("Could not get clair notification")
		return nil, fmt.Errorf("Could not get clair notifications: %s", err)
	}
	log.WithFields(log.Fields{"notification": notification,
		"url": ntf_url,
		"response": resp.Status}).Debug("Fetched clair vuln info")
	var vuln map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&vuln); err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("Could not decode clair response")
		return nil, fmt.Errorf("Could not decode: %s", err)
	}
	return vuln, nil
}

func getVulnerabilities(notification string) error {
	page := ""
	for {
		vuln, err := getVulnPage(notification, page)
		if err != nil {
			return err
		}
		new, oknew := vuln["notification"].(map[string]interface{})["new"]
		old, okold := vuln["notification"].(map[string]interface{})["old"]
		if !oknew  {
			err := fmt.Errorf("Malformed clair json response")
			log.WithFields(log.Fields{"notification": notification,
			"err": err, "data": vuln}).Warn("Could not get vulnerability")
			return err
		}
		if !okold {
			old = nil
		}
		event := eventsources.EventData{Name: "new_vuln",
			Data: map[string]interface{}{
				"old": old,
				"new": new,
			}}
		config.ch <- event
		next, ok := vuln["notification"].(map[string]interface{})["nextPage"]
		if !ok {
			break
		} else {
			page = next.(string)
		}
	}
	return nil
}

func (m *clairNotificationHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{"eventsource": "clair"}).Debug("Received request on webhook")
	var params map[string]interface{}
	json.NewDecoder(req.Body).Decode(&params)
	id, ok := params["Notification"].(map[string]interface{})
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	notification, ok := id["Name"].(string)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	log.WithFields(log.Fields{"eventsource": "clair",
		"notification": notification}).Debug("Received notification from clair")
	go getVulnerabilities(notification)
}

func listenForClairNotifications(host string) {
	http.ListenAndServe(host, &clairNotificationHandler{})
}

func (clair clairEventSourceImplementation) Setup(ch chan eventsources.EventData, rawcfg []byte) error {
	var cfg clairConfigFromYaml
	if err := yaml.Unmarshal(rawcfg, &cfg); err != nil {
		return fmt.Errorf("Bad config: %s", err)
	}
	clairUrl , err := url.Parse(cfg.Notifurl)
	if err != nil {
		return fmt.Errorf("Bad notification url: %s", err)
	}
	config.path = clairUrl
	config.ch = ch
	go listenForClairNotifications(cfg.Webhook)
	return nil
}

