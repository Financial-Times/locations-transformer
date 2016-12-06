package main

import (
	"crypto/tls"
	"fmt"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/service-status-go/gtg"
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
	graphiteTCPAddress := app.String(cli.StringOpt{
		Name:   "graphiteTCPAddress",
		Value:  "",
		Desc:   "Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)",
		EnvVar: "GRAPHITE_ADDRESS",
	})
	graphitePrefix := app.String(cli.StringOpt{
		Name:   "graphitePrefix",
		Value:  "",
		Desc:   "Prefix to use. Should start with content, include the environment, and the host name. e.g. content.test.public.content.by.concept.api.ftaps59382-law1a-eu-t",
		EnvVar: "GRAPHITE_PREFIX",
	})
	logMetrics := app.Bool(cli.BoolOpt{
		Name:   "logMetrics",
		Value:  false,
		Desc:   "Whether to log metrics. Set to true if running locally and you want metrics output",
		EnvVar: "LOG_METRICS",
	})

	tmeTaxonomyName := "GL"

	app.Action = func() {
		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)
		client := getResilientClient()

		mf := new(locationTransformer)
		s, err := newLocationService(tmereader.NewTmeRepository(client, *tmeBaseURL, *username, *password, *token, *maxRecords, *slices, tmeTaxonomyName, &tmereader.AuthorityFiles{}, mf), *baseURL, tmeTaxonomyName, *maxRecords)
		if err != nil {
			log.Errorf("Error while creating LocationsService: [%v]", err.Error())
		}

		h := newLocationsHandler(s)
		m := mux.NewRouter()

		m.HandleFunc("/transformers/locations", h.getLocations).Methods("GET")
		m.HandleFunc("/transformers/locations/__count", h.getCount).Methods("GET")
		m.HandleFunc("/transformers/locations/__ids", h.getIds).Methods("GET")
		m.HandleFunc("/transformers/locations/__reload", h.reload).Methods("POST")
		m.HandleFunc("/transformers/locations/{uuid:([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})}", h.getLocationByUUID).Methods("GET")

		var monitoringRouter http.Handler = m
		monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), monitoringRouter)
		monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

		http.HandleFunc(status.PingPath, status.PingHandler)
		http.HandleFunc(status.PingPathDW, status.PingHandler)
		http.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
		http.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)

		http.HandleFunc("/__health", v1a.Handler("Locations Transformer Healthchecks", "Checks for accessing TME", h.HealthCheck()))
		g2gHandler := status.NewGoodToGoHandler(gtg.StatusChecker(h.G2GCheck))
		http.HandleFunc(status.GTGPath, g2gHandler)

		http.Handle("/", monitoringRouter)

		log.Printf("listening on %d", *port)
		err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
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
