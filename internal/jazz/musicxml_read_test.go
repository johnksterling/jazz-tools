package jazz

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

	// Verify melody notes extraction
	if !tune.HasMelody() {
		t.Errorf("Expected tune.HasMelody() to be true for Jordu")
	}
	melodyCount := tune.MelodyNotesCount()
	t.Logf("Total melody notes extracted in Jordu: %d", melodyCount)
	if melodyCount == 0 {
		t.Errorf("Expected non-zero melody notes in Jordu")
	}

	// Verify Measure 1 melody notes
	m1 := tune.Measures[0]
	t.Logf("Measure 1 has %d melody note elements", len(m1.Melody))
	var m1PitchNames []string
	for _, n := range m1.Melody {
		m1PitchNames = append(m1PitchNames, n.PitchName())
	}
	t.Logf("Measure 1 Melody notes: %v", m1PitchNames)
	// Expected opening pickup line: Rest, G3, C4, D4, Eb4, F4, G4, Eb4
	if len(m1PitchNames) < 8 || m1PitchNames[0] != "Rest" || m1PitchNames[1] != "G3" || m1PitchNames[4] != "Eb4" {
		t.Errorf("Unexpected measure 1 melody notes: %v", m1PitchNames)
	}
}
