package main

import (
	"fmt"
)

// Chromatic scale notes. The indices are used for calculating intervals.
var notes = []string{
	"C", "C#", "D", "D#", "E", "F",
	"F#", "G", "G#", "A", "A#", "B",
}

// A map to quickly look up the index of a note.
var noteToIndex = map[string]int{
	"C":  0, "C#": 1, "D":  2, "D#": 3, "E":  4, "F":  5,
	"F#": 6, "G":  7, "G#": 8, "A":  9, "A#": 10, "B": 11,
}

// Major scale intervals in semitones: Whole, Whole, Half, Whole, Whole, Whole, Half
var majorScaleIntervals = []int{2, 2, 1, 2, 2, 2, 1}

// Diatonic 7th chord qualities for a major scale.
var chordQualities = []string{
	"Major 7th", "minor 7th", "minor 7th", "Major 7th",
	"Dominant 7th", "minor 7th", "Half-diminished",
}

// generateDiatonicChords takes a root note and returns an array of the
// diatonic chords for its major scale.
func generateDiatonicChords(rootNote string) ([]string, error) {
	// Get the starting index of the root note.
	startIndex, ok := noteToIndex[rootNote]
	if !ok {
		return nil, fmt.Errorf("invalid root note: %s", rootNote)
	}

	// Build the major scale.
	majorScale := make([]string, 7)
	currentIndex := startIndex
	for i := 0; i < 7; i++ {
		majorScale[i] = notes[currentIndex]
		currentIndex = (currentIndex + majorScaleIntervals[i]) % 12
	}

	// Build the diatonic chords using the notes of the major scale.
	diatonicChords := make([]string, 7)
	for i := 0; i < 7; i++ {
		// A 7th chord is built from the root, third, fifth, and seventh of the scale.
		root := majorScale[i]
		third := majorScale[(i+2)%7]
		fifth := majorScale[(i+4)%7]
		seventh := majorScale[(i+6)%7]

		// Construct the chord string with its quality.
		chord := fmt.Sprintf("%s %s (%s, %s, %s, %s)", root, chordQualities[i], root, third, fifth, seventh)
		diatonicChords[i] = chord
	}

	return diatonicChords, nil
}
