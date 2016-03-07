package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

const MaxRecords = 10000
const slices = 10
const chunks = MaxRecords / slices
const TaxonomyName = "GL"

type repository interface {
	getLocationsTaxonomy(int) (taxonomy, error)
	getSingleLocationTaxonomy(string) (term, error)
}

type tmeRepository struct {
	httpClient httpClient
	tmeBaseURL string
	userName   string
	password   string
}

func newTmeRepository(client httpClient, tmeBaseURL string, userName string, password string) repository {
	return &tmeRepository{httpClient: client, tmeBaseURL: tmeBaseURL, userName: userName, password: password}
}

func (t *tmeRepository) getLocationsTaxonomy(startRecord int) (taxonomy, error) {
	chanResponse := make(chan *response, slices)
	go func() {
		var wg sync.WaitGroup
		wg.Add(slices)
		for i := 0; i < slices; i++ {
			startPosition := startRecord + i * chunks

			go func(startPosition int) {
				tax, err := t.getLocationsInChunks(startPosition, chunks)
				chanResponse <- &response{Taxonomy:tax, Err: err}
				wg.Done()
			}(startPosition)
		}
		wg.Wait()

		close(chanResponse)
	}()
	terms := make([]term, 0, MaxRecords)
	var err error = nil
	for resp := range chanResponse {
		terms = append(terms, resp.Taxonomy.Terms...)
		if (resp.Err != nil) {
			err = resp.Err
		}
	}
	return taxonomy{Terms:terms}, err
}

func (t *tmeRepository) getLocationsInChunks(startPosition int, maxRecords int) (taxonomy, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rs/authorityfiles/GL/terms?maximumRecords=%d&startRecord=%d", t.tmeBaseURL, maxRecords, startPosition), nil)
	if err != nil {
		return taxonomy{}, err
	}
	req.Header.Add("Accept", "application/xml;charset=utf-8")
	req.SetBasicAuth(t.userName, t.password)

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
	req.SetBasicAuth(t.userName, t.password)
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
