package main

import (
	"encoding/base64"
	"encoding/xml"
	"github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

func transformLocation(contents []byte, taxonomyName string) location {

	tmeTerm, err := readTmeTerm(contents)

	if err != nil {
		return location{}
	}

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

func readTmeTerm(contents []byte) (term, error) {

	tmeTerm := term{}
	err := xml.Unmarshal(contents, &tmeTerm)
	if err != nil {
		logrus.Errorf("Error on unmarshalling object =%v\n", err)
		return term{}, err
	}
	return tmeTerm, nil

}

func readTmeTermsTaxonomy(contents []byte) (taxonomy, error) {

	tmeTaxonomy := taxonomy{}
	err := xml.Unmarshal(contents, &tmeTaxonomy)
	if err != nil {
		logrus.Errorf("Error on unmarshalling object =%v\n", err)
		return taxonomy{}, err
	}
	return tmeTaxonomy, nil

}
