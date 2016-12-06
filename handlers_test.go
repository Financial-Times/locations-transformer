package main

import (
	"fmt"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/service-status-go/gtg"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

const (
	testUUID                  = "bba39990-c78d-3629-ae83-808c333c6dbc"
	getLocationsResponse      = `[{"apiUrl":"http://localhost:8080/transformers/locations/bba39990-c78d-3629-ae83-808c333c6dbc"}]`
	getLocationByUUIDResponse = `{"uuid":"bba39990-c78d-3629-ae83-808c333c6dbc","alternativeIdentifiers":{"TME":["MTE3-U3ViamVjdHM="],"uuids":["bba39990-c78d-3629-ae83-808c333c6dbc"]},"prefLabel":"SomeLocation","type":"Location"}`
	getLocationsCountResponse = `1`
	getLocationsIdsResponse   = `{"id":"bba39990-c78d-3629-ae83-808c333c6dbc"}`
)

func TestHandlers(t *testing.T) {
	tests := []struct {
		name         string
		req          *http.Request
		dummyService locationService
		statusCode   int
		contentType  string // Contents of the Content-Type header
		body         string
	}{
		{"Success - get location by uuid", newRequest("GET", fmt.Sprintf("/transformers/locations/%s", testUUID)), &dummyService{found: true, locations: []location{getDummyLocation(testUUID, "SomeLocation", "MTE3-U3ViamVjdHM=")}}, http.StatusOK, "application/json", getLocationByUUIDResponse},
		{"Not found - get location by uuid", newRequest("GET", fmt.Sprintf("/transformers/locations/%s", testUUID)), &dummyService{found: false, locations: []location{{}}}, http.StatusNotFound, "application/json", ""},
		{"Success - get locations", newRequest("GET", "/transformers/locations"), &dummyService{found: true, locations: []location{{UUID: testUUID}}}, http.StatusOK, "application/json", getLocationsResponse},
		{"Not found - get locations", newRequest("GET", "/transformers/locations"), &dummyService{found: false, locations: []location{}}, http.StatusNotFound, "application/json", ""},
		{"Test Location Count", newRequest("GET", "/transformers/locations/__count"), &dummyService{found: true, locations: []location{{UUID: testUUID}}}, http.StatusOK, "text/plain", getLocationsCountResponse},
		{"Test Location Ids", newRequest("GET", "/transformers/locations/__ids"), &dummyService{found: true, locations: []location{{UUID: testUUID}}}, http.StatusOK, "text/plain", getLocationsIdsResponse},
		{"Test GTG - Pass", newRequest("GET", status.GTGPath), &dummyService{found: true, locations: []location{{UUID: testUUID}}}, http.StatusOK, "application/json", "OK"},
		{"Test GTG - Fail", newRequest("GET", status.GTGPath), &dummyService{found: true, locations: []location(nil)}, http.StatusServiceUnavailable, "application/json", ""},
		{"Reload - Good", newRequest("POST", "/transformers/locations/__reload"), &dummyService{dataLoaded: DataLoaded}, http.StatusAccepted, "application/json", "{\"message\": \"Reloading people\"}"},
		{"Reload - Conflict", newRequest("POST", "/transformers/locations/__reload"), &dummyService{dataLoaded: LoadingData}, http.StatusConflict, "application/json", "{\"message\": \"Currently Loading Data\"}"},
		{"Reload - Fail", newRequest("POST", "/transformers/locations/__reload"), &dummyService{dataLoaded: NotInit}, http.StatusServiceUnavailable, "application/json", "{\"message\": \"Service Unavailable\"}"},
		{"Health - Good", newRequest("GET", "/__health"), &dummyService{dataLoaded: DataLoaded}, http.StatusOK, "application/json", "regex=Check connectivity to TME\",\"ok\":true"},
		{"Health - Bad", newRequest("GET", "/__health"), &dummyService{dataLoaded: ErrorLoadingData}, http.StatusOK, "application/json", "regex=Got an error loading data from tme. Check logs"},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(test.dummyService).ServeHTTP(rec, test.req)
		assert.True(t, test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))

		if strings.HasPrefix(test.body, "regex=") {
			regex := strings.TrimPrefix(test.body, "regex=")
			body := rec.Body.String()
			matched, err := regexp.MatchString(regex, body)
			assert.NoError(t, err)
			assert.True(t, matched, fmt.Sprintf("Could not match regex:\n %s \nin body:\n %s", regex, body))
		} else {
			assert.Equal(t, strings.TrimSpace(test.body), strings.TrimSpace(rec.Body.String()), fmt.Sprintf("%s: Wrong body", test.name))
		}
	}
}

func newRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return req
}

func router(s locationService) *mux.Router {
	m := mux.NewRouter()
	h := newLocationsHandler(s)
	m.HandleFunc("/transformers/locations", h.getLocations).Methods("GET")
	m.HandleFunc("/transformers/locations/__ids", h.getIds).Methods("GET")
	m.HandleFunc("/transformers/locations/__count", h.getCount).Methods("GET")
	m.HandleFunc("/transformers/locations/__reload", h.reload).Methods("POST")
	m.HandleFunc("/transformers/locations/{uuid}", h.getLocationByUUID).Methods("GET")
	g2gHandler := status.NewGoodToGoHandler(gtg.StatusChecker(h.G2GCheck))
	m.HandleFunc(status.GTGPath, g2gHandler)
	m.HandleFunc("/__health", v1a.Handler("Locations Transformer Healthchecks", "Checks for accessing TME", h.HealthCheck()))
	return m
}

type dummyService struct {
	found       bool
	locations   []location
	initialised bool
	dataLoaded  loadStatus
}

func (s *dummyService) getLocations() ([]locationLink, bool) {
	var locationLinks []locationLink
	for _, sub := range s.locations {
		locationLinks = append(locationLinks, locationLink{APIURL: "http://localhost:8080/transformers/locations/" + sub.UUID})
	}
	return locationLinks, s.found
}

func (s *dummyService) getLocationByUUID(uuid string) (location, bool) {
	return s.locations[0], s.found
}

func (s *dummyService) getLocationCount() int {
	return len(s.locations)
}

func (s *dummyService) getLocationIds() []string {
	i := 0
	keys := make([]string, len(s.locations))

	for _, t := range s.locations {
		keys[i] = t.UUID
		i++
	}
	return keys
}

func (s *dummyService) reload() error {
	return nil
}

func (s *dummyService) getLoadStatus() loadStatus {
	return s.dataLoaded
}
