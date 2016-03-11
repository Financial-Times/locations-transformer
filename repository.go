package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

const DefaultMaxRecords = 10000
const DefaultSlices = 10
const TaxonomyName = "GL"

type repository interface {
	getLocationsTaxonomy(int) (taxonomy, error)
	getSingleLocationTaxonomy(string) (term, error)
	MaxRecords() int
}

type tmeRepository struct {
	httpClient   httpClient
	tmeBaseURL   string
	accessConfig tmeAccessConfig
	maxRecords   int
	slices       int
}

type tmeAccessConfig struct {
	userName string
	password string
	token    string
}

func (t *tmeRepository) MaxRecords() int {
	return t.maxRecords
}

func newTmeRepository(client httpClient, tmeBaseURL string, userName string, password string, token string, maxRecords int, slices int) repository {
	return &tmeRepository{httpClient: client, tmeBaseURL: tmeBaseURL, accessConfig: tmeAccessConfig{userName: userName, password: password, token: token}, maxRecords: maxRecords, slices: slices}
}

func (t *tmeRepository) getLocationsTaxonomy(startRecord int) (taxonomy, error) {
	chunks := t.maxRecords / t.slices
	chanResponse := make(chan *response, t.slices)
	go func() {
		var wg sync.WaitGroup
		wg.Add(t.slices)
		for i := 0; i < t.slices; i++ {
			startPosition := startRecord + i*chunks

			go func(startPosition int) {
				tax, err := t.getLocationsInChunks(startPosition, chunks)

				chanResponse <- &response{Taxonomy: tax, Err: err}
				wg.Done()
			}(startPosition)
		}
		wg.Wait()

		close(chanResponse)
	}()
	terms := make([]term, 0, t.maxRecords)
	var err error = nil
	for resp := range chanResponse {
		terms = append(terms, resp.Taxonomy.Terms...)
		if resp.Err != nil {
			err = resp.Err
		}
	}
	return taxonomy{Terms: terms}, err
}

func (t *tmeRepository) getLocationsInChunks(startPosition int, maxRecords int) (taxonomy, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rs/authorityfiles/GL/terms?maximumRecords=%d&startRecord=%d", t.tmeBaseURL, maxRecords, startPosition), nil)
	if err != nil {
		return taxonomy{}, err
	}
	req.Header.Add("Accept", "application/xml;charset=utf-8")
	req.SetBasicAuth(t.accessConfig.userName, t.accessConfig.password)
	req.Header.Add("X-Coco-Auth", fmt.Sprintf("%v", t.accessConfig.token))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return taxonomy{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return taxonomy{}, fmt.Errorf("TME returned %d", resp.StatusCode)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return taxonomy{}, err
	}

	tax := taxonomy{}
	err = xml.Unmarshal(contents, &tax)
	if err != nil {
		return taxonomy{}, err
	}
	return tax, nil
}

func (t *tmeRepository) getSingleLocationTaxonomy(rawId string) (term, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rs/authorityfiles/GL/terms/%s", t.tmeBaseURL, rawId), nil)
	if err != nil {
		return term{}, err
	}
	req.Header.Add("Accept", "application/xml;charset=utf-8")
	req.SetBasicAuth(t.accessConfig.userName, t.accessConfig.password)
	req.Header.Add("X-Coco-Auth", fmt.Sprintf("%v", t.accessConfig.token))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return term{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return term{}, fmt.Errorf("TME returned %d HTTP status", resp.StatusCode)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return term{}, err
	}

	locationTerm := term{}
	err = xml.Unmarshal(contents, &locationTerm)
	if err != nil {
		return term{}, err
	}
	return locationTerm, nil
}
