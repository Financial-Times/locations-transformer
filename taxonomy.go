package main

type taxonomy struct {
	Terms []term `xml:"category"`
}

//TODO revise fields
type term struct {
	CanonicalName string `xml:"name"`
	RawID         string `xml:"id"`
}
