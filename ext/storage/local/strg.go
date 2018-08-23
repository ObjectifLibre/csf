package localstorage

import (
	"encoding/json"
	"errors"
	"github.com/ObjectifLibre/csf/eventhandler"
	"github.com/ObjectifLibre/csf/storage/driver"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/mitchellh/mapstructure"
)



func init() {
	storage.RegisterStorage("localdb", &localstorageImpl{})
}

type localstorageImpl struct {}

var localdb *db.DB

func (l localstorageImpl) Init(cfg map[string]interface{}) (error) {
	var err error

	path, ok := cfg["path"].(string)
	if !ok {
		return errors.New("Expected 'path' as a string")
	}
	localdb, err = db.OpenDB(path)
	if err != nil {
		return err
	}
	needToCreateCollection := true
	for _, name := range localdb.AllCols() {
		if name == "reactions" {
			needToCreateCollection = false
		}
	}
	if needToCreateCollection {
		if err := localdb.Create("reactions"); err != nil {
			return err
		}
	}
	feeds := localdb.Use("reactions")
	feeds.Index([]string{"name"})
	feeds.Index([]string{"event"})
	return nil
}

func (l localstorageImpl) GetAllReactions() ([]handler.Reaction, error) {
	feeds := localdb.Use("reactions")
	reactions := make([]handler.Reaction, 0)
	var err error
	feeds.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		var reaction handler.Reaction
		if err = json.Unmarshal(docContent, &reaction); err != nil {
			return false
		}
		reactions = append(reactions, reaction)
		return true  // move on to the next document
	})
	if err != nil {
		return nil, err
	}
	return reactions, nil
}


func (l localstorageImpl) GetReactionsForEvent(event string) ([]handler.Reaction, error) {
	feeds := localdb.Use("reactions")
	var query interface{}
	query = []interface{}{
		map[string]interface{}{"eq": event, "in": []interface{}{"event"}},
	}
	queryResult := make(map[int]struct{})
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		return nil, err
	}
	reactions := make([]handler.Reaction, 0)
	for id := range(queryResult) {
		doc, err := feeds.Read(id)
		if err != nil {
			return nil, err
		}
		var reaction handler.Reaction
		if err := mapstructure.Decode(doc, &reaction); err != nil {
			return nil, err
		}
		reactions = append(reactions, reaction)
	}
	return reactions, nil
}

func (l localstorageImpl) GetReaction(event string, name string) (handler.Reaction, error) {
	var reaction handler.Reaction
	feeds := localdb.Use("reactions")
	var query interface{}
	query = map[string]interface{}{"n": []interface{}{
		map[string]interface{}{"eq": name, "in": []interface{}{"name"}},
		map[string]interface{}{"eq": event, "in": []interface{}{"event"}},
	},
	}
	queryResult := make(map[int]struct{})
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		return reaction, err
	}
	if len(queryResult) < 1 {
		return reaction, errors.New("No such reaction")
	}
	for id := range(queryResult) {
		doc, err := feeds.Read(id)
		if err != nil {
			return reaction, err
		}
		if err := mapstructure.Decode(doc, &reaction); err != nil {
			return reaction, err
		}
		return reaction, nil
	}
	return reaction, nil
}

func (l localstorageImpl) CreateReaction(reaction handler.Reaction) (error) {
	feeds := localdb.Use("reactions")
	var reaction_generic map[string]interface{}
	jsonfied, _ := json.Marshal(reaction)
	json.Unmarshal(jsonfied, &reaction_generic)
	_, err := feeds.Insert(reaction_generic)
	if err != nil {
		return err
	}
	return nil
}

func (l localstorageImpl) DeleteReaction(event string, name string) (error) {
	feeds := localdb.Use("reactions")
	var query interface{}
	query = map[string]interface{}{"n": []interface{}{
		map[string]interface{}{"eq": name, "in": []interface{}{"name"}},
		map[string]interface{}{"eq": event, "in": []interface{}{"event"}},
	},
	}
	queryResult := make(map[int]struct{})
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		return err
	}
	for id := range(queryResult) {
		if err := feeds.Delete(id); err != nil {
			return err
		}
	}
	return nil
}

func (l localstorageImpl) Stop() error {
	return localdb.Close()
}
