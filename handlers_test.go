package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testUUID = "bba39990-c78d-3629-ae83-808c333c6dbc"
const getLocationsResponse = "[{\"apiUrl\":\"http://localhost:8080/transformers/locations/bba39990-c78d-3629-ae83-808c333c6dbc\"}]\n"
const getLocationByUUIDResponse = "{\"uuid\":\"bba39990-c78d-3629-ae83-808c333c6dbc\",\"canonicalName\":\"Metals Markets\",\"tmeIdentifier\":\"MTE3-U3ViamVjdHM=\",\"type\":\"Location\"}\n"

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
		{"Success - get location by uuid", newRequest("GET", fmt.Sprintf("/transformers/locations/%s", testUUID)), &dummyService{found: true, locations: []location{location{UUID: testUUID, CanonicalName: "Metals Markets", TmeIdentifier: "MTE3-U3ViamVjdHM=", Type: "Location"}}}, http.StatusOK, "application/json", getLocationByUUIDResponse},
		{"Not found - get location by uuid", newRequest("GET", fmt.Sprintf("/transformers/locations/%s", testUUID)), &dummyService{found: false, locations: []location{location{}}}, http.StatusNotFound, "application/json", ""},
		{"Success - get locations", newRequest("GET", "/transformers/locations"), &dummyService{found: true, locations: []location{location{UUID: testUUID}}}, http.StatusOK, "application/json", getLocationsResponse},
		{"Not found - get locations", newRequest("GET", "/transformers/locations"), &dummyService{found: false, locations: []location{}}, http.StatusNotFound, "application/json", ""},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(test.dummyService).ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.Equal(test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
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
