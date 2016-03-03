package main

import (
	"github.com/pborman/uuid"
	"encoding/base64"
)

func transformLocation(t term) location {
	tmeIdentifier := buildTmeIdentifier(t.RawID)

	locationVariants := make([]locationVariation, len(t.Variations))
	for i, variant := range t.Variations {
		locationVariants[i] = locationVariation{Name: variant.Name, Weight:variant.Weight, Case:variant.Case, Accent:variant.Accent, Languages:variant.Languages}
	}

	return location{
		UUID:          uuid.NewMD5(uuid.UUID{}, []byte(tmeIdentifier)).String(),
		CanonicalName: t.CanonicalName,
		TmeIdentifier: tmeIdentifier,
		Type:          "Location",
		Variations: locationVariants,
	}
}

func buildTmeIdentifier(rawId string) string {
	id := base64.StdEncoding.EncodeToString([]byte(rawId))
	taxonomyName := base64.StdEncoding.EncodeToString([]byte(TaxonomyName))
	return id + "-" + taxonomyName
}
