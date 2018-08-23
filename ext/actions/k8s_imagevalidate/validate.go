// Package k8s_imagevalidate is the action associated with the eventsource
// k8s_imagevalidator used to validate or not a container image
package k8s_imagevalidate

import (
	"fmt"

	"github.com/ObjectifLibre/csf/ext/sources/k8s_imagevalidator"
	"github.com/ObjectifLibre/csf/actions"
)

func init() {
	actions.RegisterActionModule("k8s_imagevalidate", &k8sImageValidate{})
}

var _ actions.ActionModuleInterface = k8sImageValidate{}


type k8sImageValidate struct {}

func (validate k8sImageValidate) Actions() (map[string][]actions.ArgType, map[string][]actions.ArgType) {
	in := map[string][]actions.ArgType{
		"validate_image": {
			{T: "uuid", N: "uuid"},
			{T: "bool", N: "validate"},
		}}
	out := map[string][]actions.ArgType{
		"validate_image": {}}
	return in, out
}

func k8sImageValidateOrNot(data map[string]interface{}) (map[string]interface{}, error) {
	resUUID, ok := data["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("Could not get 'uuid' parameter")
	}
	validateOrNot, ok := data["validate"].(bool)
	if !ok {
		return nil, fmt.Errorf("Could not get 'validate' parameter")
	}
	res, err := k8s_imagevalidator.GetResponseChan(resUUID)
	if err != nil {
		return nil, err
	}
	res <- validateOrNot
	return nil, nil
}

func actionHandler(action string, data map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "validate_image":
		return k8sImageValidateOrNot(data)
	default:
		return nil, fmt.Errorf("No such action '%s'", action)
	}
}

func (validate k8sImageValidate) Setup(config []byte) error {
	return nil
}

func (validate k8sImageValidate) ActionHandler() func(string, map[string]interface{}) (map[string]interface{}, error) {
	return actionHandler
}
