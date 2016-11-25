package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
			[]term{term{CanonicalName: "test_location", RawID: "b8337559-ac08-3404-9025-bad51ebe2fc7"}, term{CanonicalName: "Feature", RawID: "mNGQ2MWQ0NDMtMDc5Mi00NWExLTlkMGQtNWZhZjk0NGExOWU2-Z2VucVz"}},
			[]locationLink{locationLink{APIURL: "localhost:8080/transformers/locations/e559b6c0-2241-35b9-b970-e55cb8be4cba"},
				locationLink{APIURL: "localhost:8080/transformers/locations/ab4861b5-ba5e-3b67-9871-3bb3e52db103"}}, true, nil},
		{"Error on init", "localhost:8080/transformers/locations/", []term{}, []locationLink(nil), false, errors.New("Error getting taxonomy")},
	}

	for _, test := range tests {
		repo := dummyRepo{terms: test.terms, err: test.err}
		service, err := newLocationService(&repo, test.baseURL, "Locations", 10000)
		expectedLocations, found := service.getLocations()
		assert.Equal(test.locations, expectedLocations, fmt.Sprintf("%s: Expected locations link incorrect", test.name))
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
		{"Success", []term{term{CanonicalName: "Test_location", RawID: "b8337559-ac08-3404-9025-bad51ebe2fc7"}, term{CanonicalName: "Test_location", RawID: "NGQ2MWQ0NDMtMDc5Mi00NWExLTlkMGQtNWZhZjk0NGExOWU2-Z2VucmVz"}},
			"f7de594e-daa7-3d0e-a997-da4440d0c3b6", getDummyLocation("f7de594e-daa7-3d0e-a997-da4440d0c3b6", "Test_location", "TkdRMk1XUTBORE10TURjNU1pMDBOV0V4TFRsa01HUXROV1poWmprME5HRXhPV1UyLVoyVnVjbVZ6-R0w="), true, nil},
		{"Not found", []term{term{CanonicalName: "Test_location", RawID: "845dc7d7-ae89-4fed-a819-9edcbb3fe507"}, term{CanonicalName: "Feature", RawID: "NGQ2MWdefsdfsfcmVz"}},
			"some uuid", location{}, false, nil},
		{"Error on init", []term{}, "some uuid", location{}, false, errors.New("Error getting taxonomy")},
	}

	for _, test := range tests {
		repo := dummyRepo{terms: test.terms, err: test.err}
		service, err := newLocationService(&repo, "", "GL", 10000)
		expectedLocation, found := service.getLocationByUUID(test.uuid)
		assert.Equal(test.location, expectedLocation, fmt.Sprintf("%s: Expected location incorrect", test.name))
		assert.Equal(test.found, found)
		assert.Equal(test.err, err)
	}
}

type dummyRepo struct {
	terms []term
	err   error
}

func (d *dummyRepo) GetTmeTermsFromIndex(startRecord int) ([]interface{}, error) {
	if startRecord > 0 {
		return nil, d.err
	}
	var interfaces = make([]interface{}, len(d.terms))
	for i, data := range d.terms {
		interfaces[i] = data
	}
	return interfaces, d.err
}
func (d *dummyRepo) GetTmeTermById(uuid string) (interface{}, error) {
	return d.terms[0], d.err
}

func getDummyLocation(uuid string, prefLabel string, tmeID string) location {
	return location{
		UUID:                   uuid,
		PrefLabel:              prefLabel,
		PrimaryType:            primaryType,
		TypeHierarchy:          locationTypes,
		AlternativeIdentifiers: alternativeIdentifiers{TME: []string{tmeID}, Uuids: []string{uuid}}}
}
