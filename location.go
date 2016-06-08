package main

//type location struct {
//	UUID          string `json:"uuid"`
//	CanonicalName string `json:"canonicalName"`
//	TmeIdentifier string `json:"tmeIdentifier"`
//	Type          string `json:"type"`
//}

type location struct {
	UUID                   string                 `json:"uuid"`
	AlternativeIdentifiers alternativeIdentifiers `json:"alternativeIdentifiers,omitempty"`
	PrefLabel              string                 `json:"prefLabel"`
	Type                   string                 `json:"type"`
}

type alternativeIdentifiers struct {
	TME               []string `json:"TME,omitempty"`
	FactsetIdentifier string   `json:"factsetIdentifier,omitempty"`
	LeiCode           string   `json:"leiCode,omitempty"`
	Uuids             []string `json:"uuids,omitempty"`
}

type locationVariation struct {
	Name      string   `json:"name"`
	Weight    string   `json:"weight"`
	Case      string   `json:"case"`
	Accent    string   `json:"accent"`
	Languages []string `json:"languages"`
}

type locationLink struct {
	APIURL string `json:"apiUrl"`
}
