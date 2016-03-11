package main

import (
	"encoding/base64"
	"github.com/pborman/uuid"
)

func transformLocation(t term) location {
	tmeIdentifier := buildTmeIdentifier(t.RawID)

	return location{
		UUID:          uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String(),
		CanonicalName: t.CanonicalName,
		TmeIdentifier: tmeIdentifier,
		Type:          "Location",
	}
}

func buildTmeIdentifier(rawId string) string {
	id := base64.StdEncoding.EncodeToString([]byte(rawId))
	taxonomyName := base64.StdEncoding.EncodeToString([]byte(TaxonomyName))
	return id + "-" + taxonomyName
}
