package jazz

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildDeterministicHarmonicSummary(t *testing.T) {
	tune, err := LoadTune("Beatrice")
	if err != nil {
		t.Fatalf("failed to load Beatrice: %v", err)
	}

	summary := BuildDeterministicHarmonicSummary(tune)
	if !strings.Contains(summary, "Beatrice") {
		t.Errorf("expected summary to contain tune title, got: %s", summary)
	}
	if !strings.Contains(summary, "F") {
		t.Errorf("expected summary to mention key F, got: %s", summary)
	}
}

func TestGenerateHarmonicSummary_MockAndFallback(t *testing.T) {
	tune, err := LoadTune("Beatrice")
	if err != nil {
		t.Fatalf("failed to load Beatrice: %v", err)
	}

	mockSummary, isLLM, err := GenerateHarmonicSummary(tune, TranscribeOptions{APIKey: "mock"})
	if err != nil {
		t.Fatalf("unexpected error in mock summary: %v", err)
	}
	if !isLLM {
		t.Errorf("expected isLLM=true in mock mode")
	}
	if !strings.Contains(mockSummary, "Beatrice") {
		t.Errorf("expected mock summary to contain tune title, got: %s", mockSummary)
	}
}

func TestGenerateLilyPondFullCompanion(t *testing.T) {
	tune, err := LoadTune("Beatrice")
	if err != nil {
		t.Fatalf("failed to load Beatrice: %v", err)
	}

	lyStr, err := GenerateLilyPondFullCompanion(tune, CompanionPDFOptions{
		AnnotationConfig:      DefaultAnnotationConfig(),
		IncludeAnalysisReport: true,
		HarmonicSummary:       "Beatrice centers on F major with modal interchange and a brief D minor cadence.",
		IsLLMSummary:          true,
	})
	if err != nil {
		t.Fatalf("GenerateLilyPondFullCompanion failed: %v", err)
	}

	for _, want := range []string{
		"\\bookpart",
		"\\markuplist",
		"Harmonic Overview",
		"Harmonic Structure Summary (LLM Analysis)",
		"Harmonic Modulation Trajectory",
		"Identified Berklee Harmonic Devices",
		"Measure-by-Measure Changes & Guide Tones",
	} {
		if !strings.Contains(lyStr, want) {
			t.Errorf("expected LilyPond output to contain %q", want)
		}
	}

	// If lilypond is installed, verify the generated .ly compiles cleanly to PDF
	if _, err := exec.LookPath("lilypond"); err == nil {
		tmpDir := t.TempDir()
		lyPath := filepath.Join(tmpDir, "full.ly")
		if err := os.WriteFile(lyPath, []byte(lyStr), 0644); err != nil {
			t.Fatalf("failed to write .ly: %v", err)
		}
		if err := CompileLilyPondToPDF(lyPath, tmpDir); err != nil {
			t.Fatalf("LilyPond compilation of full companion failed: %v", err)
		}
		pdfInfo, err := os.Stat(filepath.Join(tmpDir, "full.pdf"))
		if err != nil || pdfInfo.Size() == 0 {
			t.Fatalf("expected non-empty full.pdf, err=%v", err)
		}
	}
}
