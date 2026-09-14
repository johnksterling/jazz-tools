package main

import (
	"strings"
	"testing"
)

func TestTonalCenterPalette(t *testing.T) {
	// Exact keys
	f := GetTonalCenterInfo("F")
	if f.Name != "F" || f.ColorHex != "#1D4ED8" {
		t.Errorf("expected F info, got %+v", f)
	}

	bb := GetTonalCenterInfo("Bb")
	if bb.Name != "Bb" || bb.ColorHex != "#D97706" {
		t.Errorf("expected Bb info, got %+v", bb)
	}

	// Enharmonic substitution
	cs := GetTonalCenterInfo("C#")
	if cs.Name != "Db" {
		t.Errorf("expected C# to map to Db, got %s", cs.Name)
	}

	fs := GetTonalCenterInfo("F#")
	if fs.Name != "Gb" {
		t.Errorf("expected F# to map to Gb, got %s", fs.Name)
	}
}

func TestTonalCenterDetectionWaltzForDebby(t *testing.T) {
	// Waltz For Debby iReal URL in F major (3/4 waltz, 40 measures)
	irealURL := "irealb://Waltz%20For%20Debby=Bill%20Evans=Medium%20Waltz=F==*A{T34F^7 |D-7 |G-7 |C7 |F^7 |D-7 |G-7 |C7 }*B[F^7 |F7 |Bb^7 |Bb-7 |A-7 |D7 |G-7 |C7 ]*C{F^7 |D-7 |G-7 |C7 |F^7 |F7 |Bb^7 |Bb-7 }*D[A-7 |D7 |G-7 |C7 |Eb^7 |Ab^7 |Db^7 |C7 ]*E{F^7 |D-7 |G-7 |C7 |F^7 |G-7 C7|F^7 |C7 Z"

	tunes, err := ParseIRealURL(irealURL)
	if err != nil {
		t.Fatalf("ParseIRealURL failed: %v", err)
	}
	tune := tunes[0]

	AnalyzeTonalCenters(tune)

	// Measure 1 should be F
	if m1Chord := tune.Measures[0].Chords[0]; m1Chord.TonalCenter != "F" {
		t.Errorf("expected Measure 1 (F^7) to be F, got %s", m1Chord.TonalCenter)
	}

	// Measure 10 (F7) should transition towards Bb
	if m10Chord := tune.Measures[9].Chords[0]; m10Chord.TonalCenter != "Bb" {
		t.Errorf("expected Measure 10 (F7) to be Bb, got %s", m10Chord.TonalCenter)
	}
	if !tune.Measures[9].Chords[0].IsTonalCenterChange {
		t.Errorf("expected Measure 10 to be marked as tonal center change")
	}

	// Measure 11 (Bb^7) should be Bb
	if m11Chord := tune.Measures[10].Chords[0]; m11Chord.TonalCenter != "Bb" {
		t.Errorf("expected Measure 11 (Bb^7) to be Bb, got %s", m11Chord.TonalCenter)
	}

	// Measure 13-14 (A-7 D7) should be G min
	if m14Chord := tune.Measures[13].Chords[0]; m14Chord.TonalCenter != "G min" {
		t.Errorf("expected Measure 14 (D7) to be G min, got %s", m14Chord.TonalCenter)
	}

	// Measures 29-31: cycle of 4ths modulation (Eb^7 -> Ab^7 -> Db^7)
	m29 := tune.Measures[28].Chords[0]
	if m29.TonalCenter != "Eb" {
		t.Errorf("expected Measure 29 (Eb^7) to be Eb, got %s", m29.TonalCenter)
	}
	m30 := tune.Measures[29].Chords[0]
	if m30.TonalCenter != "Ab" {
		t.Errorf("expected Measure 30 (Ab^7) to be Ab, got %s", m30.TonalCenter)
	}
	m31 := tune.Measures[30].Chords[0]
	if m31.TonalCenter != "Db" {
		t.Errorf("expected Measure 31 (Db^7) to be Db, got %s", m31.TonalCenter)
	}

	// Measure 32 (C7) returns to F
	m32 := tune.Measures[31].Chords[0]
	if m32.TonalCenter != "F" {
		t.Errorf("expected Measure 32 (C7) to return to F, got %s", m32.TonalCenter)
	}

	journey := HarmonicJourney(tune)
	t.Logf("Waltz For Debby Harmonic Journey:\n  %s", journey)

	if !strings.Contains(journey, "Bb") || !strings.Contains(journey, "Eb") || !strings.Contains(journey, "Ab") {
		t.Errorf("expected journey to mention modulations, got: %s", journey)
	}
}

func TestTonalCenterDetectionJordu(t *testing.T) {
	tune, err := ParseMusicXMLFile("testdata/Jordu.musicxml")
	if err != nil {
		t.Fatalf("failed to parse testdata/Jordu.musicxml: %v", err)
	}

	AnalyzeTonalCenters(tune)

	// Jordu's home key is Eb
	var firstChord *TimedChord
	for _, m := range tune.Measures {
		if len(m.Chords) > 0 {
			firstChord = &m.Chords[0]
			break
		}
	}
	if firstChord == nil {
		t.Fatalf("expected chords in Jordu, got none")
	}

	journey := HarmonicJourney(tune)
	t.Logf("Jordu Harmonic Journey:\n  %s", journey)
}
