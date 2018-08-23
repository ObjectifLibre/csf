package localstorage

import (
	"os"
	"io/ioutil"
	"strconv"
	"math/rand"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/ObjectifLibre/csf/eventhandler"
	"github.com/ObjectifLibre/csf/storage/driver"
)
type LocalStorageTestSuite struct {
	suite.Suite
	dir string
	db  storage.StorageInterface
	sampleReaction handler.Reaction
}


func (suite *LocalStorageTestSuite) SetupSuite() {
	db, err := storage.GetStorage("localdb")
	if err != nil {
		panic(err)
	}

	suite.sampleReaction = handler.Reaction{
		Name: "test_name",
		Event: "test_event",
		Script: "console.log('hello');",
		Actions: map[string]handler.Action{
			"action1": {
				Script: "console.log('test');",
				Action: "test_action",
				Module: "test_module",
			},
		},
	}

	//Clean db
	suite.dir, err = ioutil.TempDir("", "csf_db")
	if err != nil {
		panic(err)
	}
	cfg :=  map[string]interface{}{"path": suite.dir}
	if err = db.Init(cfg); err != nil {
		panic(err)
	}
	suite.db = db
}


func (suite *LocalStorageTestSuite) TestInsertAndDeleteSingleAction() {
	//Check that the db is empty
	reactions, err := suite.db.GetAllReactions()
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 0, len(reactions))
	//Add a reaction
	err = suite.db.CreateReaction(suite.sampleReaction)
	assert.Equal(suite.T(), nil, err)
	//Check that the new reaction exists
	reactions, err = suite.db.GetAllReactions()
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), len(reactions), 1)
	assert.Equal(suite.T(), reactions[0], suite.sampleReaction)
	//Remove it and check if it is gone
	err = suite.db.DeleteReaction("test_event", "test_name")
	assert.Equal(suite.T(), nil, err)
	reactions, err = suite.db.GetAllReactions()
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 0, len(reactions))
}

func (suite *LocalStorageTestSuite) TestInsertAndDeleteMultipleActions() {
	//Check that the db is empty
	reactions, err := suite.db.GetAllReactions()
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 0, len(reactions))

	//Generate some samples
	events := []string{"michel", "jeanne", "roger"}
	reactionsSample := make([]handler.Reaction, 0)
	for _, event := range(events) {
		reaction := suite.sampleReaction
		reaction.Name = event + "_name"
		reaction.Event = event
		reactionsSample = append(reactionsSample, reaction)
	}

	//Push them to the db
	for _, reaction := range(reactionsSample) {
		err = suite.db.CreateReaction(reaction)
		assert.Equal(suite.T(), nil, err)
	}

	//Check that they all exists
	reactions, err = suite.db.GetAllReactions()
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), len(reactions), 3)

	//Get them one by one and remove them
	for _, reaction := range(reactionsSample) {
		cur_reaction, err := suite.db.GetReaction(reaction.Event, reaction.Name)
		assert.Equal(suite.T(), nil, err)
		assert.Equal(suite.T(), reaction, cur_reaction)
		err = suite.db.DeleteReaction(reaction.Event, reaction.Name)
		assert.Equal(suite.T(), nil, err)
	}

	//Check that they are all gone
	reactions, err = suite.db.GetAllReactions()
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 0, len(reactions))
}

func (suite *LocalStorageTestSuite) TestInsertAndDeleteForOneEvent() {
	//Check that the db is empty
	reactions, err := suite.db.GetAllReactions()
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 0, len(reactions))

	//Generate some samples
	reactionsSample := make([]handler.Reaction, 0)
	for i := 0; i < 4; i += 1 {
		reaction := suite.sampleReaction
		reaction.Name =  string(strconv.Itoa(rand.Int()))
		reaction.Event = "same_event"
		reactionsSample = append(reactionsSample, reaction)
		err = suite.db.CreateReaction(reaction)
		assert.Equal(suite.T(), nil, err)
	}


	//Check that they all exists
	reactions, err = suite.db.GetReactionsForEvent("same_event")

	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 4, len(reactions))

	//Get them one by one and remove them
	for _, reaction := range(reactionsSample) {
		cur_reaction, err := suite.db.GetReaction(reaction.Event, reaction.Name)
		assert.Equal(suite.T(), nil, err)
		assert.Equal(suite.T(), reaction, cur_reaction)
		err = suite.db.DeleteReaction(reaction.Event, reaction.Name)
		assert.Equal(suite.T(), nil, err)
	}

	//Check that they are all gone
	reactions, err = suite.db.GetAllReactions()
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 0, len(reactions))
}

func (suite *LocalStorageTestSuite) TearDownSuite() {
	os.RemoveAll(suite.dir)
}

func TestLocalStorageTestSuite(t *testing.T) {
	suite.Run(t, new(LocalStorageTestSuite))
}
