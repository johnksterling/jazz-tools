package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func printUsage() {
	fmt.Println("jazz-tools - Command line utility to analyze jazz tunes and generate companion sheets")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  jazz-tools companion <file.musicxml | irealb://... | file.html> [-o out.musicxml] [--pdf]")
	fmt.Println("  jazz-tools analyze <file.musicxml | irealb://... | file.html>")
	fmt.Println("  jazz-tools <file.musicxml | irealb://...>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  companion    Generate a companion score (MusicXML & optional PDF) with distinct")
	fmt.Println("               voice-led 3rds and 7ths on the staff for each chord's duration.")
	fmt.Println("  analyze      Display tune structure, timing, chord progression, and guide tones.")
	fmt.Println()
	fmt.Println("Flags for companion:")
	fmt.Println("  -o string    Output MusicXML path (default: <title>_guide_tones.musicxml)")
	fmt.Println("  --pdf        Compile print-ready PDF via system LilyPond")
	fmt.Println()
}

// loadTune loads a tune from either a MusicXML file or an iReal Pro URL / file.
func loadTune(target string) (*Tune, error) {
	if strings.HasPrefix(target, "irealb://") || strings.HasPrefix(target, "irealbook://") {
		tunes, err := ParseIRealURL(target)
		if err != nil {
			return nil, err
		}
		return tunes[0], nil
	}

	// Check if file exists
	if _, err := os.Stat(target); err != nil {
		return nil, fmt.Errorf("file not found: %s", target)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "irealb://") || strings.Contains(contentStr, "irealbook://") {
		tunes, err := ParseIRealURL(contentStr)
		if err == nil && len(tunes) > 0 {
			return tunes[0], nil
		}
	}

	ext := strings.ToLower(filepath.Ext(target))
	if ext == ".html" || ext == ".htm" {
		tunes, err := ParseIRealURL(contentStr)
		if err != nil {
			return nil, err
		}
		return tunes[0], nil
	}

	// Default to MusicXML parser
	return ParseMusicXML(strings.NewReader(contentStr))
}

func runAnalyze(tune *Tune) {
	fmt.Println("==================================================")
	fmt.Println("               JAZZ TUNE ANALYSIS                 ")
	fmt.Println("==================================================")
	if tune.Title != "" {
		fmt.Printf("Title:       %s\n", tune.Title)
	}
	if tune.Composer != "" {
		fmt.Printf("Composer:    %s\n", tune.Composer)
	}
	fmt.Printf("Key:         %s\n", tune.KeyName())
	fmt.Printf("Time Sig:    %d/%d\n", tune.TimeSignature[0], tune.TimeSignature[1])
	fmt.Printf("Measures:    %d\n", len(tune.Measures))

	allChords := tune.AllChords()
	fmt.Printf("Chords:      %d\n", len(allChords))
	fmt.Println("--------------------------------------------------")

	pairs, err := VoiceLeadChords(allChords)
	if err != nil {
		fmt.Printf("Error calculating guide tones: %v\n", err)
		return
	}

	fmt.Println("Measure | Beat | Chord         | Hold (beats) | Voice 1 (3/7) | Voice 2 (3/7)")
	fmt.Println("--------+------+---------------+--------------+---------------+--------------")

	pairIdx := 0
	for _, m := range tune.Measures {
		for _, tc := range m.Chords {
			var p GuideTonePair
			if pairIdx < len(pairs) {
				p = pairs[pairIdx]
				pairIdx++
			}
			fmt.Printf(" %5d  | %4.1f | %-13s | %12.1f | %-13s | %-13s\n",
				m.Number,
				tc.BeatOffset+1.0, // 1-indexed beat for musician readability
				tc.Chord.String(),
				tc.DurationBeats,
				p.Voice1.String(),
				p.Voice2.String(),
			)
		}
	}
	fmt.Println("==================================================")
}

func runCompanion(tune *Tune, outputPath string, generatePDF bool) {
	baseName := strings.ReplaceAll(tune.Title, " ", "_")
	if baseName == "" {
		baseName = "tune"
	}

	if outputPath == "" {
		outputPath = baseName + "_guide_tones.musicxml"
	}

	err := WriteMusicXMLGuideTones(tune, outputPath)
	if err != nil {
		fmt.Printf("Error writing companion MusicXML: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created MusicXML companion sheet: %s\n", outputPath)

	if generatePDF {
		lyStr, err := GenerateLilyPondScore(tune)
		if err != nil {
			fmt.Printf("Error generating LilyPond score: %v\n", err)
			return
		}

		lyFile := baseName + "_guide_tones.ly"
		if err := os.WriteFile(lyFile, []byte(lyStr), 0644); err != nil {
			fmt.Printf("Error saving LilyPond file: %v\n", err)
			return
		}

		pdfFile := baseName + "_guide_tones.pdf"
		cmdErr := CompileLilyPondToPDF(lyFile, ".")
		if cmdErr != nil {
			fmt.Printf("Error running LilyPond: %v\n", cmdErr)
		} else {
			fmt.Printf("✓ Created print-ready PDF companion sheet: %s\n", pdfFile)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "help", "-h", "--help":
		printUsage()
		return

	case "companion":
		compCmd := flag.NewFlagSet("companion", flag.ExitOnError)
		outFile := compCmd.String("o", "", "Output MusicXML file path")
		pdfFlag := compCmd.Bool("pdf", false, "Compile PDF using LilyPond")

		if len(os.Args) < 3 {
			fmt.Println("Error: please provide a MusicXML file or iReal Pro URL.")
			fmt.Println("Usage: jazz-tools companion <file.musicxml | irealb://...> [-o out.musicxml] [--pdf]")
			os.Exit(1)
		}

		target := os.Args[2]
		compCmd.Parse(os.Args[3:])

		tune, err := loadTune(target)
		if err != nil {
			fmt.Printf("Error loading tune: %v\n", err)
			os.Exit(1)
		}

		runCompanion(tune, *outFile, *pdfFlag)

	case "analyze":
		if len(os.Args) < 3 {
			fmt.Println("Error: please provide a MusicXML file or iReal Pro URL.")
			fmt.Println("Usage: jazz-tools analyze <file.musicxml | irealb://...>")
			os.Exit(1)
		}
		target := os.Args[2]
		tune, err := loadTune(target)
		if err != nil {
			fmt.Printf("Error loading tune: %v\n", err)
			os.Exit(1)
		}
		runAnalyze(tune)

	default:
		// Direct file argument without explicit subcommand
		target := os.Args[1]
		tune, err := loadTune(target)
		if err != nil {
			fmt.Printf("Error loading tune %s: %v\n", target, err)
			os.Exit(1)
		}
		runAnalyze(tune)
		fmt.Println()
		fmt.Printf("Tip: Run 'jazz-tools companion %s --pdf' to generate printable guide tone companion sheets.\n", target)
	}
}
