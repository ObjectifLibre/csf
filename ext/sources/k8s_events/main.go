// Package k8sevents provides events fetched from kubernetes using parts of the
// client-go packages, the official go client for kubernetes.
package k8spodsevents

import (
	"time"
	"fmt"
	"github.com/ObjectifLibre/csf/eventsources"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/rest"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
)

func init() {
	eventsources.RegisterEventSource("k8s", &kubeEventSourceImplementation{})
}

var _ eventsources.EventSourceInterface = kubeEventSourceImplementation{}

var eventChannel chan eventsources.EventData

type kubeEventSourceImplementation struct {}

func (kube kubeEventSourceImplementation) Events() map[string][]eventsources.ArgType {
	Events := map[string][]eventsources.ArgType{
		"new_pod": {{T: "k8s pod struct k8s.io/api/core/v1.Pod", N: "pod"}},
	}
	return Events
}

func podCreated(obj interface{}) {
	pod := obj.(*v1.Pod)
	event := eventsources.EventData{
		Name: "new_pod",
		Data: map[string]interface{}{"pod": *pod},
	}
	eventChannel <- event
}

func watchPods(client rest.Interface) {
	//Define what we want to look for (Pods)
	watchlist := cache.NewListWatchFromClient(client, "pods", v1.NamespaceAll, fields.Everything())
	resyncPeriod := 300 * time.Second
	//Setup an informer to call functions when the watchlist changes
	_, eController := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    podCreated,
			DeleteFunc: nil,
		},
	)
	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)
}


func (kube kubeEventSourceImplementation) Setup(ch chan eventsources.EventData, cfg []byte) error {
	k8sconfig, err := clientcmd.RESTConfigFromKubeConfig(cfg)
	if err != nil {
		return fmt.Errorf("Cloud not get client from config: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		return fmt.Errorf("Could not get client from config: %s", err)
	}
	restclient := clientset.CoreV1().RESTClient()
	eventChannel = ch
	watchPods(restclient)
	return nil
}
