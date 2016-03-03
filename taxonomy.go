package main

type taxonomy struct {
	Terms []term `xml:"term"`
}
//TODO revise fields
type term struct {
	CanonicalName string        `xml:"name"`
	RawID         string        `xml:"id"`
	Enabled       string        `xml:"enabled"`
	Variations    []variation   `xml:"variations>variation"`

}

type variation struct {
	Name      string   `xml:"name"`
	Weight    string   `xml:"weight"`
	Case      string   `xml:"case"`
	Accent    string   `xml:"accent"`
	Languages []string `xml:"languages>language"`
}

