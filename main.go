package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

// A simplified set of structs to parse the key elements of a MusicXML file
// necessary for generating a summary.
// We are only parsing the elements we need, not the entire specification.

// ScorePartwise is the root element for a partwise MusicXML score.
type ScorePartwise struct {
	XMLName  xml.Name `xml:"score-partwise"`
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
}

// Note represents a single musical note.
type Note struct {
	XMLName xml.Name `xml:"note"`
}

// main function to execute the program.
func main() {
	// Check if a file path was provided as a command-line argument.
	if len(os.Args) < 2 {
		fmt.Println("Please provide a MusicXML file path as an argument.")
		fmt.Println("Example: go run main.go path/to/your/file.musicxml")
		os.Exit(1)
	}

	// Read the file path from the command-line arguments.
	filePath := os.Args[1]

	// Read the entire file content into a byte slice.
	xmlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Create a ScorePartwise struct to hold the parsed XML data.
	var score ScorePartwise

	// Unmarshal the XML data from the byte slice into the score struct.
	// This maps the XML tags to the fields in our Go structs.
	err = xml.Unmarshal(xmlFile, &score)
	if err != nil {
		fmt.Printf("Error unmarshalling XML: %v\n", err)
		os.Exit(1)
	}

	// --- Print the summary ---
	fmt.Println("---------------------------------")
	fmt.Println("MusicXML File Summary")
	fmt.Println("---------------------------------")

	// Print the title of the work.
	if score.Work.Title != "" {
		fmt.Printf("Title: %s\n", score.Work.Title)
	} else {
		fmt.Println("Title: (Not specified)")
	}

	// Count and print the number of parts.
	numParts := len(score.PartList.Parts)
	fmt.Printf("Number of Parts: %d\n", numParts)

	// List the name and ID of each part.
	fmt.Println("Part Names:")
	for _, part := range score.PartList.Parts {
		fmt.Printf("  - %s (ID: %s)\n", part.PartName, part.ID)
	}

	// Count the total number of measures and notes.
	totalMeasures := 0
	totalNotes := 0
	for _, part := range score.Parts {
		totalMeasures += len(part.Measures)
		for _, measure := range part.Measures {
			totalNotes += len(measure.Notes)
		}
	}
	fmt.Printf("Total Measures (across all parts): %d\n", totalMeasures)
	fmt.Printf("Total Notes (across all parts): %d\n", totalNotes)
	fmt.Println("---------------------------------")
}
