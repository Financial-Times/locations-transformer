package main

import (
	"github.com/Financial-Times/tme-reader/tmereader"
	log "github.com/Sirupsen/logrus"
)

type locationService interface {
	getLocations() ([]locationLink, bool)
	getLocationByUUID(uuid string) (location, bool)
	checkConnectivity() error
	getLocationCount() int
	getLocationIds() []string
	reload() error
}

type locationServiceImpl struct {
	repository    tmereader.Repository
	baseURL       string
	locationsMap  map[string]location
	locationLinks []locationLink
	taxonomyName  string
	maxTmeRecords int
}

func newLocationService(repo tmereader.Repository, baseURL string, taxonomyName string, maxTmeRecords int) (locationService, error) {
	s := &locationServiceImpl{repository: repo, baseURL: baseURL, taxonomyName: taxonomyName, maxTmeRecords: maxTmeRecords}
	err := s.reload()
	if err != nil {
		return &locationServiceImpl{}, err
	}
	return s, nil
}

func (s *locationServiceImpl) getLocations() ([]locationLink, bool) {
	if len(s.locationLinks) > 0 {
		return s.locationLinks, true
	}
	return s.locationLinks, false
}

func (s *locationServiceImpl) getLocationByUUID(uuid string) (location, bool) {
	location, found := s.locationsMap[uuid]
	return location, found
}

func (s *locationServiceImpl) checkConnectivity() error {
	// TODO: Can we just hit an endpoint to check if TME is available? Or do we need to make sure we get location taxonmies back?
	//	_, err := s.repository.GetTmeTermsFromIndex()
	//	if err != nil {
	//		return err
	//	}
	return nil
}

func (s *locationServiceImpl) initLocationsMap(terms []interface{}) {
	for _, iTerm := range terms {
		t := iTerm.(term)
		top := transformLocation(t, s.taxonomyName)
		s.locationsMap[top.UUID] = top
		s.locationLinks = append(s.locationLinks, locationLink{APIURL: s.baseURL + top.UUID})
	}
}

func (s *locationServiceImpl) getLocationCount() int {
	return len(s.locationLinks)
}

func (s *locationServiceImpl) getLocationIds() []string {
	i := 0
	keys := make([]string, len(s.locationsMap))

	for k := range s.locationsMap {
		keys[i] = k
		i++
	}
	return keys
}

func (s *locationServiceImpl) reload() error {
	s.locationsMap = make(map[string]location)
	var links []locationLink
	s.locationLinks = links
	responseCount := 0
	log.Println("Fetching locations from TME")
	for {
		terms, err := s.repository.GetTmeTermsFromIndex(responseCount)
		if err != nil {
			return err
		}

		if len(terms) < 1 {
			log.Println("Finished fetching locations from TME")
			break
		}
		s.initLocationsMap(terms)
		responseCount += s.maxTmeRecords
	}
	log.Printf("Added %d location links\n", len(s.locationLinks))

	return nil
}
