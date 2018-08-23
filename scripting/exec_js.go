
// Package scripting executes action's JavaScript code in a JS vm.
package scripting

import (
	"fmt"
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"
)



// This code was used to expose a go function to the js environment
// Not sure if there's a use case for this

// func ExposeFunction(fn_name string,
// 	fn func(otto.FunctionCall)(otto.Value)) {
// 	err := vm.Set(fn_name, fn)
// 	if err != nil {
// 		log.WithFields(log.Fields{"fn_name": fn_name,
// 			"function": fn}).Warn("Could not register function in JS vm")
// 	}
// }

// Exec executes js code in a vm, exposing the event data and the data from
// the previous action. The goal is to have as much data as possible to allow
// automatic decision making.
// "event" is the event data sent by a eventsource, "prev" is the data from the
// previous action, code is the JS script to execute.
func Exec(event map[string]interface{},	prevActionData map[string]interface{}, code string) (string, map[string]interface{}, error) {

	vm := otto.New()

	vm.Set("event", event)
	vm.Set("actionData", prevActionData)
	vm.Set("nextAction", "")
	vm.Set("err", "")

	log.WithFields(log.Fields{"event_data": event,
		"code": code, "action_data": prevActionData}).Debug("Running script")
	if _, err := vm.Run(code); err != nil {
		log.WithFields(log.Fields{"event_data": event,
			"code": code, "err": err}).Warn("JS script exec failed")
		return "",  nil, err
	}

	// Get err string from JS VM
	if res, err := vm.Get("err"); err != nil {
		log.WithFields(log.Fields{"event_data": event,
			"code": code, "err": err}).Warn("Could not get err variable")
		return "",  nil, err
	// Convert it to go string
	} else if js_err, err := res.ToString(); err != nil  {
		log.WithFields(log.Fields{"event_data": event,
			"code": code, "err": err}).Warn("err js variable is not a string")
		return "",  nil, err
	// Check if err is not empty string
	} else if len(js_err) > 0 {
		log.WithFields(log.Fields{"event_data": event,
			"code": code, "err": js_err}).Warn("JS script returned errror")
		return "",  nil, fmt.Errorf("Error in js vm: %s", err)
	// Get the "nextAction" string to know which action to run next
	} else if res, err = vm.Get("nextAction"); err != nil {
		return "",  nil, err
	// Convert it to a boolean
	} else if nextAction, err := res.ToString(); err != nil  {
		return "",  nil, err
	// Don't bother to go further if we won't continue the pipeline
	} else if len(nextAction) == 0 {
		return "",  nil, nil
	// Get the "result" object from the vm
	} else if res, err = vm.Get("result"); err != nil {
		return "",  nil, err
	// Get it as interface{}
	} else if out, err := res.Export(); err != nil {
		return "",  nil, err
	} else if result, ok := out.(map[string]interface{}); !ok {
		return "",  nil, fmt.Errorf("'result' is not an object")
	} else {
		return nextAction, result, nil
	}
}
