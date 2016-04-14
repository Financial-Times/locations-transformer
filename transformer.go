package main

import (
	"encoding/base64"
	"encoding/xml"
	"github.com/pborman/uuid"
)

func transformLocation(tmeTerm term, taxonomyName string) location {
	tmeIdentifier := buildTmeIdentifier(tmeTerm.RawID, taxonomyName)

	return location{
		UUID:          uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String(),
		CanonicalName: tmeTerm.CanonicalName,
		TmeIdentifier: tmeIdentifier,
		Type:          "Location",
	}
}

func buildTmeIdentifier(rawId string, tmeTermTaxonomyName string) string {
	id := base64.StdEncoding.EncodeToString([]byte(rawId))
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
	var interfaces []interface{} = make([]interface{}, len(taxonomy.Terms))
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