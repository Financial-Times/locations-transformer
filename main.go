package main

import (
	"crypto/tls"
	"fmt"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Financial-Times/tme-reader/tmereader"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
	"github.com/sethgrid/pester"
	"net"
	"net/http"
	"os"
	"time"
)

func init() {
	log.SetFormatter(new(log.JSONFormatter))
}

func main() {
	app := cli.App("locations-transformer", "A RESTful API for transforming TME locations to UP json")
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
	token := app.String(cli.StringOpt{
		Name:   "token",
		Value:  "",
		Desc:   "Token to be used for accessing TME",
		EnvVar: "TOKEN",
	})
	baseURL := app.String(cli.StringOpt{
		Name:   "base-url",
		Value:  "http://localhost:8080/transformers/locations/",
		Desc:   "Base url",
		EnvVar: "BASE_URL",
	})
	tmeBaseURL := app.String(cli.StringOpt{
		Name:   "tme-base-url",
		Value:  "https://tme.ft.com",
		Desc:   "TME base url",
		EnvVar: "TME_BASE_URL",
	})
	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "Port to listen on",
		EnvVar: "PORT",
	})
	maxRecords := app.Int(cli.IntOpt{
		Name:   "maxRecords",
		Value:  int(10000),
		Desc:   "Maximum records to be queried to TME",
		EnvVar: "MAX_RECORDS",
	})
	slices := app.Int(cli.IntOpt{
		Name:   "slices",
		Value:  int(10),
		Desc:   "Number of requests to be executed in parallel to TME",
		EnvVar: "SLICES",
	})

	tmeTaxonomyName := "GL"

	app.Action = func() {
		client := getResilientClient()

		mf := new(locationTransformer)
		s, err := newLocationService(tmereader.NewTmeRepository(client, *tmeBaseURL, *username, *password, *token, *maxRecords, *slices, tmeTaxonomyName, &tmereader.AuthorityFiles{}, mf), *baseURL, tmeTaxonomyName, *maxRecords)
		if err != nil {
			log.Errorf("Error while creating LocationsService: [%v]", err.Error())
		}

		h := newLocationsHandler(s)
		m := mux.NewRouter()

		// The top one of these feels more correct, but the lower one matches what we have in Dropwizard,
		// so it's what apps expect currently same as ping
		m.HandleFunc(status.PingPath, status.PingHandler)
		m.HandleFunc(status.PingPathDW, status.PingHandler)
		m.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
		m.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)
		m.HandleFunc("/__health", v1a.Handler("Locations Transformer Healthchecks", "Checks for accessing TME", h.HealthCheck()))
		m.HandleFunc("/__gtg", h.GoodToGo)

		m.HandleFunc("/transformers/locations", h.getLocations).Methods("GET")
		m.HandleFunc("/transformers/locations/__count", h.getCount).Methods("GET")
		m.HandleFunc("/transformers/locations/__ids", h.getIds).Methods("GET")
		m.HandleFunc("/transformers/locations/__reload", h.getIds).Methods("GET")
		m.HandleFunc("/transformers/locations/{uuid}", h.getLocationByUUID).Methods("GET")

		http.Handle("/", m)

		log.Printf("listening on %d", *port)
		err = http.ListenAndServe(fmt.Sprintf(":%d", *port),
			httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry,
				httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), m)))
		if err != nil {
			log.Errorf("Error by listen and serve: %v", err.Error())
		}

	}
	app.Run(os.Args)
}

func getResilientClient() *pester.Client {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 128,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}
	c := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(30 * time.Second),
	}
	client := pester.NewExtendedClient(c)
	client.Backoff = pester.ExponentialBackoff
	client.MaxRetries = 5
	client.Concurrency = 1

	return client
}
