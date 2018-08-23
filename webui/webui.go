// Package webui is a completly-not-finished web interface for CSF.
package webui

import (
	"net/http"
	"context"
	"time"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/rakyll/statik/fs"
	"github.com/gorilla/mux"
	_ "github.com/ObjectifLibre/csf/statik"
	"github.com/ObjectifLibre/csf/templates"
	"github.com/ObjectifLibre/csf/eventsources"
	"github.com/ObjectifLibre/csf/actions"
)

var stopChannel = make(chan bool, 1)

func ServeWebUI(bindAddress string) {
	go serve(bindAddress, stopChannel)
}

func StopWebUI() {
	log.Info("Stopping Web UI")
	stopChannel <- true
}

func serve(bindAddress string, ch chan bool) {
	statikFS, err := fs.New()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Could not get statik fs")
	}

	r := mux.NewRouter()

	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(statikFS)))
	r.HandleFunc("/", homeHandler)
//	r.HandleFunc("/", newReactionHandler)


	log.WithFields(log.Fields{"address": bindAddress}).Info("Serving web interface")

	srv := &http.Server{
		Addr:         bindAddress,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
                ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	<-ch
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
}

func homeHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, templates.HomeHello("WORLD"))
}

func newReactionHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, templates.NewReaction(eventsources.GetEventSourcesList(),
		actions.GetActionModulesList()))
}
