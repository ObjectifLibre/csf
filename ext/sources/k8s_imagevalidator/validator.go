// Package k8s_imagevalidator is and endpoint for the ImagePolicyWebhook
// admission controller of kubernetes. Used with action k8s imagevalidate, it
// validates or not each image.
package k8s_imagevalidator

import (
	"fmt"
	"encoding/json"
	"sync"
	"net/http"
	"io"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"github.com/google/uuid"

	"github.com/ObjectifLibre/csf/eventsources"
)

func init() {
	eventsources.RegisterEventSource("k8s_imagevalidator", &k8sImgPolicyWebhook{})
}

var validateTrue = `{
  "apiVersion": "imagepolicy.k8s.io/v1alpha1",
  "kind": "ImageReview",
  "status": {
    "allowed": true
  }
}`

var validateFalse = `
{
  "apiVersion": "imagepolicy.k8s.io/v1alpha1",
  "kind": "ImageReview",
  "status": {
    "allowed": false
  }
}`

var _ eventsources.EventSourceInterface = k8sImgPolicyWebhook{}

type k8sImgPolicyWebhook struct {}

func (clair k8sImgPolicyWebhook) Events() map[string][]eventsources.ArgType {
	Events := map[string][]eventsources.ArgType{
		"k8s_image_validation_request": {{T: "k8s object ImageReview", N: "image_review"},
			{T: "uuid of the request", N: "uuid"}},
	}
	return Events
}

type k8sImgPolicyWebhookHandler struct {
}


var eventsCh chan eventsources.EventData

type k8sImgPolicyWebhookConfigFromYaml struct {
	Webhook string `yaml:"webhook_endpoint"`
	TLSCert string `yaml:"certificate"`
	TLSKey string  `yaml:"key"`
}


var requests = make(map[string](chan bool))
var requestsM = sync.RWMutex{}

func GetResponseChan(resUUID string) (chan bool, error) {
	requestsM.Lock()
	defer requestsM.Unlock()
	res, ok := requests[resUUID]
	if !ok {
		return nil, fmt.Errorf("No response associated with UUID %s", resUUID)
	} else {
		return res, nil
	}
}

func (m *k8sImgPolicyWebhookHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{"eventsource": "k8s-ImagePolicyWebhook"}).Debug("Received request on webhook")

	var validationrequest map[string]interface{}
	json.NewDecoder(req.Body).Decode(&validationrequest)

	log.WithFields(log.Fields{"eventsource": "clair",
		"request": validationrequest}).Info("Received image validation request")

	curuuid := uuid.New().String()

	responseCh := make(chan bool, 1)
	requestsM.Lock()
	requests[curuuid] = responseCh
	requestsM.Unlock()

	event := eventsources.EventData{
		Name: "k8s_image_validation_request",
		Data: map[string]interface{}{
			"image_review": validationrequest,
			"uuid": curuuid,
		},
	}
	eventsCh <- event
	//Wait for ze response from the action
	resp := <- responseCh
	if resp {
		io.WriteString(res, validateTrue)
		log.WithFields(log.Fields{"action_module": "k8s_imagevalidate",
			"request": validationrequest,}).Info("Image accepted")
	} else {
		io.WriteString(res, validateFalse)
		log.WithFields(log.Fields{"action_module": "k8s_imagevalidate",
			"request": validationrequest,}).Info("Image rejected")
	}
	requestsM.Lock()
	delete(requests, curuuid)
	requestsM.Unlock()
}

func listenForValidationRequests(host string, cert string, key string) {
	http.ListenAndServeTLS(host, cert, key, &k8sImgPolicyWebhookHandler{})
}

func (clair k8sImgPolicyWebhook) Setup(ch chan eventsources.EventData, rawcfg []byte) error {
	var cfg k8sImgPolicyWebhookConfigFromYaml
	if err := yaml.Unmarshal(rawcfg, &cfg); err != nil {
		return fmt.Errorf("Bad config: %s", err)
	}
	eventsCh = ch
	go listenForValidationRequests(cfg.Webhook, cfg.TLSCert, cfg.TLSKey)
	return nil
}

