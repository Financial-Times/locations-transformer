package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const testUUID = "bba39990-c78d-3629-ae83-808c333c6dbc"
const getLocationsResponse = `[{"apiUrl":"http://localhost:8080/transformers/locations/bba39990-c78d-3629-ae83-808c333c6dbc"}]`
const getLocationByUUIDResponse = `{"uuid":"bba39990-c78d-3629-ae83-808c333c6dbc","alternativeIdentifiers":{"TME":["MTE3-U3ViamVjdHM="],"uuids":["bba39990-c78d-3629-ae83-808c333c6dbc"]},"prefLabel":"SomeLocation","type":"Location","types":["Thing","Concept","Location"]}`
const getLocationsCountResponse = `1`
const getLocationsIdsResponse = `{"id":"bba39990-c78d-3629-ae83-808c333c6dbc"}`

func TestHandlers(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name         string
		req          *http.Request
		dummyService locationService
		statusCode   int
		contentType  string // Contents of the Content-Type header
		body         string
	}{
		{"Success - get location by uuid", newRequest("GET", fmt.Sprintf("/transformers/locations/%s", testUUID)), &dummyService{found: true, locations: []location{getDummyLocation(testUUID, "SomeLocation", "MTE3-U3ViamVjdHM=")}}, http.StatusOK, "application/json", getLocationByUUIDResponse},
		{"Not found - get location by uuid", newRequest("GET", fmt.Sprintf("/transformers/locations/%s", testUUID)), &dummyService{found: false, locations: []location{location{}}}, http.StatusNotFound, "application/json", ""},
		{"Success - get locations", newRequest("GET", "/transformers/locations"), &dummyService{found: true, locations: []location{location{UUID: testUUID}}}, http.StatusOK, "application/json", getLocationsResponse},
		{"Not found - get locations", newRequest("GET", "/transformers/locations"), &dummyService{found: false, locations: []location{}}, http.StatusNotFound, "application/json", ""},
		{"Test Location Count", newRequest("GET", "/transformers/locations/__count"), &dummyService{found: true, locations: []location{location{UUID: testUUID}}}, http.StatusOK, "text/plain", getLocationsCountResponse},
		{"Test Location Ids", newRequest("GET", "/transformers/locations/__ids"), &dummyService{found: true, locations: []location{location{UUID: testUUID}}}, http.StatusOK, "text/plain", getLocationsIdsResponse},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(test.dummyService).ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.Equal(strings.TrimSpace(test.body), strings.TrimSpace(rec.Body.String()), fmt.Sprintf("%s: Wrong body", test.name))
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
	return m
}

type dummyService struct {
	found     bool
	locations []location
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

func (s *dummyService) checkConnectivity() error {
	return nil
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
