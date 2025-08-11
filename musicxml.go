package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

type MusicXml struct {
     Filename string
}


// A simplified set of structs to parse the key elements of a MusicXML file
// necessary for generating a summary.
// We are only parsing the elements we need, not the entire specification.

// Score is the root element for a partwise MusicXML score.
type Score struct {
	XMLName  xml.Name `xml:"score-partwise"`
	Title string `xml:"movement-title"`
	Work     Work     `xml:"work"`
	PartList PartList `xml:"part-list"`
	Parts    []Part   `xml:"part"`
}

// Work contains the title of the musical work.
type Work struct {
	XMLName xml.Name `xml:"work"`
	Title   string   `xml:"work-title"`
}

// PartList contains the list of instruments or voices in the score.
type PartList struct {
	XMLName xml.Name    `xml:"part-list"`
	Parts   []ScorePart `xml:"score-part"`
}

// ScorePart represents a single instrument part in the score.
type ScorePart struct {
	XMLName  xml.Name `xml:"score-part"`
	ID       string   `xml:"id,attr"`
	PartName string   `xml:"part-name"`
}

// Part contains the musical data for a single instrument part.
type Part struct {
	XMLName xml.Name  `xml:"part"`
	ID      string    `xml:"id,attr"`
	Measures []Measure `xml:"measure"`
}

// Measure represents a single measure of music.
type Measure struct {
	XMLName xml.Name `xml:"measure"`
	Number  string   `xml:"number,attr"`
	Notes   []Note   `xml:"note"`
	Harmonies   []Harmony   `xml:"harmony"`
	Attributes struct {
	    Key struct {
	        Fifths int `xml:"fifths"`
		Mode string `xml:"mode"`
            } `xml:"key"`
	} `xml:"attributes"`
}

// Note represents a single musical note.
type Note struct {
	XMLName xml.Name `xml:"note"`
}

// Note represents a single musical note.
type Harmony struct {
	XMLName xml.Name `xml:"harmony"`
	Root struct {
	     RootStep string  `xml:"root-step"`
	     RootAlter string  `xml:"root-alter"`
	} `xml:"root"`
	Kind string  `xml:"kind"`
}

func (parser MusicXml) ParseScore() (Score, error) {
	// Create a Score struct to hold the parsed XML data.
	var score Score
	// Read the entire file content into a byte slice.
	xmlFile, err := ioutil.ReadFile(parser.Filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return score, err
	}

	// Unmarshal the XML data from the byte slice into the score struct.
	// This maps the XML tags to the fields in our Go structs.
	err = xml.Unmarshal(xmlFile, &score)
	if err != nil {
		fmt.Printf("Error unmarshalling XML: %v\n", err)
		return score, err
	}
	return score, nil
}