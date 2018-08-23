// Package actions provides types and interfaces necessary to write action
// modules.
package actions

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

// ActionModuleInterface is the interface actions modules must implement.
// Setup() is called once when CSF starts and is used to configure and or
// init modules. Actions() returns a map of in and out parameters where Keys
// are actions names and values are lists of parameters of the given action.
// ActionHandler() is called everytime an action of the module needs to be
// Executed. It is basically a switch that call the right function.
type ActionModuleInterface interface {
	Setup(config []byte) error
	Actions() (map[string][]ArgType, map[string][]ArgType)
	ActionHandler() func(string, map[string]interface{}) (map[string]interface{}, error)
}

// actionModule is used internally to keep track of the different modules
// registered.
type actionModule struct {
	i ActionModuleInterface
	handler func(string, map[string]interface{}) (map[string]interface{}, error)
	actionsIn map[string][]ArgType
	actionsOut map[string][]ArgType
}

var modules = make(map[string]actionModule)
var modulesM sync.RWMutex


// RegisterActionModule registers the action module. Must be called from the
// init() of modules.
func RegisterActionModule(name string, ac ActionModuleInterface) {
	modulesM.Lock()
	defer modulesM.Unlock()
	actionsIn, actionsOut := ac.Actions()
	modules[name] = actionModule{i: ac,
		handler: ac.ActionHandler(),
		actionsIn: actionsIn,
		actionsOut: actionsOut,
	}
}

// SetupActionModules gets the configuration and calls the Setup() function
// of the modules specified by actionslist.
func SetupActionModules(cfp configprovider.ConfigProviderInterface, actionslist []string) {
	modulesM.Lock()
	defer modulesM.Unlock()
	log.Debug("Starting action modules setup")
	for _, name := range(actionslist) {
		mod, ok := modules[name]
		if !ok {
			log.WithFields(log.Fields{"action_module": name}).Error("No such action module")
			continue
		}
		log.WithFields(log.Fields{"action_module": name}).Debug("Action module setup started")
		go func(module actionModule, modulename string){
			config, err := cfp.GetActionModuleConfig(modulename)
			if err != nil {
				log.WithFields(log.Fields{"err": err, "action_module": modulename}).Warn("Could not get config, using empty conf")
				config = make([]byte, 0)
			}
			if err = module.i.Setup(config); err != nil {
				log.WithFields(log.Fields{
					"action_module": modulename, "err": err}).Error("Could not setup action module")
				return
			}
			log.WithFields(log.Fields{"action_module": modulename}).Debug("Action module set up")
		}(mod, name)
	}
}

// RunAction calls the ActionHandler() of the module containing the action.
func RunAction(module string, action string, data map[string]interface{}) (map[string]interface{}, error) {
	log.WithFields(log.Fields{"action_module": module, "action_name": action,
		"data": data}).Debug("Running action")
	modulesM.RLock()
	defer modulesM.RUnlock()
	return modules[module].handler(action, data)
}

// Type describing an action
type ActionDesc struct {
	In []ArgType `json:"data_in"`
	Out []ArgType `json:"data_out"`
	Name string `json:"name"`
}

func GetAllActions() map[string][]ActionDesc {
	allActions := make(map[string][]ActionDesc)

	modulesM.RLock()
	defer modulesM.RUnlock()

	for name, module := range(modules) {
		allActions[name] = make([]ActionDesc, 0)
		for actionname, args := range(module.actionsIn) {
			action := ActionDesc{
				Name: actionname,
				In: args,
				Out: module.actionsOut[actionname],
			}
			allActions[name] = append(allActions[name], action)
		}
	}
	return allActions
}

// GetActionModulesList returns the list of all action modules.
func GetActionModulesList() (moduleslist []string)  {
	moduleslist = make([]string, len(modules))
	modulesM.RLock()
	defer modulesM.RUnlock()
	for module := range(modules) {
		moduleslist = append(moduleslist, module)
	}
	return
}
