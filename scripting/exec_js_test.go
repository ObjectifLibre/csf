package scripting

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)
type ScriptingTestSuite struct {
	suite.Suite
}


func (suite *ScriptingTestSuite) TestExecGood() {
	event := map[string]interface{}{
		"key": "value",
		"other_key": 42,
	}
	code := `if (event["key"] === "value") {
nextAction = 'actionSuivante';
result = {"somekey": "somevalue"};
} else {
nextAction = 'foo';
}`
	nextAction, result, err := Exec(event, nil, code)
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), "actionSuivante", nextAction)
	assert.Equal(suite.T(), map[string]interface{}{"somekey": "somevalue"}, result)
}

func (suite *ScriptingTestSuite) TestExecErr() {
	event := map[string]interface{}{
		"key": "value",
		"other_key": 42,
	}
	code := `if (event["key"] === "value") {
nextAction = 'actionSuivante';
result = {"somekey": "somevalue"};
err = "Im an error"
} else {
nextAction = 'foo';
}`
	nextAction, result, err := Exec(event, nil, code)
	assert.NotEqual(suite.T(), nil, err)
	assert.Equal(suite.T(), "", nextAction)
	assert.Equal(suite.T(), map[string]interface{}(nil), result)
}

func (suite *ScriptingTestSuite) TestExecWrongVarType() {
	event := map[string]interface{}{
		"key": "value",
		"other_key": 42,
	}
	code := `if (event["key"] === "value") {
nextAction = 'toto';
result = 42;
} else {
nextAction = 'foo';
}`
	nextAction, result, err := Exec(event, nil, code)
	assert.NotEqual(suite.T(), nil, err)
	assert.Equal(suite.T(), "", nextAction)
	assert.Equal(suite.T(), map[string]interface{}(nil), result)
}

func TestScriptingTestSuite(t *testing.T) {
    suite.Run(t, new(ScriptingTestSuite))
}
