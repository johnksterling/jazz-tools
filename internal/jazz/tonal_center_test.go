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
	// Verify high-contrast color variety between home key (F) and modulation (D min)
	if spans[0].ColorHex != "#1D4ED8" {
		t.Errorf("expected span 0 (F) to have Royal Blue #1D4ED8, got %s", spans[0].ColorHex)
	}
	if spans[1].ColorHex != "#D97706" {
		t.Errorf("expected span 1 (D min) to have Warm Amber #D97706, got %s", spans[1].ColorHex)
	}
	if spans[2].ColorHex != "#1D4ED8" {
		t.Errorf("expected span 2 (F return) to have Royal Blue #1D4ED8, got %s", spans[2].ColorHex)
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

func TestTonalCenterDetectionRecordaMe(t *testing.T) {
	// Construct Joe Henderson's "Recorda Me" (16 bars):
	// Bars 1-4: Am7 with prominent F# in melody -> A Dorian (parent: G major / 1 sharp)
	// Bars 5-7: Cm7 with prominent A♮ in melody -> C Dorian (parent: Bb major / 2 flats)
	// Bars 8-9: Cm7 - F7 -> Bbmaj7
	// Bars 10-11: Bbm7 - Eb7 -> Abmaj7
	// Bars 12-13: Abm7 - Db7 -> Gbmaj7
	// Bars 14-15: Gm7 - C7 -> Fmaj7
	// Bar 16: E7#9 (turnaround resolving to A dorian)
	tune := &Tune{
		Title:         "Recorda Me",
		Composer:      "Joe Henderson",
		Key:           "A minor",
		TimeSignature: [2]int{4, 4},
		Measures:      make([]TimedMeasure, 16),
	}

	for i := range tune.Measures {
		tune.Measures[i].Number = i + 1
	}

	// Bars 1-4: Am7 with F# in melody
	pE4, _ := ParsePitch("E4")
	pG4, _ := ParsePitch("G4")
	pFSharp4, _ := ParsePitch("F#4")
	pD4, _ := ParsePitch("D4")
	pC4, _ := ParsePitch("C4")
	pB3, _ := ParsePitch("B3")
	pA3, _ := ParsePitch("A3")

	am7, _ := ParseChord("Am7")
	for m := 0; m < 4; m++ {
		tune.Measures[m].Chords = []TimedChord{
			{Chord: am7, MeasureNumber: m + 1, BeatOffset: 0.0, DurationBeats: 4.0},
		}
		tune.Measures[m].Melody = []MelodyNote{
			{Pitch: &pE4, DurationBeats: 1.0, BeatOffset: 0.0},
			{Pitch: &pG4, DurationBeats: 1.0, BeatOffset: 1.0},
			{Pitch: &pFSharp4, DurationBeats: 2.0, BeatOffset: 2.0}, // Prominent F# (Dorian 6th)
		}
	}

	// Bars 5-7: Cm7 with A natural in melody
	pBb4, _ := ParsePitch("Bb4")
	pA4, _ := ParsePitch("A4")
	pF4, _ := ParsePitch("F4")
	pEb4, _ := ParsePitch("Eb4")

	cm7, _ := ParseChord("Cm7")
	for m := 4; m < 7; m++ {
		tune.Measures[m].Chords = []TimedChord{
			{Chord: cm7, MeasureNumber: m + 1, BeatOffset: 0.0, DurationBeats: 4.0},
		}
		tune.Measures[m].Melody = []MelodyNote{
			{Pitch: &pG4, DurationBeats: 1.0, BeatOffset: 0.0},
			{Pitch: &pBb4, DurationBeats: 1.0, BeatOffset: 1.0},
			{Pitch: &pA4, DurationBeats: 2.0, BeatOffset: 2.0}, // Prominent A♮ (Dorian 6th)
		}
	}

	f7, _ := ParseChord("F7")
	bbMaj7, _ := ParseChord("Bbmaj7")
	bbm7, _ := ParseChord("Bbm7")
	eb7, _ := ParseChord("Eb7")
	abMaj7, _ := ParseChord("Abmaj7")
	abm7, _ := ParseChord("Abm7")
	db7, _ := ParseChord("Db7")
	gbMaj7, _ := ParseChord("Gbmaj7")
	gm7, _ := ParseChord("Gm7")
	c7, _ := ParseChord("C7")
	fMaj7, _ := ParseChord("Fmaj7")
	e7, _ := ParseChord("E7")

	// Bar 8: Cm7 (2 beats), F7 (2 beats)
	tune.Measures[7].Chords = []TimedChord{
		{Chord: cm7, MeasureNumber: 8, BeatOffset: 0.0, DurationBeats: 2.0},
		{Chord: f7, MeasureNumber: 8, BeatOffset: 2.0, DurationBeats: 2.0},
	}
	// Bar 9: Bbmaj7
	tune.Measures[8].Chords = []TimedChord{
		{Chord: bbMaj7, MeasureNumber: 9, BeatOffset: 0.0, DurationBeats: 4.0},
	}
	// Bar 10: Bbm7 (2), Eb7 (2)
	tune.Measures[9].Chords = []TimedChord{
		{Chord: bbm7, MeasureNumber: 10, BeatOffset: 0.0, DurationBeats: 2.0},
		{Chord: eb7, MeasureNumber: 10, BeatOffset: 2.0, DurationBeats: 2.0},
	}
	// Bar 11: Abmaj7
	tune.Measures[10].Chords = []TimedChord{
		{Chord: abMaj7, MeasureNumber: 11, BeatOffset: 0.0, DurationBeats: 4.0},
	}
	// Bar 12: Abm7 (2), Db7 (2)
	tune.Measures[11].Chords = []TimedChord{
		{Chord: abm7, MeasureNumber: 12, BeatOffset: 0.0, DurationBeats: 2.0},
		{Chord: db7, MeasureNumber: 12, BeatOffset: 2.0, DurationBeats: 2.0},
	}
	// Bar 13: Gbmaj7
	tune.Measures[12].Chords = []TimedChord{
		{Chord: gbMaj7, MeasureNumber: 13, BeatOffset: 0.0, DurationBeats: 4.0},
	}
	// Bar 14: Gm7 (2), C7 (2)
	tune.Measures[13].Chords = []TimedChord{
		{Chord: gm7, MeasureNumber: 14, BeatOffset: 0.0, DurationBeats: 2.0},
		{Chord: c7, MeasureNumber: 14, BeatOffset: 2.0, DurationBeats: 2.0},
	}
	// Bar 15: Fmaj7
	tune.Measures[14].Chords = []TimedChord{
		{Chord: fMaj7, MeasureNumber: 15, BeatOffset: 0.0, DurationBeats: 4.0},
	}
	// Bar 16: E7
	tune.Measures[15].Chords = []TimedChord{
		{Chord: e7, MeasureNumber: 16, BeatOffset: 0.0, DurationBeats: 4.0},
	}

	AnalyzeTonalCenters(tune)
	spans := GetHarmonicJourneySpans(tune)

	t.Logf("Recorda Me Harmonic Journey: %s", HarmonicJourney(tune))

	if len(spans) < 7 {
		t.Fatalf("expected at least 7 journey spans for Recorda Me, got %d", len(spans))
	}

	// 1. Bars 1-4: A Dorian (parent: G major notes)
	if spans[0].Center != "A dorian" || spans[0].ParentKey != "G" || spans[0].StartBar != 1 || spans[0].EndBar != 4 {
		t.Errorf("expected span 0 to be A dorian (G notes, bars 1-4), got %+v", spans[0])
	}

	// 2. Bars 5-7: C Dorian (parent: Bb major notes)
	if spans[1].Center != "C dorian" || spans[1].ParentKey != "Bb" || spans[1].StartBar != 5 || spans[1].EndBar != 7 {
		t.Errorf("expected span 1 to be C dorian (Bb notes, bars 5-7), got %+v", spans[1])
	}

	// 3. Bars 8-9: Bb major
	if spans[2].Center != "Bb" || spans[2].StartBar != 8 || spans[2].EndBar != 9 {
		t.Errorf("expected span 2 to be Bb (bars 8-9), got %+v", spans[2])
	}

	// 4. Bars 10-11: Ab major
	if spans[3].Center != "Ab" || spans[3].StartBar != 10 || spans[3].EndBar != 11 {
		t.Errorf("expected span 3 to be Ab (bars 10-11), got %+v", spans[3])
	}

	// 5. Bars 12-13: Gb major
	if spans[4].Center != "Gb" || spans[4].StartBar != 12 || spans[4].EndBar != 13 {
		t.Errorf("expected span 4 to be Gb (bars 12-13), got %+v", spans[4])
	}

	// 6. Bars 14-15: F major
	if spans[5].Center != "F" || spans[5].StartBar != 14 || spans[5].EndBar != 15 {
		t.Errorf("expected span 5 to be F (bars 14-15), got %+v", spans[5])
	}

	// 7. Bar 16: A dorian turnaround
	if spans[6].Center != "A dorian" || spans[6].StartBar != 16 || spans[6].EndBar != 16 {
		t.Errorf("expected span 6 to be A dorian (bar 16), got %+v", spans[6])
	}

	// Verify distinct high-contrast colors are assigned to each modulation
	colors := make(map[string]string)
	for i := 0; i < 6; i++ {
		center := spans[i].Center
		color := spans[i].ColorHex
		if prevCenter, exists := colors[color]; exists {
			t.Errorf("duplicate color %s assigned to %s and %s", color, prevCenter, center)
		}
		colors[color] = center
	}

	// Also verify that without F#, a minor tune with F natural stays natural minor (A min)
	tuneNatural := &Tune{
		Title:         "Natural Minor Tune",
		Key:           "A minor",
		TimeSignature: [2]int{4, 4},
		Measures:      make([]TimedMeasure, 2),
	}
	tuneNatural.Measures[0].Number = 1
	tuneNatural.Measures[1].Number = 2
	pF4Natural, _ := ParsePitch("F4")
	tuneNatural.Measures[0].Chords = []TimedChord{{Chord: am7, MeasureNumber: 1, BeatOffset: 0.0, DurationBeats: 4.0}}
	tuneNatural.Measures[0].Melody = []MelodyNote{{Pitch: &pF4Natural, DurationBeats: 2.0, BeatOffset: 0.0}}
	tuneNatural.Measures[1].Chords = []TimedChord{{Chord: am7, MeasureNumber: 2, BeatOffset: 0.0, DurationBeats: 4.0}}
	AnalyzeTonalCenters(tuneNatural)
	spansNat := GetHarmonicJourneySpans(tuneNatural)
	if spansNat[0].Center != "A min" {
		t.Errorf("expected natural minor tune with F natural to stay 'A min', got %s", spansNat[0].Center)
	}

	// Also verify that an Am6 chord detects as A dorian even without melody
	am6, _ := ParseChord("Am6")
	tuneM6 := &Tune{
		Title:         "Am6 Vamp",
		Key:           "A minor",
		TimeSignature: [2]int{4, 4},
		Measures:      make([]TimedMeasure, 2),
	}
	tuneM6.Measures[0].Number = 1
	tuneM6.Measures[1].Number = 2
	tuneM6.Measures[0].Chords = []TimedChord{{Chord: am6, MeasureNumber: 1, BeatOffset: 0.0, DurationBeats: 4.0}}
	tuneM6.Measures[1].Chords = []TimedChord{{Chord: am6, MeasureNumber: 2, BeatOffset: 0.0, DurationBeats: 4.0}}
	AnalyzeTonalCenters(tuneM6)
	spansM6 := GetHarmonicJourneySpans(tuneM6)
	if spansM6[0].Center != "A dorian" {
		t.Errorf("expected Am6 chord to detect as 'A dorian', got %s", spansM6[0].Center)
	}
	_ = pD4
	_ = pC4
	_ = pB3
	_ = pA3
	_ = pF4
	_ = pEb4
}
