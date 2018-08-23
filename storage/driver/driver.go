// Package storage provides an interface to implement storage providers.
package storage

import (
	"fmt"
	"sync"

	"github.com/ObjectifLibre/csf/eventhandler"
)

// StorageInterface is the interface defining a storage provider.
type StorageInterface interface {
	Init(cfg map[string]interface{}) error
	GetAllReactions() ([]handler.Reaction, error)
	GetReactionsForEvent(event string) ([]handler.Reaction, error)
	GetReaction(event string, name string) (handler.Reaction, error)
	CreateReaction(handler.Reaction) error
	DeleteReaction(event string, name string) error
	Stop() error
}

var storages = make(map[string]StorageInterface)
var storagesM sync.RWMutex

// RegisterStorage registers a storage. Must be called from the init() function
// of the storage implementation.
func RegisterStorage(name string, strg StorageInterface) {
	storagesM.Lock()
	defer storagesM.Unlock()
	storages[name] = strg
}

// GetStorage returns the storage provider by its name.
func GetStorage(name string) (StorageInterface, error) {
	storagesM.RLock()
	defer storagesM.RUnlock()
	storage, ok := storages[name]
	if !ok {
		return nil, fmt.Errorf("No storage named %s", name)
	} else {
		return storage, nil
	}
}
