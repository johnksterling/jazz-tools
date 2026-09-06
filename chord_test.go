package main

import "testing"

func TestGuideToneExtraction(t *testing.T) {
	tests := []struct {
		symbol       string
		expected3rd  string
		expected7th  string
	}{
		// Eb major 7th -> G, D
		{"Ebmaj7", "G", "D"},
		// D7 -> F#, C
		{"D7", "F#", "C"},
		// G7 -> B, F
		{"G7", "B", "F"},
		// Cm7 -> Eb, Bb
		{"Cm7", "Eb", "Bb"},
		// Fm7 -> Ab, Eb
		{"Fm7", "Ab", "Eb"},
		// Bb7 -> D, Ab
		{"Bb7", "D", "Ab"},
		// Dm7b5 -> F, C
		{"Dm7b5", "F", "C"},
		// Bdim7 -> D, Ab
		{"Bdim7", "D", "Ab"},
		// C6 -> E, A
		{"C6", "E", "A"},
		// Cm6 -> Eb, A
		{"Cm6", "Eb", "A"},
		// F#7 -> A#, E
		{"F#7", "A#", "E"},
		// Ab7 -> C, Gb
		{"Ab7", "C", "Gb"},
	}

	for _, tt := range tests {
		c, err := ParseChord(tt.symbol)
		if err != nil {
			t.Fatalf("ParseChord(%s) failed: %v", tt.symbol, err)
		}
		third, seventh, err := c.GetGuideTones()
		if err != nil {
			t.Fatalf("GetGuideTones(%s) failed: %v", tt.symbol, err)
		}
		if third.Name() != tt.expected3rd {
			t.Errorf("Chord(%s) 3rd = %s; want %s", tt.symbol, third.Name(), tt.expected3rd)
		}
		if seventh.Name() != tt.expected7th {
			t.Errorf("Chord(%s) 7th = %s; want %s", tt.symbol, seventh.Name(), tt.expected7th)
		}
	}
}
