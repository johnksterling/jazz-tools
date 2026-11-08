package jazz

import (
	"strings"
	"testing"
)

func TestDetectHarmonicDevices_MajorTwoFive(t *testing.T) {
	tune := &Tune{
		Title: "Test ii-V-I",
		Key:   "Bb",
		Measures: []TimedMeasure{
			{Number: 1, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'C'}, Quality: QualityMinor7}, BeatOffset: 0, DurationBeats: 4}}},
			{Number: 2, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'F'}, Quality: QualityDominant7}, BeatOffset: 0, DurationBeats: 4}}},
			{Number: 3, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'B', Alter: -1}, Quality: QualityMajor7}, BeatOffset: 0, DurationBeats: 4}}},
		},
	}

	devices := DetectHarmonicDevices(tune)
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	d := devices[0]
	if d.Type != DeviceMajorTwoFive {
		t.Errorf("expected DeviceMajorTwoFive, got %v", d.Type)
	}
	if d.Label != "ii - V" {
		t.Errorf("expected label 'ii - V', got '%s'", d.Label)
	}
	if !d.Resolves {
		t.Errorf("expected Resolves=true")
	}
	if d.StartMeasure != 0 || d.EndMeasure != 2 {
		t.Errorf("expected span [0, 2], got [%d, %d]", d.StartMeasure, d.EndMeasure)
	}
}

func TestDetectHarmonicDevices_MinorTwoFive(t *testing.T) {
	tune := &Tune{
		Title: "Test iiø-V-i",
		Key:   "G minor",
		Measures: []TimedMeasure{
			{Number: 1, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'A'}, Quality: QualityHalfDiminished}, BeatOffset: 0, DurationBeats: 4}}},
			{Number: 2, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'D'}, Quality: QualityDominant7}, BeatOffset: 0, DurationBeats: 4}}},
			{Number: 3, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'G'}, Quality: QualityMinor6}, BeatOffset: 0, DurationBeats: 4}}},
		},
	}

	devices := DetectHarmonicDevices(tune)
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	d := devices[0]
	if d.Type != DeviceMinorTwoFive {
		t.Errorf("expected DeviceMinorTwoFive, got %v", d.Type)
	}
	if d.Label != "iiø - V" {
		t.Errorf("expected label 'iiø - V', got '%s'", d.Label)
	}
	if !d.Resolves {
		t.Errorf("expected Resolves=true")
	}
	if d.TargetKey != "G min" {
		t.Errorf("expected TargetKey 'G min', got '%s'", d.TargetKey)
	}
}

func TestDetectHarmonicDevices_SecondaryDominant(t *testing.T) {
	tune := &Tune{
		Title: "All Of Me Bars 1-4",
		Key:   "C",
		Measures: []TimedMeasure{
			{Number: 1, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'C'}, Quality: QualityMajor7}, BeatOffset: 0, DurationBeats: 4}}},
			{Number: 2, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'E'}, Quality: QualityDominant7}, BeatOffset: 0, DurationBeats: 4}}},
			{Number: 3, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'A'}, Quality: QualityDominant7}, BeatOffset: 0, DurationBeats: 4}}},
			{Number: 4, TimeBeats: 4, Chords: []TimedChord{{Chord: Chord{Root: Pitch{Step: 'D'}, Quality: QualityMinor7}, BeatOffset: 0, DurationBeats: 4}}},
		},
	}

	devices := DetectHarmonicDevices(tune)
	// Expect E7 -> A7 and A7 -> Dm7
	if len(devices) < 2 {
		t.Fatalf("expected at least 2 secondary dominant devices, got %d", len(devices))
	}
	if devices[0].Label != "V / vi" && devices[0].Label != "V / A" {
		t.Logf("device 0 label: %s", devices[0].Label)
	}
	if devices[1].Label != "V / ii" {
		t.Errorf("expected device 1 label 'V / ii', got '%s'", devices[1].Label)
	}
}

func TestDetectHarmonicDevices_Beatrice(t *testing.T) {
	tune, err := LoadTune("Beatrice")
	if err != nil {
		t.Fatalf("failed to load Beatrice: %v", err)
	}

	devices := DetectHarmonicDevices(tune)
	t.Logf("Beatrice identified %d devices:", len(devices))
	for _, d := range devices {
		t.Logf("  - %s (bars %d-%d) resolves=%v targetKey=%s", d.Label, d.StartMeasure+1, d.EndMeasure+1, d.Resolves, d.TargetKey)
	}

	foundGbNeighbor := false
	foundEbModalInterchange := false
	foundEbNeighborMinor := false
	foundIvSubdominant := false

	for _, d := range devices {
		if d.StartMeasure == 1 && d.Label == "bIIΔ (Phrygian neighbor)" {
			foundGbNeighbor = true
		}
		if d.StartMeasure == 3 && d.Label == "bVIIΔ (modal interchange)" {
			foundEbModalInterchange = true
		}
		if d.StartMeasure == 5 && d.Label == "bIIΔ (Phrygian neighbor)" {
			foundEbNeighborMinor = true
		}
		if d.StartMeasure == 7 && d.Label == "iv-7 (minor subdominant)" {
			foundIvSubdominant = true
		}
	}

	if !foundGbNeighbor {
		t.Errorf("expected bar 2 Gb^7 to be recognized as bIIΔ (Phrygian neighbor)")
	}
	if !foundEbModalInterchange {
		t.Errorf("expected bar 4 Eb^7 to be recognized as bVIIΔ (modal interchange)")
	}
	if !foundEbNeighborMinor {
		t.Errorf("expected bar 6 Eb^7 to be recognized as bIIΔ (Phrygian neighbor) to Dm7")
	}
	if !foundIvSubdominant {
		t.Errorf("expected bar 8 Bbm7 to be recognized as iv-7 (minor subdominant)")
	}
}

func TestDetectHarmonicDevices_LadyBird(t *testing.T) {
	tune, err := LoadTune("Lady Bird")
	if err != nil {
		t.Fatalf("failed to load Lady Bird: %v", err)
	}

	devices := DetectHarmonicDevices(tune)
	foundBackdoor := false
	for _, d := range devices {
		t.Logf("Lady Bird device: %s (bars %d-%d) resolves=%v", d.Label, d.StartMeasure+1, d.EndMeasure+1, d.Resolves)
		if strings.Contains(d.Label, "backdoor") {
			foundBackdoor = true
		}
	}

	if !foundBackdoor {
		t.Errorf("expected backdoor resolution device in Lady Bird (Fm7 Bb7 -> C)")
	}
}

func TestDetectHarmonicDevices_Nardis(t *testing.T) {
	tune, err := LoadTune("Nardis")
	if err != nil {
		t.Fatalf("failed to load Nardis: %v", err)
	}

	devices := DetectHarmonicDevices(tune)
	foundPhrygian := false
	for _, d := range devices {
		t.Logf("Nardis device: %s (bars %d-%d) resolves=%v", d.Label, d.StartMeasure+1, d.EndMeasure+1, d.Resolves)
		if strings.Contains(d.Label, "Phrygian") {
			foundPhrygian = true
		}
	}

	if !foundPhrygian {
		t.Errorf("expected Phrygian chord device in Nardis (Fmaj7 in E minor)")
	}
}

func TestDetectHarmonicDevices_NightInTunisia(t *testing.T) {
	tune, err := LoadTune("A Night In Tunisia")
	if err != nil {
		t.Fatalf("failed to load A Night In Tunisia: %v", err)
	}

	devices := DetectHarmonicDevices(tune)
	foundTritoneSub := false
	for _, d := range devices {
		t.Logf("Tunisia device: %s (bars %d-%d) resolves=%v", d.Label, d.StartMeasure+1, d.EndMeasure+1, d.Resolves)
		if strings.Contains(d.Label, "tritone") || strings.Contains(d.Label, "subV") {
			foundTritoneSub = true
		}
	}

	if !foundTritoneSub {
		t.Errorf("expected tritone sub device in A Night in Tunisia (Eb7 -> Dm or F7 -> Em)")
	}
}
