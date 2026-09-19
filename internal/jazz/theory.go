package jazz

import (
	"fmt"
)

// Diatonic 7th chord qualities for a major scale.
var chordQualities = []string{
	"Major 7th", "minor 7th", "minor 7th", "Major 7th",
	"Dominant 7th", "minor 7th", "Half-diminished",
}

// generateDiatonicChords takes a root note and returns an array of the
// diatonic chords for its major scale with exact enharmonic spelling.
func generateDiatonicChords(rootNote string) ([]string, error) {
	sg := NewScaleGenerator()
	majorScale, err := sg.Generate(rootNote, "major")
	if err != nil {
		return nil, err
	}

	diatonicChords := make([]string, 7)
	for i := 0; i < 7; i++ {
		root := majorScale[i]
		third := majorScale[(i+2)%7]
		fifth := majorScale[(i+4)%7]
		seventh := majorScale[(i+6)%7]

		chord := fmt.Sprintf("%s %s (%s, %s, %s, %s)", root, chordQualities[i], root, third, fifth, seventh)
		diatonicChords[i] = chord
	}

	return diatonicChords, nil
}
