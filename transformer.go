package main

import (
	"encoding/base64"
	"encoding/xml"
	"github.com/pborman/uuid"
)

func transformLocation(tmeTerm term, taxonomyName string) location {
	tmeIdentifier := buildTmeIdentifier(tmeTerm.RawID, taxonomyName)
	uuid := uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String()

	return location{
		UUID:                   uuid,
		PrefLabel:              tmeTerm.CanonicalName,
		AlternativeIdentifiers: alternativeIdentifiers{TME: []string{tmeIdentifier}, Uuids: []string{uuid}},
		Type: "Location",
	}
}

func buildTmeIdentifier(rawID string, tmeTermTaxonomyName string) string {
	id := base64.StdEncoding.EncodeToString([]byte(rawID))
	taxonomyName := base64.StdEncoding.EncodeToString([]byte(tmeTermTaxonomyName))
	return id + "-" + taxonomyName
}

type locationTransformer struct {
}

func (*locationTransformer) UnMarshallTaxonomy(contents []byte) ([]interface{}, error) {
	taxonomy := taxonomy{}
	err := xml.Unmarshal(contents, &taxonomy)
	if err != nil {
		return nil, err
	}
	interfaces := make([]interface{}, len(taxonomy.Terms))
	for i, d := range taxonomy.Terms {
		interfaces[i] = d
	}
	return interfaces, nil
}

func (*locationTransformer) UnMarshallTerm(content []byte) (interface{}, error) {
	dummyTerm := term{}
	err := xml.Unmarshal(content, &dummyTerm)
	if err != nil {
		return term{}, err
	}
	return dummyTerm, nil
}
