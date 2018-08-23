// Package configprovider provides a simple interface to implement configuration
// providers for action modules and event sources.
package configprovider

import (
	"fmt"
	"sync"
)

// ConfigProviderInterface is the interface of configuration providers.
// Setup() is called when CSF starts and can be used to configure the
// configuration provider. \o/
// GetEventSourceConfig() should return the configuration of the given
// event source and GetActionModuleConfig() should returns the configuration
// of the given action module.
type ConfigProviderInterface interface {
	GetEventSourceConfig(eventsource string) ([]byte, error)
	GetActionModuleConfig(actionmodule string) ([]byte, error)
	Setup(config map[string]interface{}) error
}

var providers = make(map[string]ConfigProviderInterface)
var providersM sync.RWMutex

// RegisterConfigProvider registers a config provider. Must be called from the
// init() of the configprovider.
func RegisterConfigProvider(name string, provider ConfigProviderInterface) {
	providersM.Lock()
	defer providersM.Unlock()
	providers[name] = provider
}

// GetConfigProvider returns a configprovider by its name.
func GetConfigProvider(name string) (ConfigProviderInterface, error) {
	providersM.RLock()
	defer providersM.RUnlock()
	provider, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("No config provider named %s", name)
	} else {
		return provider, nil
	}
}
