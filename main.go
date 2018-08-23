package main

import (
	"fmt"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
	"github.com/ObjectifLibre/csf/api"
	"github.com/ObjectifLibre/csf/eventsources"
	"github.com/ObjectifLibre/csf/actions"
	"github.com/ObjectifLibre/csf/eventhandler"
	"github.com/ObjectifLibre/csf/storage/driver"
	"github.com/ObjectifLibre/csf/configprovider"
	"github.com/ObjectifLibre/csf/webui"
	"github.com/ObjectifLibre/csf/metrics"
	_ "github.com/ObjectifLibre/csf/ext/config/localfile"
	_ "github.com/ObjectifLibre/csf/ext/storage/local"
	_ "github.com/ObjectifLibre/csf/ext/sources/onetime"
	_ "github.com/ObjectifLibre/csf/ext/sources/clair"
	_ "github.com/ObjectifLibre/csf/ext/sources/k8s_events"
	_ "github.com/ObjectifLibre/csf/ext/sources/k8s_imagevalidator"
	_ "github.com/ObjectifLibre/csf/ext/sources/openstack"
	_ "github.com/ObjectifLibre/csf/ext/actions/mail"
	_ "github.com/ObjectifLibre/csf/ext/actions/k8s"
	_ "github.com/ObjectifLibre/csf/ext/actions/k8s_imagevalidate"
	_ "github.com/ObjectifLibre/csf/ext/actions/clair"
	_ "github.com/ObjectifLibre/csf/ext/actions/vuls"
	_ "github.com/ObjectifLibre/csf/dummy_source"
	_ "github.com/ObjectifLibre/csf/dummy_action"
)

type apiConfig struct {
	BindAddress string
}

type webuiConfig struct {
	BindAddress string
	Enabled bool
}

type csfConfig struct {
	LogLevel string
	LogFormat string
	EventsBufferSize int
	ConfigProvider string
	EventSources []string
	ActionsModules []string
}

type storageConfig struct {
	Provider string
	Config map[string]interface{}
}

type configProviderConfig struct {
	Provider string
	Config map[string]interface{}
}

type metricsConfig struct {
	PrometheusEndpoint string
}

type config struct {
	Api apiConfig
	Csf csfConfig
	Storage storageConfig
	WebUI webuiConfig
	ConfigProvider configProviderConfig
	Metrics metricsConfig
}

func main() {

	var configFolder string

	flag.StringVar(&configFolder, "c", "./csf_config", "Configuration folder")
	flag.Parse()

	viper.SetConfigName("config")
	viper.AddConfigPath(configFolder)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	var configuration config
	if viper.Unmarshal(&configuration); err != nil {
		panic(fmt.Errorf("Fatal error decoding config: %s \n", err))
	}

	switch configuration.Csf.LogLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARNING":
		log.SetLevel(log.WarnLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	switch configuration.Csf.LogFormat {
	case "logfmt":
		log.SetFormatter(&log.TextFormatter{})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}

	events := make(chan eventsources.EventData, configuration.Csf.EventsBufferSize)

	metrics.ServeMetrics(configuration.Metrics.PrometheusEndpoint)

	cfgprovider, err := configprovider.GetConfigProvider(configuration.ConfigProvider.Provider)
	if err != nil {
		panic(fmt.Errorf("Could not get config provider: %s \n", err))
	}

	err = cfgprovider.Setup(configuration.ConfigProvider.Config)
	if err != nil {
		panic(fmt.Errorf("Could init configuration provider: %s \n", err))
	}
	eventsources.SetupEventSources(events, cfgprovider, configuration.Csf.EventSources)
	actions.SetupActionModules(cfgprovider, configuration.Csf.ActionsModules)

	db, err := storage.GetStorage(configuration.Storage.Provider)
	if err != nil {
		panic(fmt.Errorf("Could not get storage: %s \n", err))
	}

	if err := db.Init(configuration.Storage.Config); err != nil {
		panic(fmt.Errorf("Could not init storage: %s \n", err))
	}

	if reactions, err := db.GetAllReactions(); err != nil {
		panic(fmt.Errorf("Could not get data from storage: %s \n", err))
	} else {
		for _, reaction := range(reactions) {
			handler.AddReaction(reaction)
		}
	}

	go handler.HandleEvents(events)

	api.ServeAPI(configuration.Api.BindAddress, db)
	if configuration.WebUI.Enabled {
		webui.ServeWebUI(configuration.WebUI.BindAddress)
	}

	//Wait for shutdown signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	//Shutdown everything
	api.StopAPI()
	webui.StopWebUI()
	db.Stop()
}
