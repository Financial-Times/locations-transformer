package main

import (
	"github.com/Financial-Times/tme-reader"
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
	repository    tme.Repository
	baseURL       string
	IdMap         map[string]string
	locationLinks []locationLink
	taxonomyName  string
	maxTmeRecords int
}

func newLocationService(repo tme.Repository, baseURL string, taxonomyName string, maxTmeRecords int) (locationService, error) {

	s := &locationServiceImpl{repository: repo, baseURL: baseURL, taxonomyName: taxonomyName, maxTmeRecords: maxTmeRecords}
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
		taxonomyContent, err := s.repository.GetTmeTermsFromIndex(responseCount)
		if err != nil {
			return err
		}
		taxonomy, err := readTmeTermsTaxonomy(taxonomyContent)

		if len(taxonomy.Terms) < 1 {
			log.Printf("Finished fetching locations from TME\n")
			break
		}
		s.initLocationsMap(taxonomy.Terms)
		responseCount += s.maxTmeRecords
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
	content, err := s.repository.GetTmeTermById(rawId)
	if err != nil {
		return location{}, false
	}
	return transformLocation(content, s.taxonomyName), true
}

func (s *locationServiceImpl) initLocationsMap(terms []term) {
	for _, t := range terms {
		tmeIdentifier := buildTmeIdentifier(t.RawID, s.taxonomyName)
		uuid := uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String()
		s.IdMap[uuid] = t.RawID
		s.locationLinks = append(s.locationLinks, locationLink{APIURL: s.baseURL + uuid})
	}
}
