package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
)

type locationsHandler struct {
	service locationService
}

func newLocationsHandler(service locationService) locationsHandler {
	return locationsHandler{service: service}
}

func (h *locationsHandler) getLocations(writer http.ResponseWriter, req *http.Request) {
	obj, found := h.service.getLocations()
	writeJSONResponse(obj, found, writer)
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
