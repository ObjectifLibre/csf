// Package metrics exposes promotheus metrics
package metrics

import (
	"fmt"
	"time"
	"net/http"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/gorilla/mux"
)


var (
	Events = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:       "csf_events_count",
			Help:       "Number of events received",
		},
		[]string{"matched"},
	)

	Reactions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:       "csf_reactions_count",
			Help:       "Number of reactions executed",
		},
		[]string{"status"},
	)

	Goroutines = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "runtime",
			Name:      "csf_goroutines_count",
			Help:      "Number of goroutines that currently exist.",
		},
		func() float64 { return float64(runtime.NumGoroutine()) },
	)

	MemUsage = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "runtime",
			Name:      "csf_mem_allocated_mb",
			Help:      "System memory allocated to csf",
		},
		func() float64 {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			return float64(memStats.Sys / (1024 * 1024))
		},
	)
)

func ServeMetrics(endpoint string) {
	prometheus.MustRegister(Events)
	prometheus.MustRegister(Reactions)
	prometheus.MustRegister(Goroutines)
	prometheus.MustRegister(MemUsage)

	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
                 Addr:         endpoint,
                 WriteTimeout: time.Second * 15,
                 ReadTimeout:  time.Second * 15,
                 IdleTimeout:  time.Second * 60,
                 Handler: r,
	}
	go func() {
                   if err := srv.ListenAndServe(); err != nil {
			   panic(fmt.Errorf("Could not serve metrics: %s", err))
                   }
	}()
}
