package jazz

import "testing"

func TestPitchParsingAndIntervals(t *testing.T) {
	tests := []struct {
		rootStr       string
		degDelta      int
		semitoneDelta int
		expectedName  string
	}{
		// Eb major 3rd and 7th
		{"Eb", 2, 4, "G"},
		{"Eb", 6, 11, "D"},
		{"Eb", 6, 10, "Db"},

		// F# major 3rd and minor 7th
		{"F#", 2, 4, "A#"},
		{"F#", 6, 10, "E"},

		// C minor 3rd and minor 7th
		{"C", 2, 3, "Eb"},
		{"C", 6, 10, "Bb"},

		// B diminished 7th
		{"B", 6, 9, "Ab"},

		// G# major 3rd
		{"G#", 2, 4, "B#"},

		// Bb dominant 7th
		{"Bb", 2, 4, "D"},
		{"Bb", 6, 10, "Ab"},

		// Dm7b5 guide tones
		{"D", 2, 3, "F"},
		{"D", 6, 10, "C"},
	}

	for _, tt := range tests {
		p, err := ParsePitch(tt.rootStr)
		if err != nil {
			t.Fatalf("Failed to parse pitch %s: %v", tt.rootStr, err)
		}
		target := p.AddInterval(tt.degDelta, tt.semitoneDelta)
		if target.Name() != tt.expectedName {
			t.Errorf("Pitch(%s).AddInterval(%d, %d) = %s; want %s",
				tt.rootStr, tt.degDelta, tt.semitoneDelta, target.Name(), tt.expectedName)
		}
	}
}

func TestKeyToFifths(t *testing.T) {
	tests := []struct {
		keyStr         string
		expectedFifths int
		expectedMode   string
	}{
		{"C", 0, "major"},
		{"Eb", -3, "major"},
		{"F#", 6, "major"},
		{"D-", -1, "minor"},
		{"Am", 0, "minor"},
		{"C-", -3, "minor"},
		{"G minor", -2, "minor"},
		{"Bb major", -2, "major"},
	}

	for _, tt := range tests {
		f, m := KeyToFifths(tt.keyStr)
		if f != tt.expectedFifths || m != tt.expectedMode {
			t.Errorf("KeyToFifths(%q) = (%d, %s); want (%d, %s)",
				tt.keyStr, f, m, tt.expectedFifths, tt.expectedMode)
		}
	}
}
