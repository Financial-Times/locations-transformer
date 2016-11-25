package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransform(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name     string
		term     term
		location location
	}{
		{"Transform term to location", term{
			CanonicalName: "Location1",
			RawID:         "UjB4Zk1UWTBPRE0xLVIyVnVjbVZ6-R0w="},
			location{
				UUID:      "6334792f-baf0-3764-8936-fc4f240ca53c",
				PrefLabel: "Location1",
				AlternativeIdentifiers: alternativeIdentifiers{
					TME:   []string{"VWpCNFprMVVXVEJQUkUweExWSXlWblZqYlZaNi1SMHc9-R0w="},
					Uuids: []string{"6334792f-baf0-3764-8936-fc4f240ca53c"},
				},
				Type: "Location"}},
	}

	for _, test := range tests {
		expectedLocation := transformLocation(test.term, "GL")
		assert.Equal(test.location, expectedLocation, fmt.Sprintf("%s: Expected location incorrect", test.name))
	}

}
