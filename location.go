package main

type location struct {
	UUID                   string                 `json:"uuid"`
	AlternativeIdentifiers alternativeIdentifiers `json:"alternativeIdentifiers,omitempty"`
	PrefLabel              string                 `json:"prefLabel"`
	PrimaryType            string                 `json:"type"`
	TypeHierarchy          []string               `json:"types"`
}

type alternativeIdentifiers struct {
	TME   []string `json:"TME,omitempty"`
	Uuids []string `json:"uuids,omitempty"`
}

type locationLink struct {
	APIURL string `json:"apiUrl"`
}

var locationTypes = []string{"Thing", "Concept", "Location"}
var primaryType = "Location"
