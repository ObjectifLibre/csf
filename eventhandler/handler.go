//Package hander handles reactions. A reaction is basically a series of actions
// trigerred by an event.
package handler

import (
	"sync"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ObjectifLibre/csf/eventsources"
	"github.com/ObjectifLibre/csf/scripting"
	"github.com/ObjectifLibre/csf/actions"
	"github.com/ObjectifLibre/csf/metrics"
)

// Action represents an action. It contains the JavaScript code to execute
// after the action, the name of the action and the name of the action module
// containing the action.
type Action struct {
	Script string `json:"script"` // JS code
	Action string `json:"action"` // Name of the action
	Module string `json:"module"` // Name of the module containing the action
}

// Reaction represents a reaction. It contains a list of actions to execute,
// the name of the reaction and the name of the event.
type Reaction struct {
	Actions map[string]Action `json:"actions"`
	Name string `json:"name"`
	Event string `json:"event"`
	Script string `json:"script"`
}

var reactions = make(map[string][]Reaction) //Keys are event names and values are list of reactions
var reactionsM  sync.RWMutex


// AddReaction adds a reaction to the local map of reactions.
func AddReaction(reaction Reaction) {
	reactionsM.Lock()
	defer reactionsM.Unlock()
	reactions[reaction.Event] = append(reactions[reaction.Event], reaction)
}

// RemoveReaction removes a reaction from the local map of reactions.
func RemoveReaction(event string, name string) (error) {
	reactionsM.Lock()
	defer reactionsM.Unlock()
	if _, ok := reactions[event]; !ok {
		return fmt.Errorf("No reactions matching event '%s'", event)
	}
	for i, reaction := range(reactions[event]) {
		if reaction.Name == name {
			reactions[event] = append(reactions[event][:i], reactions[event][i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Reaction '%s' not found", name)
}

// launchReactionPipeline executes a reaction. It first executes the reaction
// script that will return which action to execute next. Then we run actions as
// long as the actions scripts return an action to execute.
func launchReactionPipeline(reaction Reaction, event eventsources.EventData) {
	var scriptResult map[string]interface{} = nil

	log.WithFields(log.Fields{"reaction": reaction, "event": event}).Debug("Starting reaction pipeline")

	var action Action
	var actionName string

	if nextAction, result, err := scripting.Exec(event.Data, nil, reaction.Script); err != nil {
		log.WithFields(log.Fields{"err": err, "reaction": reaction.Name}).Warn("Could not run reaction script")
		metrics.Reactions.With(prometheus.Labels{"status": "error"}).Inc()
		return
	} else if len(nextAction) == 0 {
		log.WithFields(log.Fields{"reaction": reaction.Name, "event": event}).Info("Script returned no next action, not launching pipeline")
		metrics.Reactions.With(prometheus.Labels{"status": "not launched"}).Inc()
		return
	} else {
		actionToRun, ok := reaction.Actions[nextAction]
		if !ok {
			log.WithFields(log.Fields{"reaction": reaction.Name, "action": nextAction}).Warn("No such action")
		}
		actionName = nextAction
		scriptResult = result
		action = actionToRun
	}

	for {
		log.WithFields(log.Fields{"reaction": reaction, "event": event, "action": actionName}).Debug("Starting reaction exec")

		log.WithFields(log.Fields{"event": event.Name, "reaction": reaction.Name,
			"action": actionName}).Info("Running action")
		actionResult, err := actions.RunAction(action.Module, action.Action, scriptResult)
		if err != nil {
			log.WithFields(log.Fields{"err": err,
				"action": actionName,
				"event": event.Name,
				"reaction": reaction.Name}).Warn("Could not run action")
			return
		}

		if len(action.Script) == 0 {
			log.WithFields(log.Fields{"event": event.Name,
				"reaction": reaction.Name}).Info("No script for action, stopping pipeline")
			metrics.Reactions.With(prometheus.Labels{"status": "done"}).Inc()
			return
		}

		nextAction, result, err := scripting.Exec(event.Data, actionResult, action.Script)
		if err != nil {
			log.WithFields(log.Fields{"reaction": reaction.Name,
				"script": action.Script,
				"err": err}).Warn("Action script exec error")
			metrics.Reactions.With(prometheus.Labels{"status": "error"}).Inc()
			return
		} else if len(nextAction) == 0 {
			log.WithFields(log.Fields{"event": event.Name,
				"reaction": reaction.Name}).Info("No next action provided, stopping pipeline")
			metrics.Reactions.With(prometheus.Labels{"status": "done"}).Inc()
			return
		} else {
			scriptResult = result
		}

		action = reaction.Actions[nextAction]
		actionName = nextAction
	}
}

// handleEvent calls launchReactionPipeline for every reaction matching the
// event in a goroutine.
func handleEvent(event eventsources.EventData) {
	reactionsM.RLock()
	defer reactionsM.RUnlock()
	if _, ok := reactions[event.Name]; !ok {
		log.WithFields(log.Fields{
			"event": event.Name}).Info("No reaction matching event")
		metrics.Events.With(prometheus.Labels{"matched": "false"}).Inc()
		return
	}
	metrics.Events.With(prometheus.Labels{"matched": "true"}).Inc()
	for _, reaction := range(reactions[event.Name]) {
		go launchReactionPipeline(reaction, event)
	}
}

// HandleEvents wait for events and calls handleEvent in a goroutine for each
// event.
func HandleEvents (ch chan eventsources.EventData) {
	for {
		event := <- ch
		log.WithFields(log.Fields{"event": event.Name}).Info("Received event")
		go handleEvent(event)
	}
}
