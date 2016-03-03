package main

type location struct {
	UUID          string `json:"uuid"`
	CanonicalName string `json:"canonicalName"`
	TmeIdentifier string `json:"tmeIdentifier"`
	Type          string `json:"type"`
	Variations    []locationVariation `json:"variations"`
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
