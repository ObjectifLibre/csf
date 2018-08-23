package localfilesconfig

import (
	"os"
	"io/ioutil"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/ObjectifLibre/csf/configprovider"
)
type LocalConfigTestSuite struct {
	suite.Suite
	dir string
	cfgprovider configprovider.ConfigProviderInterface
}


func (suite *LocalConfigTestSuite) SetupSuite() {
	cfgprovider, err := configprovider.GetConfigProvider("localfiles")
	if err != nil {
		panic(err)
	}
	suite.cfgprovider = cfgprovider
	suite.dir, err = ioutil.TempDir("", "test_cfgprovider")
	if err != nil {
		panic(err)
	}
	cfg := map[string]interface{}{"path": suite.dir}
	if err := suite.cfgprovider.Setup(cfg); err != nil {
		panic(err)
	}
}


func (suite *LocalConfigTestSuite) TestGetActionModuleConfig() {
	raw_config := []byte(`sample config file.
Config file can be of any format as they are passed as []byte to modules.
This allows a generic interface and implementation-defined configuration formats.`)
	tmpfn := filepath.Join(suite.dir, "testModule.cfg")
	if err := ioutil.WriteFile(tmpfn, raw_config, 0666); err != nil {
		panic(err)
	}
	config, err := suite.cfgprovider.GetEventSourceConfig("testModule")
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), raw_config, config)
}

func (suite *LocalConfigTestSuite) TearDownSuite() {
         os.RemoveAll(suite.dir)
}


func TestLocalConfigTestSuite(t *testing.T) {
    suite.Run(t, new(LocalConfigTestSuite))
}
