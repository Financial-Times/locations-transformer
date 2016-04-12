package main

import (
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransform(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name     string
		term     term
		location location
	}{
		{"Trasform term to location", term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}, location{UUID: "095b89cd-4d4c-3195-ba78-e366fbe47291", CanonicalName: "Banksville, New York", TmeIdentifier: "TnN0ZWluX0dMX1VTX05ZX011bmljaXBhbGl0eV85NDI5Njg=-R0w=", Type: "Location"}},
	}

	for _, test := range tests {
		bytes, err := ToByte(test.term)
		assert.Equal(err, nil)

		expectedLocation := transformLocation(bytes, "GL")

		assert.Equal(test.location, expectedLocation, fmt.Sprintf("%s: Expected location incorrect", test.name))
	}

}

func ToByte(termStruct interface{}) ([]byte, error) {
	content, err := xml.Marshal(termStruct)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}
