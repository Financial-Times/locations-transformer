package main

import (
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
		expectedLocation := transformLocation(test.term, "GL")

		assert.Equal(test.location, expectedLocation, fmt.Sprintf("%s: Expected location incorrect", test.name))
	}

}
