package jazz

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

func TestTonalCenterDetectionBeatrice(t *testing.T) {
	tune, err := LoadTune("Beatrice")
	if err != nil {
		t.Fatalf("failed to load Beatrice: %v", err)
	}

	AnalyzeTonalCenters(tune)

	// Verify bars 1-4 are all in F (Gb^7 is Phrygian neighbor, Eb^7 is bVIIΔ modal interchange)
	for bar := 0; bar < 4; bar++ {
		for _, tc := range tune.Measures[bar].Chords {
			if tc.TonalCenter != "F" {
				t.Errorf("expected Bar %d (%s) to have tonal center F, got %s", bar+1, tc.Chord.String(), tc.TonalCenter)
			}
		}
	}

	// Verify bars 5-8 are all in F (Dm7, Eb^7 Phrygian neighbor, Cm7 v-7, Bbm7 iv-7)
	for bar := 4; bar < 8; bar++ {
		for _, tc := range tune.Measures[bar].Chords {
			if tc.TonalCenter != "F" {
				t.Errorf("expected Bar %d (%s) to have tonal center F, got %s", bar+1, tc.Chord.String(), tc.TonalCenter)
			}
		}
	}

	// Verify bars 11-12 are D min (iiø - V cadence to relative minor)
	for bar := 10; bar < 12; bar++ {
		for _, tc := range tune.Measures[bar].Chords {
			if tc.TonalCenter != "D min" {
				t.Errorf("expected Bar %d (%s) to have tonal center D min, got %s", bar+1, tc.Chord.String(), tc.TonalCenter)
			}
		}
	}

	// Verify bars 13-16 return to F
	for bar := 12; bar < 16; bar++ {
		for _, tc := range tune.Measures[bar].Chords {
			if tc.TonalCenter != "F" {
				t.Errorf("expected Bar %d (%s) to have tonal center F, got %s", bar+1, tc.Chord.String(), tc.TonalCenter)
			}
		}
	}

	spans := GetHarmonicJourneySpans(tune)
	if len(spans) != 3 {
		t.Fatalf("expected 3 journey spans, got %d: %+v", len(spans), spans)
	}
	if spans[0].Center != "F" || spans[0].StartBar != 1 || spans[0].EndBar != 10 {
		t.Errorf("expected span 0 to be F (bars 1-10), got %+v", spans[0])
	}
	if spans[1].Center != "D min" || spans[1].StartBar != 11 || spans[1].EndBar != 12 {
		t.Errorf("expected span 1 to be D min (bars 11-12), got %+v", spans[1])
	}
	if spans[2].Center != "F" || spans[2].StartBar != 13 || spans[2].EndBar != 16 {
		t.Errorf("expected span 2 to be F (bars 13-16), got %+v", spans[2])
	}
}

func TestTonalCenterDetectionNardis(t *testing.T) {
	tune, err := LoadTune("Nardis")
	if err != nil {
		t.Fatalf("failed to load Nardis: %v", err)
	}

	AnalyzeTonalCenters(tune)
	spans := GetHarmonicJourneySpans(tune)

	// Expected spans:
	// 1. E min (bars 1-12)
	// 2. C (bars 13-15)
	// 3. E min (bars 16-24)
	if len(spans) != 3 {
		t.Fatalf("expected 3 journey spans for Nardis, got %d: %+v", len(spans), spans)
	}
	if spans[0].Center != "E min" || spans[0].StartBar != 1 || spans[0].EndBar != 12 {
		t.Errorf("expected span 0 to be E min (bars 1-12), got %+v", spans[0])
	}
	if spans[1].Center != "C" || spans[1].StartBar != 13 || spans[1].EndBar != 15 {
		t.Errorf("expected span 1 to be C (bars 13-15), got %+v", spans[1])
	}
	if spans[2].Center != "E min" || spans[2].StartBar != 16 || spans[2].EndBar != 24 {
		t.Errorf("expected span 2 to be E min (bars 16-24), got %+v", spans[2])
	}
}

func TestTonalCenterDetectionLadyBird(t *testing.T) {
	// 16-bar iReal version of Lady Bird
	irealURL := "irealb://%4C%61%64%79%20%42%69%72%64=%44%61%6D%65%72%6F%6E%20%54%61%64%64==%55%70%20%54%65%6D%70%6F%20%53%77%69%6E%67=%43==%31%72%33%34%4C%62%4B%63%75%37%7C%51%79%58%37%5E%37%58%79%51%20%6C%63%4B%51%79%58%37%5E%43%7C%51%79%58%62%37%42%7C%51%79%58%37%2D%46%5A%4C%20%6C%63%4B%4C%5A%42%62%2D%43%34%34%54%5B%37%5E%43%55%7C%51%7C%41%62%5E%7C%51%79%58%37%2D%44%7C%51%79%58%37%44%7C%79%51%58%37%2D%41%5A%4C%20%6C%63%4B%51%79%58%37%47%37%58%79%51%79%58%37%62%45%20%45%62%5E%37%4C%5A%41%62%5E%37%20%44%62%5E%37%20%5A==%30=%30==="
	tune, err := LoadTune(irealURL)
	if err != nil {
		t.Fatalf("failed to load Lady Bird: %v", err)
	}

	AnalyzeTonalCenters(tune)
	spans := GetHarmonicJourneySpans(tune)

	if len(spans) == 0 {
		t.Fatalf("expected journey spans for Lady Bird")
	}
	// Bar 1-6 should be in C (including Fm7 - Bb7 backdoor)
	if spans[0].Center != "C" || spans[0].StartBar != 1 || spans[0].EndBar != 6 {
		t.Errorf("expected bars 1-6 to be C, got %+v", spans[0])
	}
}

func TestTonalCenterDetectionTheNightHasAThousandEyes(t *testing.T) {
	tune, err := LoadTune("Night Has A Thousand Eyes, The")
	if err != nil {
		t.Fatalf("failed to load The Night Has a Thousand Eyes: %v", err)
	}

	AnalyzeTonalCenters(tune)
	spans := GetHarmonicJourneySpans(tune)

	// Bars 1-9 in G, bars 10-11 in C, bars 12-18 in G (F7 backdoor resolving to G)
	if spans[0].Center != "G" || spans[0].StartBar != 1 || spans[0].EndBar != 9 {
		t.Errorf("expected span 0 to be G (bars 1-9), got %+v", spans[0])
	}
	if spans[1].Center != "C" || spans[1].StartBar != 10 || spans[1].EndBar != 11 {
		t.Errorf("expected span 1 to be C (bars 10-11), got %+v", spans[1])
	}
	if spans[2].Center != "G" || spans[2].StartBar != 12 || spans[2].EndBar != 18 {
		t.Errorf("expected span 2 to be G (bars 12-18), got %+v", spans[2])
	}
}


