package main

import (
	"testing"
)

func TestParseJorduMusicXML(t *testing.T) {
	tune, err := ParseMusicXMLFile("testdata/Jordu.musicxml")
	if err != nil {
		t.Fatalf("Failed to parse testdata/Jordu.musicxml: %v", err)
	}

	t.Logf("Parsed Tune: %s by %s, Key: %s, Measures: %d",
		tune.Title, tune.Composer, tune.KeyName(), len(tune.Measures))

	if tune.Title != "Jordu" {
		t.Errorf("Expected title 'Jordu', got '%s'", tune.Title)
	}

	allChords := tune.AllChords()
	t.Logf("Total chords found: %d", len(allChords))

	if len(allChords) == 0 {
		t.Fatalf("Expected chords in Jordu, got 0")
	}

	// Verify that Root alterations are correctly captured (e.g. Eb is not E!)
	hasEb := false
	for _, c := range allChords {
		if c.Root.Name() == "Eb" {
			hasEb = true
			break
		}
	}
	if !hasEb {
		t.Errorf("Expected to find Eb chords in Jordu, but none found (alteration bug?)")
	}

	// Check Measure 2 chords (D7, G7, Cmin7)
	m2 := tune.Measures[1] // 0-indexed measure 2
	t.Logf("Measure 2 has %d chords", len(m2.Chords))
	for _, tc := range m2.Chords {
		t.Logf("  Measure 2 Chord: %s, BeatOffset: %.2f, DurationBeats: %.2f",
			tc.Chord.Symbol, tc.BeatOffset, tc.DurationBeats)
	}
}
