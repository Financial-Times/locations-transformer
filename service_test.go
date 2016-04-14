package main

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLocations(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name      string
		baseURL   string
		terms     []term
		locations []locationLink
		found     bool
		err       error
	}{
		{"Success", "localhost:8080/transformers/locations/",
			[]term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
			[]locationLink{locationLink{APIURL: "localhost:8080/transformers/locations/095b89cd-4d4c-3195-ba78-e366fbe47291"}}, true, nil},
		{"Error on init", "localhost:8080/transformers/locations/", []term{}, []locationLink(nil), false, errors.New("Error getting taxonomy")},
	}

	for _, test := range tests {
		repo := dummyRepo{terms: test.terms, err: test.err}
		service, err := newLocationService(&repo, test.baseURL, "GL", 10000)
		actualLocations, found := service.getLocations()
		assert.Equal(test.locations, actualLocations, fmt.Sprintf("%s: Expected locations link incorrect", test.name))
		assert.Equal(test.found, found)
		assert.Equal(test.err, err)
	}
}

func TestGetLocationByUuid(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name     string
		terms    []term
		uuid     string
		location location
		found    bool
		err      error
	}{
		{"Success", []term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
			"095b89cd-4d4c-3195-ba78-e366fbe47291", location{UUID: "095b89cd-4d4c-3195-ba78-e366fbe47291", CanonicalName: "Banksville, New York", TmeIdentifier: "TnN0ZWluX0dMX1VTX05ZX011bmljaXBhbGl0eV85NDI5Njg=-R0w=", Type: "Location"}, true, nil},
		{"Not found", []term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
			"some uuid", location{}, false, nil},
		{"Error on init", []term{}, "some uuid", location{}, false, errors.New("Error getting taxonomy")},
	}
	for _, test := range tests {
		repo := dummyRepo{terms: test.terms, err: test.err}
		service, err := newLocationService(&repo, "", "GL", 10000)
		actualLocation, found := service.getLocationByUUID(test.uuid)
		assert.Equal(test.location, actualLocation, fmt.Sprintf("%s: Expected location incorrect", test.name))
		assert.Equal(test.found, found)
		assert.Equal(test.err, err)
	}
}

type dummyRepo struct {
	terms []term
	err error
}

func (d *dummyRepo) GetTmeTermsFromIndex(startRecord int) ([]interface{}, error) {
	if startRecord > 0 {
		return nil, d.err
	}
	var interfaces []interface{} = make([]interface{}, len(d.terms))
	for i, data := range d.terms {
		interfaces[i] = data
	}
	return interfaces, d.err
}
func (d *dummyRepo) GetTmeTermById(uuid string) (interface{}, error) {
	return d.terms[0], d.err
}
