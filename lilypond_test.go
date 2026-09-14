package main

import (
	"os"
	"testing"
)

func TestLilyPondScoreGenerationAndCompile(t *testing.T) {
	tune, err := ParseMusicXMLFile("testdata/Jordu.musicxml")
	if err != nil {
		t.Fatalf("Failed to parse testdata/Jordu.musicxml: %v", err)
	}

	lyStr, err := GenerateLilyPondScore(tune)
	if err != nil {
		t.Fatalf("GenerateLilyPondScore failed: %v", err)
	}

	lyFile := "test_jordu.ly"
	if err := os.WriteFile(lyFile, []byte(lyStr), 0644); err != nil {
		t.Fatalf("Failed to write ly file: %v", err)
	}
	defer os.Remove(lyFile)
	defer os.Remove("test_jordu.pdf")

	if err := CompileLilyPondToPDF(lyFile, "."); err != nil {
		t.Fatalf("CompileLilyPondToPDF failed: %v", err)
	}

	info, err := os.Stat("test_jordu.pdf")
	if err != nil {
		t.Fatalf("PDF file was not created: %v", err)
	}
	t.Logf("Successfully compiled print-ready PDF (%d bytes)", info.Size())
}

func TestDetermineBarsPerLine(t *testing.T) {
	tests := []struct {
		name         string
		measures     int
		userBars     int
		expectedBars int
	}{
		{"empty tune", 0, 0, 8},
		{"12-bar blues auto", 12, 0, 4},
		{"16-bar tune auto", 16, 0, 4},
		{"20-bar tune auto", 20, 0, 4},
		{"24-bar tune auto (Autumn Leaves)", 24, 0, 4},
		{"26-bar tune auto", 26, 0, 8},
		{"32-bar tune auto", 32, 0, 8},
		{"54-bar tune auto (MFT)", 54, 0, 8},
		{"64-bar tune auto", 64, 0, 10},
		{"72-bar tune auto", 72, 0, 12},
		{"user override 8 on 16-bar", 16, 8, 8},
		{"user override 4 on 32-bar", 32, 4, 4},
	}

	for _, tc := range tests {
		tune := &Tune{}
		for i := 0; i < tc.measures; i++ {
			tune.Measures = append(tune.Measures, TimedMeasure{Number: i + 1})
		}
		got := DetermineBarsPerLine(tune, tc.userBars)
		if got != tc.expectedBars {
			t.Errorf("[%s] DetermineBarsPerLine = %d, expected %d", tc.name, got, tc.expectedBars)
		}
	}
}
