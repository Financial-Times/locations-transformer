package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestGetLocationsTaxonomy(t *testing.T) {
	assert := assert.New(t)
	locationsXML, err := os.Open("sample_locations.xml")
	log.Printf("%v\n", err)
	tests := []struct {
		name string
		repo repository
		tax  taxonomy
		err  error
	}{
		{"Success", repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(locationsXML)}}),
			taxonomy{Terms: []term{
				term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}}}, nil},
		{"Error", repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(locationsXML)}, err: errors.New("Some error")}),
			taxonomy{Terms: []term{}}, errors.New("Some error")},
		{"Non 200 from structure service", repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusBadRequest, Body: ioutil.NopCloser(locationsXML)}}),
			taxonomy{Terms: []term{}}, errors.New("TME returned 400")},
		{"Unmarshalling error", repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewReader([]byte("Non xml")))}}),
			taxonomy{Terms: []term{}}, errors.New("EOF")},
	}

	for _, test := range tests {
		expectedTax, err := test.repo.getLocationsTaxonomy(0)
		assert.Equal(test.tax, expectedTax, fmt.Sprintf("%s: Expected taxonomy incorrect", test.name))
		assert.Equal(test.err, err)
	}

}

func repo(c dummyClient) repository {
	return &tmeRepository{httpClient: &c, tmeBaseURL: c.tmeBaseURL, accessConfig: tmeAccessConfig{userName: "test", password: "test", token: "test"}, maxRecords: 1, slices: 1}
}

type dummyClient struct {
	assert     *assert.Assertions
	resp       http.Response
	err        error
	tmeBaseURL string
}

func (d *dummyClient) Do(req *http.Request) (resp *http.Response, err error) {
	d.assert.Contains(req.URL.String(), fmt.Sprintf("%s/rs/authorityfiles/GL/terms?maximumRecords=", d.tmeBaseURL), fmt.Sprintf("Expected url incorrect"))
	return &d.resp, d.err
}
