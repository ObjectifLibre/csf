// Package dummyaction is a dummy implementation of an action module
// for reference and debugging purposes.
package dummyaction

import (
	"fmt"
	"github.com/ObjectifLibre/csf/actions"
)

func init() {
	actions.RegisterActionModule("dummy", &dummyActionModuleImplementation{})
}

var _ actions.ActionModuleInterface = dummyActionModuleImplementation{}

type dummyActionModuleImplementation struct {}

// Actions returns only one action named "dummy_action". It only prints what it
// receives.
func (dummy dummyActionModuleImplementation) Actions() (map[string][]actions.ArgType, map[string][]actions.ArgType) {
	in := map[string][]actions.ArgType{
		"dummy_action": {{T: "string", N: "dummy_string"},
		}}
	out := map[string][]actions.ArgType{
		"dummy_action": {}}
	return in, out
}

func dummyActionHandler(action string, data map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("DummyAction: Recieved action", action, " with data", data)
	result := map[string]interface{}{"dummykey": "dummyvalue"}
	return result, nil
}

// Setup only print the configuration it receives, for debugging purposes.
func (dummy dummyActionModuleImplementation) Setup(config []byte) error {
	fmt.Println("Dummy Action received conf: " + string(config))
	return nil
}

func (dummy dummyActionModuleImplementation) ActionHandler() func(string, map[string]interface{}) (map[string]interface{}, error) {
	return dummyActionHandler
}
