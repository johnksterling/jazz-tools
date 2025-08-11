package main

import (
	"fmt"
)

// The order of keys with sharps follows the Circle of Fifths, starting from C.
var sharpKeys = []string{"C", "G", "D", "A", "E", "B", "F♯", "C♯"}

// The order of keys with flats follows the reverse Circle of Fifths, starting from C.
var flatKeys = []string{"C", "F", "B♭", "E♭", "A♭", "D♭", "G♭", "C♭"}

// FindKey takes a single integer representing the number of accidentals.
// Positive numbers are sharps, negative numbers are flats, and 0 is the key of C.
// It returns the corresponding key name and an error for invalid inputs.
func FindKey(accidentals int) (string, error) {
	// The number of accidentals must be within the range of -7 to 7.
	if accidentals > 7 || accidentals < -7 {
		return "", fmt.Errorf("number of accidentals must be between -7 and 7")
	}

	if accidentals > 0 {
		// Use the sharpKeys slice for positive inputs (sharps).
		// The number of sharps corresponds to the index.
		return sharpKeys[accidentals], nil
	}

	if accidentals < 0 {
		// Use the flatKeys slice for negative inputs (flats).
		// The absolute value of the number corresponds to the index.
		return flatKeys[-accidentals], nil
	}

	// If the input is 0, it's the key of C major.
	return "C", nil
}