package main

import (
	"github.com/Financial-Times/tme-reader/tmereader"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"sync"
	"sync/atomic"
)

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type locationService interface {
	getLocations() ([]locationLink, bool)
	getLocationByUUID(uuid string) (location, bool)
	checkConnectivity() error
	getLocationCount() int
	getLocationIds() []string
	reload() error
	getLoadStatus() loadStatus
}

type loadStatus string

const (
	NotInit          = loadStatus("NotInit")
	LoadingData      = loadStatus("Loading")
	DataLoaded       = loadStatus("DataLoaded")
	ErrorLoadingData = loadStatus("ErrorLoadingData")
)

type locationServiceImpl struct {
	sync.Mutex
	repository    tmereader.Repository
	baseURL       string
	locationsMap  atomic.Value
	locationLinks atomic.Value
	taxonomyName  string
	maxTmeRecords int
	status        atomic.Value
}

type locationsMap map[string]location
type locationLinks []locationLink

func (s *locationServiceImpl) getLoadStatus() loadStatus {
	i := s.status.Load()
	if i == nil {
		return NotInit
	}
	return i.(loadStatus)
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
	val := s.locationLinks.Load()
	if val == nil {
		return nil, false
	}

	tmp := val.(locationLinks)
	if len(tmp) > 0 {
		return tmp, true
	}
	return tmp, false
}

func (s *locationServiceImpl) getLocationByUUID(uuid string) (location, bool) {
	val := s.locationsMap.Load()
	if val != nil {
		location, found := val.(locationsMap)[uuid]
		return location, found
	}
	return location{}, false
}

func (s *locationServiceImpl) checkConnectivity() error {
	// TODO: Can we just hit an endpoint to check if TME is available? Or do we need to make sure we get location taxonmies back?
	//	_, err := s.repository.GetTmeTermsFromIndex()
	//	if err != nil {
	//		return err
	//	}
	return nil
}

func (s *locationServiceImpl) initLocationsMap(terms []interface{}) (map[string]location, []locationLink) {
	lMap := make(map[string]location)
	ll := make([]locationLink, len(terms))
	for i, iTerm := range terms {
		t := iTerm.(term)
		top := transformLocation(t, s.taxonomyName)
		lMap[top.UUID] = top
		ll[i] = locationLink{APIURL: s.baseURL + top.UUID}
	}
	return lMap, ll
}

func (s *locationServiceImpl) getLocationCount() int {
	val := s.locationLinks.Load()
	if val == nil {
		return 0
	}
	return len(val.(locationLinks))
}

func (s *locationServiceImpl) getLocationIds() []string {
	i := 0
	val := s.locationsMap.Load()

	if val == nil {
		return make([]string, i)
	}

	lm := val.(locationsMap)
	keys := make([]string, len(lm))

	for k := range lm {
		keys[i] = k
		i++
	}
	return keys
}

func (s *locationServiceImpl) reload() error {
	s.Lock() // lock as updating the stores
	defer s.Unlock()
	s.locationsMap.Store(make(locationsMap))
	s.locationLinks.Store(make(locationLinks, 0))
	s.status.Store(LoadingData)
	responseCount := 0
	log.Println("Fetching locations from TME")

	tempLocationsMap := make(locationsMap)
	tempLocationLinks := make(locationLinks, 0)
	for {
		terms, err := s.repository.GetTmeTermsFromIndex(responseCount)
		if err != nil {
			log.Warnf("Got an error loading data from tme '%v'", err)
			s.status.Store(ErrorLoadingData)
			return err
		}

		tc := len(terms)
		if tc < 1 {
			log.Info("Finished fetching locations from TME")
			break
		}
		log.Infof("Processing '%v' terms", tc)

		lMap, ll := s.initLocationsMap(terms)

		tempLocationLinks = append(tempLocationLinks, ll...)
		for k, v := range lMap {
			tempLocationsMap[k] = v
		}

		responseCount += s.maxTmeRecords
	}
	s.locationsMap.Store(tempLocationsMap)
	s.locationLinks.Store(tempLocationLinks)
	s.status.Store(DataLoaded)
	log.Infof("Added %d location links\n", s.getLocationCount())
	return nil
}
