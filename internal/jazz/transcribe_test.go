package jazz

import (
	"os"
	"testing"
)

func TestNormalizeChordSymbol(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CΔ", "Cmaj7"},
		{"Eb^7", "Ebmaj7"},
		{"Am7b5", "Am7b5"},
		{"Aø7", "Am7b5"},
		{"B°", "Bdim7"},
		{"Bo7", "Bdim7"},
		{"Dmin7", "Dm7"},
		{"Fma7", "Fmaj7"},
		{"GM7", "Gmaj7"},
	}

	for _, tt := range tests {
		got := normalizeChordSymbol(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeChordSymbol(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestConvertRawLeadSheetToTune(t *testing.T) {
	raw := &RawExtractedLeadSheet{
		Title:         "Blue Bossa",
		Composer:      "Kenny Dorham",
		Key:           "C minor",
		TimeSignature: []int{4, 4},
		Measures: []RawExtractedMeasure{
			{Number: 1, Chords: []RawExtractedChord{{Symbol: "Cm7", Beat: 1.0, DurationBeats: 4.0}}},
			{Number: 2, Chords: []RawExtractedChord{{Symbol: "Fm7", Beat: 1.0, DurationBeats: 4.0}}},
			{Number: 3, Chords: []RawExtractedChord{{Symbol: "Dm7b5", Beat: 1.0, DurationBeats: 2.0}, {Symbol: "G7", Beat: 3.0, DurationBeats: 2.0}}},
			{Number: 4, Chords: []RawExtractedChord{{Symbol: "Cm7", Beat: 1.0, DurationBeats: 4.0}}},
		},
	}

	tune, err := convertRawLeadSheetToTune(raw)
	if err != nil {
		t.Fatalf("convertRawLeadSheetToTune failed: %v", err)
	}

	if tune.Title != "Blue Bossa" {
		t.Errorf("expected Title Blue Bossa, got %s", tune.Title)
	}
	if len(tune.Measures) != 4 {
		t.Fatalf("expected 4 measures, got %d", len(tune.Measures))
	}
	if len(tune.Measures[2].Chords) != 2 {
		t.Errorf("expected measure 3 to have 2 chords, got %d", len(tune.Measures[2].Chords))
	}

	spans := GetHarmonicJourneySpans(tune)
	if len(spans) == 0 {
		t.Errorf("expected harmonic journey spans")
	}
}

func TestTranscribeChartMock(t *testing.T) {
	dummyBytes := []byte("dummy image data")
	tune, err := TranscribeChart(dummyBytes, "image/png", TranscribeOptions{APIKey: "mock"})
	if err != nil {
		t.Fatalf("TranscribeChart mock failed: %v", err)
	}

	if tune == nil || len(tune.Measures) == 0 {
		t.Fatalf("expected transcribed tune with measures")
	}

	if tune.Title == "" {
		t.Errorf("expected tune title")
	}

	spans := GetHarmonicJourneySpans(tune)
	if len(spans) == 0 {
		t.Errorf("expected journey spans")
	}
}

func TestTranscribeChartMissingKey(t *testing.T) {
	origKey := os.Getenv("GEMINI_API_KEY")
	_ = os.Unsetenv("GEMINI_API_KEY")
	defer func() {
		if origKey != "" {
			_ = os.Setenv("GEMINI_API_KEY", origKey)
		}
	}()

	dummyBytes := []byte("dummy image data")
	_, err := TranscribeChart(dummyBytes, "image/png", TranscribeOptions{APIKey: ""})
	if err == nil {
		t.Fatalf("expected error when GEMINI_API_KEY is not configured")
	}
}

func TestCleanAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"AIzaSyABC123", "AIzaSyABC123"},
		{"  AIzaSyABC123  \n", "AIzaSyABC123"},
		{`"AIzaSyABC123"`, "AIzaSyABC123"},
		{"export GEMINI_API_KEY=AIzaSyABC123", "AIzaSyABC123"},
		{"GEMINI_API_KEY=\"AIzaSyABC123\"", "AIzaSyABC123"},
	}

	for _, tt := range tests {
		got := cleanAPIKey(tt.input)
		if got != tt.expected {
			t.Errorf("cleanAPIKey(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

