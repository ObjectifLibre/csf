// Package localfilesconfig is a configuration provider for event sources and
// action modules that uses simple files in a local directory and uses file
// names modules to match configurations.
package localfilesconfig

import (
	"os"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"github.com/ObjectifLibre/csf/configprovider"
)

func init () {
	configprovider.RegisterConfigProvider("localfiles", &LocalFilesConfigProvider{})
}

var configPath string

type LocalFilesConfigProvider struct {}

func (l LocalFilesConfigProvider) Setup(cfg map[string]interface{}) error {
	if path, ok := cfg["path"].(string); !ok {
		return errors.New("Expected 'path' as string")
	} else {
		configPath = path
		return nil
	}
}

func (l LocalFilesConfigProvider) GetEventSourceConfig(ds string) ([]byte, error) {
	return getFileBasedOnName(ds)
}

func (l LocalFilesConfigProvider) GetActionModuleConfig(am string) ([]byte, error) {
	return getFileBasedOnName(am)
}

func getFileBasedOnName(configname string) ([]byte, error)  {
	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		return nil, err
	}
	for _, file := range(files) {
		ext := filepath.Ext(file.Name())
		filename := file.Name()[0:len(file.Name())-len(ext)]
		path := filepath.Join(configPath, file.Name())
		if strings.EqualFold(filename, configname) {
			if buff, err := ioutil.ReadFile(path); err != nil {
				return nil, err
			} else {
				return buff, nil
			}
		}
	}
	return nil, os.ErrNotExist
}
