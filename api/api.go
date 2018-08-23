package api

import (
	"time"
	"context"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/ObjectifLibre/csf/eventhandler"
	"github.com/ObjectifLibre/csf/storage/driver"
	"github.com/ObjectifLibre/csf/eventsources"
	"github.com/ObjectifLibre/csf/actions"
)

var db storage.StorageInterface
var stopChannel = make(chan bool)

func getAllReactions(w http.ResponseWriter, r *http.Request) {
	reactions, err := db.GetAllReactions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonbody, err := json.Marshal(reactions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonbody)
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	events := eventsources.GetAllEvents()
	body, _ := json.Marshal(events)
	w.Write(body)
}

func getAllActions(w http.ResponseWriter, r *http.Request) {
	actionList := actions.GetAllActions()
	body, _ := json.Marshal(actionList)
	w.Write(body)
}

func getReactionsForEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	events, err := db.GetReactionsForEvent(vars["event"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errbody, _ := json.Marshal(map[string]interface{}{"err": err.Error()})
		w.Write(errbody)
	}
	body, _ := json.Marshal(events)
	w.Write(body)
}

func postReactionForEvent(w http.ResponseWriter, r *http.Request) {
	var reaction handler.Reaction

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	err = json.Unmarshal(b, &reaction)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err = db.CreateReaction(reaction); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	handler.AddReaction(reaction)
}

func getReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reaction, err := db.GetReaction(vars["event"], vars["reaction"])
	if err != nil {
		if err.Error() == "No such reaction" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		errbody, _ := json.Marshal(map[string]interface{}{"err": err.Error()})
		w.Write(errbody)
		return
	}
	body, _ := json.Marshal(reaction)
	w.Write(body)
}

func delReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := db.DeleteReaction(vars["event"], vars["reaction"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	handler.RemoveReaction(vars["event"], vars["reaction"])
}

func launchAPI(host string, ch chan bool) {
	r := mux.NewRouter()
	r.HandleFunc("/v1/reactions", getAllReactions).Methods("GET")
	r.HandleFunc("/v1/actions", getAllActions).Methods("GET")
	r.HandleFunc("/v1/events", getEvents).Methods("GET")
	r.HandleFunc("/v1/events/{event}/reactions", getReactionsForEvent).Methods("GET")
	r.HandleFunc("/v1/events/{event}/reactions", postReactionForEvent).Methods("POST")
	r.HandleFunc("/v1/events/{event}/reactions/{reaction}", getReaction).Methods("GET")
	r.HandleFunc("/v1/events/{event}/reactions/{reaction}", delReaction).Methods("DELETE")
	srv := &http.Server{
		Addr:         host,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: r,
	}
	log.WithFields(log.Fields{"address": host}).Info("Starting API")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	<-ch
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
}

func StopAPI() {
	log.Info("Stopping API")
	stopChannel <- true
}

func ServeAPI(host string, database storage.StorageInterface) {
	db = database
	go launchAPI(host, stopChannel)
}
