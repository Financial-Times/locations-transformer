package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestGetLocations(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		terms     []term
		locations []locationLink
		found     bool
		err       error
	}{
		{"Success", "localhost:8080/transformers/locations/",
			[]term{{CanonicalName: "test_location", RawID: "b8337559-ac08-3404-9025-bad51ebe2fc7"}, {CanonicalName: "Feature", RawID: "mNGQ2MWQ0NDMtMDc5Mi00NWExLTlkMGQtNWZhZjk0NGExOWU2-Z2VucVz"}},
			[]locationLink{{APIURL: "localhost:8080/transformers/locations/e559b6c0-2241-35b9-b970-e55cb8be4cba"},
				{APIURL: "localhost:8080/transformers/locations/ab4861b5-ba5e-3b67-9871-3bb3e52db103"}}, true, nil},
		{"Error on init", "localhost:8080/transformers/locations/", []term{}, []locationLink(nil), false, errors.New("Error getting taxonomy")},
	}

	for _, test := range tests {
		repo := dummyRepo{terms: test.terms, err: test.err}
		service, err := newLocationService(&repo, test.baseURL, "Locations", 10000)
		expectedLocations, found := service.getLocations()
		assert.Equal(t, test.locations, expectedLocations, fmt.Sprintf("%s: Expected locations link incorrect", test.name))
		assert.Equal(t, test.found, found)
		assert.Equal(t, test.err, err)
	}
}

func TestGetLocationByUuid(t *testing.T) {
	tests := []struct {
		name     string
		terms    []term
		uuid     string
		location location
		found    bool
		err      error
	}{
		{"Success", []term{{CanonicalName: "Test_location", RawID: "b8337559-ac08-3404-9025-bad51ebe2fc7"}, {CanonicalName: "Test_location", RawID: "NGQ2MWQ0NDMtMDc5Mi00NWExLTlkMGQtNWZhZjk0NGExOWU2-Z2VucmVz"}},
			"f7de594e-daa7-3d0e-a997-da4440d0c3b6", getDummyLocation("f7de594e-daa7-3d0e-a997-da4440d0c3b6", "Test_location", "TkdRMk1XUTBORE10TURjNU1pMDBOV0V4TFRsa01HUXROV1poWmprME5HRXhPV1UyLVoyVnVjbVZ6-R0w="), true, nil},
		{"Not found", []term{{CanonicalName: "Test_location", RawID: "845dc7d7-ae89-4fed-a819-9edcbb3fe507"}, {CanonicalName: "Feature", RawID: "NGQ2MWdefsdfsfcmVz"}},
			"some uuid", location{}, false, nil},
		{"Error on init", []term{}, "some uuid", location{}, false, errors.New("Error getting taxonomy")},
	}

	for _, test := range tests {
		log.Infof("Running test: %v", test.name)
		repo := dummyRepo{terms: test.terms, err: test.err}
		service, err := newLocationService(&repo, "", "GL", 10000)
		expectedLocation, found := service.getLocationByUUID(test.uuid)
		assert.Equal(t, test.location, expectedLocation, fmt.Sprintf("%s: Expected location incorrect", test.name))
		assert.Equal(t, test.found, found)
		assert.Equal(t, test.err, err)
	}
}

func TestReloadChangesStatus(t *testing.T) {
	repo := dummyLockRepo{
		terms: []term{
			{CanonicalName: "Test_location", RawID: "b8337559-ac08-3404-9025-bad51ebe2fc7"},
			{CanonicalName: "Test_location", RawID: "NGQ2MWQ0NDMtMDc5Mi00NWExLTlkMGQtNWZhZjk0NGExOWU2-Z2VucmVz"}},
		err: nil}
	service, err := newLocationService(&repo, "", "GL", 10000)
	assert.NoError(t, err)
	assert.True(t, service.isDataLoaded())
	repo.Add(1)
	go func() {
		assert.NoError(t, service.reload())
	}()

	for i := 1; i <= 1000; i++ {
		if !service.isDataLoaded() {
			log.Info("Data not loaded")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.False(t, service.isDataLoaded())
	repo.Done()

	for i := 1; i <= 1000; i++ {
		if service.isDataLoaded() {
			log.Info("Data loaded")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.True(t, service.isDataLoaded())

}

func TestGetCount(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		terms     []term
		locations []locationLink
		found     bool
		err       error
	}{
		{"Success", "localhost:8080/transformers/locations/",
			[]term{{CanonicalName: "test_location", RawID: "b8337559-ac08-3404-9025-bad51ebe2fc7"}, {CanonicalName: "Feature", RawID: "mNGQ2MWQ0NDMtMDc5Mi00NWExLTlkMGQtNWZhZjk0NGExOWU2-Z2VucVz"}},
			[]locationLink{{APIURL: "localhost:8080/transformers/locations/e559b6c0-2241-35b9-b970-e55cb8be4cba"},
				{APIURL: "localhost:8080/transformers/locations/ab4861b5-ba5e-3b67-9871-3bb3e52db103"}}, true, nil},
		{"Error on init", "localhost:8080/transformers/locations/", []term{}, []locationLink(nil), false, errors.New("Error getting taxonomy")},
	}

	for _, test := range tests {
		repo := dummyRepo{terms: test.terms, err: test.err}
		service, err := newLocationService(&repo, test.baseURL, "Locations", 10000)
		actualCount := service.getLocationCount()
		assert.Equal(t, len(test.locations), actualCount, fmt.Sprintf("%s: Expected locations count incorrect", test.name))
		assert.Equal(t, test.err, err)
	}
}

func TestReload(t *testing.T) {
	repo := dummyRepo{
		terms: []term{
			{CanonicalName: "Test_location", RawID: "b8337559-ac08-3404-9025-bad51ebe2fc7"},
			{CanonicalName: "Test_location", RawID: "NGQ2MWQ0NDMtMDc5Mi00NWExLTlkMGQtNWZhZjk0NGExOWU2-Z2VucmVz"}},
		err: nil}
	service, err := newLocationService(&repo, "", "GL", 10000)
	assert.NoError(t, err)
	assert.Equal(t, 2, service.getLocationCount())

	repo.terms = []term{
		{CanonicalName: "Test_location", RawID: "b8337559-ac08-3404-9025-bad51ebe2fc7"},
		{CanonicalName: "Test_location", RawID: "NGQ2MWQ0NDMtMDc5Mi00NWExLTlkMGQtNWZhZjk0NGExOWU2-Z2VucmVz"},
		{CanonicalName: "Test_location", RawID: "NGQ2MWQZ2VucmVz"}}

	err = service.reload()
	assert.NoError(t, err)
	assert.Equal(t, 3, service.getLocationCount())
}

type dummyLockRepo struct {
	sync.WaitGroup
	terms []term
	err   error
}

func (d *dummyLockRepo) GetTmeTermsFromIndex(startRecord int) ([]interface{}, error) {
	d.Wait()
	if startRecord > 0 {
		return nil, d.err
	}
	var interfaces []interface{} = make([]interface{}, len(d.terms))
	for i, data := range d.terms {
		interfaces[i] = data
	}
	return interfaces, d.err
}
func (d *dummyLockRepo) GetTmeTermById(uuid string) (interface{}, error) {
	return d.terms[0], d.err
}

type dummyRepo struct {
	terms []term
	err   error
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

func getDummyLocation(uuid string, prefLabel string, tmeId string) location {
	return location{
		UUID:      uuid,
		PrefLabel: prefLabel,
		Type:      "Location",
		AlternativeIdentifiers: alternativeIdentifiers{TME: []string{tmeId}, Uuids: []string{uuid}}}
}
