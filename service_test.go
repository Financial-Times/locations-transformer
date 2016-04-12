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
		tax       taxonomy
		locations []locationLink
		found     bool
		err       error
	}{
		{"Success", "localhost:8080/transformers/locations/",
			taxonomy{Terms: []term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}}},
			[]locationLink{locationLink{APIURL: "localhost:8080/transformers/locations/095b89cd-4d4c-3195-ba78-e366fbe47291"}}, true, nil},
		{"Error on init", "localhost:8080/transformers/locations/", taxonomy{}, []locationLink(nil), false, errors.New("Error getting taxonomy")},
	}

	for _, test := range tests {
		repo := dummyRepo{tax: test.tax, err: test.err}
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
		tax      taxonomy
		uuid     string
		location location
		found    bool
		err      error
	}{
		{"Success", taxonomy{Terms: []term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}}},
			"095b89cd-4d4c-3195-ba78-e366fbe47291", location{UUID: "095b89cd-4d4c-3195-ba78-e366fbe47291", CanonicalName: "Banksville, New York", TmeIdentifier: "TnN0ZWluX0dMX1VTX05ZX011bmljaXBhbGl0eV85NDI5Njg=-R0w=", Type: "Location"}, true, nil},
		{"Not found", taxonomy{Terms: []term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}}},
			"some uuid", location{}, false, nil},
		{"Error on init", taxonomy{}, "some uuid", location{}, false, errors.New("Error getting taxonomy")},
	}
	for _, test := range tests {
		repo := dummyRepo{tax: test.tax, err: test.err}
		service, err := newLocationService(&repo, "", "GL", 10000)
		actualLocation, found := service.getLocationByUUID(test.uuid)
		assert.Equal(test.location, actualLocation, fmt.Sprintf("%s: Expected location incorrect", test.name))
		assert.Equal(test.found, found)
		assert.Equal(test.err, err)
	}
}

type dummyRepo struct {
	tax taxonomy
	err error
}

func (d *dummyRepo) GetTmeTermsFromIndex(startRecord int) ([]byte, error) {
	if startRecord > 0 {
		return []byte{}, d.err
	}
	bytes, err := ToByte(d.tax)
	if err != nil {
		return bytes, err
	}
	return bytes, d.err
}

func (t *dummyRepo) GetTmeTermsInChunks(startPosition int, maxRecords int) ([]byte, error) {
	return []byte{}, nil
}

func (d *dummyRepo) GetTmeTermById(uuid string) ([]byte, error) {
	bytes, err := ToByte(d.tax.Terms[0])
	if err != nil {
		return bytes, err
	}
	return bytes, d.err
}
