package main

import (
	"fmt"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
	"net/http"
	"os"
	"time"
	"crypto/tls"
)

func init() {
	log.SetFormatter(new(log.JSONFormatter))
}

func main() {
	app := cli.App("locations-transformer", "A RESTful API for transforming TME Locations to UP json")
	username := app.String(cli.StringOpt{
		Name:   "tme-username",
		Value:  "",
		Desc:   "TME username used for http basic authentication",
		EnvVar: "TME_USERNAME",
	})
	password := app.String(cli.StringOpt{
		Name:   "tme-password",
		Value:  "",
		Desc:   "TME password used for http basic authentication",
		EnvVar: "TME_PASSWORD",
	})
	baseURL := app.String(cli.StringOpt{
		Name:   "base-url",
		Value:  "http://localhost:8080/transformers/locations/",
		Desc:   "Base url",
		EnvVar: "BASE_URL",
	})
	tmeBaseURL := app.String(cli.StringOpt{
		Name:   "tme-base-url",
		Value:  "https://tme-live.internal.ft.com:40001",
		Desc:   "TME base url",
		EnvVar: "TME_BASE_URL",
	})
	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "Port to listen on",
		EnvVar: "PORT",
	})

	app.Action = func() {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c := &http.Client{
			Transport: tr,
			Timeout: time.Duration(20 * time.Second),
		}

		s, err := newLocationService(newTmeRepository(c, *tmeBaseURL, *username, *password), *baseURL)
		if err != nil {
			log.Errorf("Error while creating LocationsService: [%v]", err.Error())
		}
		h := newLocationsHandler(s)
		m := mux.NewRouter()
		m.HandleFunc("/transformers/locations", h.getLocations).Methods("GET")
		m.HandleFunc("/transformers/locations/{uuid}", h.getLocationByUUID).Methods("GET")
		http.Handle("/", m)

		log.Printf("listening on %d", *port)
		http.ListenAndServe(fmt.Sprintf(":%d", *port),
			httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry,
				httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), m)))
	}
	app.Run(os.Args)
}
