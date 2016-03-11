package main

import (
	"github.com/pborman/uuid"
	"log"
	"net/http"
)

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type locationService interface {
	getLocations() ([]locationLink, bool)
	getLocationByUUID(uuid string) (location, bool)
}

type locationServiceImpl struct {
	repository    repository
	baseURL       string
	IdMap         map[string]string
	locationLinks []locationLink
}

func newLocationService(repo repository, baseURL string) (locationService, error) {

	s := &locationServiceImpl{repository: repo, baseURL: baseURL}
	err := s.init()
	if err != nil {
		return &locationServiceImpl{}, err
	}
	return s, nil
}

func (s *locationServiceImpl) init() error {
	s.IdMap = make(map[string]string)
	responseCount := 0
	log.Printf("Fetching locations from TME\n")
	for {
		tax, err := s.repository.getLocationsTaxonomy(responseCount)
		if err != nil {
			return err
		}
		if len(tax.Terms) < 1 {
			log.Printf("Finished fetching locations from TME\n")
			break
		}
		s.initLocationsMap(tax.Terms)
		responseCount += s.repository.MaxRecords()
	}
	log.Printf("Added %d location links\n", len(s.locationLinks))
	return nil
}

func (s *locationServiceImpl) getLocations() ([]locationLink, bool) {
	if len(s.locationLinks) > 0 {
		return s.locationLinks, true
	}
	return s.locationLinks, false
}

func (s *locationServiceImpl) getLocationByUUID(uuid string) (location, bool) {
	rawId, found := s.IdMap[uuid]
	if !found {
		return location{}, false
	}
	term, err := s.repository.getSingleLocationTaxonomy(rawId)
	if err != nil {
		return location{}, false
	}
	return transformLocation(term), true
}

func (s *locationServiceImpl) initLocationsMap(terms []term) {
	for _, t := range terms {
		tmeIdentifier := buildTmeIdentifier(t.RawID)
		uuid := uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String()
		s.IdMap[uuid] = t.RawID
		s.locationLinks = append(s.locationLinks, locationLink{APIURL: s.baseURL + uuid})
	}
}
