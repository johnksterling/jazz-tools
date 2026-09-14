package main

import (
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
