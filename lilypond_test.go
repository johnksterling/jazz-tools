package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestLilyPondScoreGenerationAndCompile(t *testing.T) {
	tune, err := ParseMusicXMLFile("Jordu.musicxml")
	if err != nil {
		t.Fatalf("Failed to parse Jordu.musicxml: %v", err)
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

	cmd := exec.Command("/opt/homebrew/bin/lilypond", "--pdf", lyFile)
	out, err := cmd.CombinedOutput()
	t.Logf("LilyPond output: %s", string(out))
	if err != nil {
		t.Fatalf("LilyPond execution failed: %v", err)
	}

	info, err := os.Stat("test_jordu.pdf")
	if err != nil {
		t.Fatalf("PDF file was not created: %v", err)
	}
	t.Logf("Successfully compiled print-ready PDF (%d bytes)", info.Size())
}
