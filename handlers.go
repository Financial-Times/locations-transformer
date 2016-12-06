package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/service-status-go/gtg"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type locationsHandler struct {
	service locationService
}

// HealthCheck does something
func (h *locationsHandler) HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Unable to respond to request for the location data from TME",
		Name:             "Check connectivity to TME",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/locations-transfomer",
		Severity:         1,
		TechnicalSummary: "Cannot connect to TME to be able to supply locations",
		Checker:          h.checker,
	}
}

func (h *locationsHandler) G2GCheck() gtg.Status {
	count := h.service.getLocationCount()
	if count > 0 {
		return gtg.Status{GoodToGo: true}
	}
	return gtg.Status{GoodToGo: false}
}

func (h *locationsHandler) checker() (string, error) {
	if ls := h.service.getLoadStatus(); ls == ErrorLoadingData {
		return "Error connecting to TME", errors.New("Got an error loading data from tme. Check logs.")
	}
	return "Connectivity to TME is ok", nil
}

func newLocationsHandler(service locationService) locationsHandler {
	return locationsHandler{service: service}
}

func (h *locationsHandler) getLocations(writer http.ResponseWriter, req *http.Request) {
	obj, found := h.service.getLocations()
	writeJSONResponse(obj, found, writer)
}

func (h *locationsHandler) getCount(writer http.ResponseWriter, req *http.Request) {
	count := h.service.getLocationCount()
	_, err := writer.Write([]byte(strconv.Itoa(count)))
	if err != nil {
		log.Warnf("Couldn't write count to HTTP response. count=%d %v\n", count, err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *locationsHandler) getIds(writer http.ResponseWriter, req *http.Request) {
	ids := h.service.getLocationIds()
	writer.Header().Add("Content-Type", "text/plain")
	if len(ids) == 0 {
		writer.WriteHeader(http.StatusOK)
		return
	}
	enc := json.NewEncoder(writer)
	type locationId struct {
		ID string `json:"id"`
	}
	for _, id := range ids {
		rID := locationId{ID: id}
		err := enc.Encode(rID)
		if err != nil {
			log.Warnf("Couldn't encode to HTTP response location with uuid=%s %v\n", id, err)
			continue
		}
	}
}

func (h *locationsHandler) reload(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Add("Content-Type", "application/json")
	st := h.service.getLoadStatus()
	if st == NotInit {
		writeJSONError(writer, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	if st == LoadingData {
		writeJSONError(writer, "Currently Loading Data", http.StatusConflict)
		return
	}
	go func() {
		err := h.service.reload()
		if err != nil {
			log.Warnf("Problem reloading terms from TME: %v", err)
		}
	}()
	writeJSONError(writer, "Reloading people", http.StatusAccepted)
}

func (h *locationsHandler) getLocationByUUID(writer http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uuid := vars["uuid"]

	obj, found := h.service.getLocationByUUID(uuid)
	writeJSONResponse(obj, found, writer)
}

func writeJSONResponse(obj interface{}, found bool, writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")

	if !found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	enc := json.NewEncoder(writer)
	if err := enc.Encode(obj); err != nil {
		log.Errorf("Error on json encoding=%v\n", err)
		writeJSONError(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeJSONError(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, fmt.Sprintf("{\"message\": \"%s\"}", errorMsg))
}
