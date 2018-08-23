// Package csfk8s provides actions to check if a specific image is currently in
// a pod or in a deployment.
package csfk8s

import (
	"strings"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ObjectifLibre/csf/actions"
)

func init() {
	actions.RegisterActionModule("k8s", &k8sModule{})
}

var clientset *kubernetes.Clientset

type k8sModule struct {}

func (k k8sModule) Actions() (map[string][]actions.ArgType, map[string][]actions.ArgType) {
	in := map[string][]actions.ArgType{
                 "is_image_deployed": {{T: "string", N: "image"}, {T: "string", N: "namespace"}},
		 "is_image_in_pods": {{T: "string", N: "image"}, {T: "string", N: "namespace"}},
	}
	out := map[string][]actions.ArgType{
		"is_image_deployed": {{T: "array of strings (names of deployments)", N: "deployments"}},
		 "is_image_in_pods": {{T: "array of strings (names of pods)", N: "pods"}},
	}
	return in, out
}

func (k k8sModule) ActionHandler() (func(string, map[string]interface{}) (map[string]interface{}, error)) {
	return actionHandler
}

func (k k8sModule) Setup(cfg []byte) error {
	k8sconfig, err := clientcmd.RESTConfigFromKubeConfig(cfg)
	if err != nil {
		return fmt.Errorf("Could not load kube config: %s", err)
	}
	clientset, err = kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		return fmt.Errorf("Could not get client from config: %s", err)
	}
	return nil
}

func actionHandler(action string, data map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "is_image_deployed":
		return isDeployed(data)
	case "is_image_in_pods":
		return isInPods(data)
	default:
		return nil, fmt.Errorf("No such action '%s'", action)
	}
}

func isInPods(data map[string]interface{}) (map[string]interface{}, error) {
	image, ok := data["image"].(string)
	if !ok {
		return nil, fmt.Errorf("Expected 'image' as string")
	}
	ns, ok := data["namespace"].(string)
	if !ok {
		ns = apiv1.NamespaceDefault
	}
	list, err := clientset.CoreV1().Pods(ns).List(metav1.ListOptions{})
        if err != nil {
		return nil, fmt.Errorf("Could not list pods: %s", err)
	}
	result := map[string]interface{}{}
	pods := []string{}
        for _, d := range list.Items {
		for _, container := range(d.Spec.Containers) {
			if strings.HasSuffix(container.Image, image) {
				pods = append(pods, d.Name)
			}
		}
        }
	result["pods"] = pods
	return result, nil
}

func isDeployed(data map[string]interface{}) (map[string]interface{}, error) {
	image, ok := data["image"].(string)
	if !ok {
		return nil, fmt.Errorf("Expected 'image' as string")
	}
	ns, ok := data["namespace"].(string)
	if !ok {
		ns = apiv1.NamespaceDefault
	}
	deploymentsClient := clientset.AppsV1().Deployments(ns)
	list, err := deploymentsClient.List(metav1.ListOptions{})
        if err != nil {
		return nil, fmt.Errorf("Could not list deployments: %s", err)
	}
	result := map[string]interface{}{}
	deployments := []string{}
        for _, d := range list.Items {
		for _, container := range(d.Spec.Template.Spec.Containers) {
			if strings.HasSuffix(container.Image, image) {
				deployments = append(deployments, d.Name)
			}
		}
        }
	result["deployments"] = deployments
	return result, nil
}
