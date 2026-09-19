package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"jazz-tools/internal/jazz"
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
	fmt.Println("  -o string            Output MusicXML path (default: <title>_guide_tones.musicxml)")
	fmt.Println("  --pdf                Compile print-ready PDF via system LilyPond")
	fmt.Println("  --bars int           Measures per line (default: auto - 4 for <= 16 bars, 8 for standard tunes, auto-scaled if longer)")
	fmt.Println("  --annotations string Annotations to include: all, none, keys, devices (comma-separated, default: devices)")
	fmt.Println("  --keys / --no-keys   Include / omit tonal center key badges (default: off)")
	fmt.Println("  --devices / --no-devices Include / omit Berklee harmonic device brackets (ii-V, etc., default: on)")
	fmt.Println()
}

func runAnalyze(tune *jazz.Tune) {
	fmt.Println("=============================================================================================")
	fmt.Printf("Tune: %s", tune.Title)
	if tune.Composer != "" {
		fmt.Printf(" by %s", tune.Composer)
	}
	fmt.Println()
	if tune.Key != "" {
		fmt.Printf("Key: %s | Time Signature: %d/%d | Measures: %d\n", tune.Key, tune.TimeSignature[0], tune.TimeSignature[1], len(tune.Measures))
	} else {
		fmt.Printf("Time Signature: %d/%d | Measures: %d\n", tune.TimeSignature[0], tune.TimeSignature[1], len(tune.Measures))
	}
	fmt.Println("---------------------------------------------------------------------------------------------")

	allChords := tune.AllChords()
	if len(allChords) == 0 {
		fmt.Println("No chords found in tune.")
		return
	}

	pairs, err := jazz.VoiceLeadChords(allChords)
	if err != nil {
		fmt.Printf("Error calculating guide tones: %v\n", err)
		return
	}

	// Run tonal center analysis
	jazz.AnalyzeTonalCenters(tune)

	fmt.Println("Measure | Beat | Chord         | Hold (beats) | Voice 1 (3/7) | Voice 2 (3/7) | Tonal Center")
	fmt.Println("--------+------+---------------+--------------+---------------+---------------+-------------")

	pairIdx := 0
	for _, m := range tune.Measures {
		for _, tc := range m.Chords {
			var p jazz.GuideTonePair
			if pairIdx < len(pairs) {
				p = pairs[pairIdx]
				pairIdx++
			}
			tcInfo := jazz.GetTonalCenterInfo(tc.TonalCenter)
			coloredCenter := fmt.Sprintf("%s%-11s\033[0m", tcInfo.AnsiColor, tc.TonalCenter)
			fmt.Printf(" %5d  | %4.1f | %-13s | %12.1f | %-13s | %-13s | %s\n",
				m.Number,
				tc.BeatOffset+1.0, // 1-indexed beat for musician readability
				tc.Chord.String(),
				tc.DurationBeats,
				p.Voice1.String(),
				p.Voice2.String(),
				coloredCenter,
			)
		}
	}
	fmt.Println("---------------------------------------------------------------------------------------------")
	journey := jazz.HarmonicJourney(tune)
	if journey != "" {
		fmt.Println("Harmonic Journey:")
		fmt.Printf("  %s\n", journey)
	}
	fmt.Println("=============================================================================================")
}

func runCompanion(tune *jazz.Tune, outputPath string, generatePDF bool, userBars int, cfg jazz.AnnotationConfig) {
	tune.Title = strings.TrimSpace(tune.Title)
	tune.Title = strings.TrimSuffix(tune.Title, "(Guide Tones: 3 & 7)")
	tune.Title = strings.TrimSuffix(tune.Title, "(Guide Tones: 3 &amp; 7)")
	tune.Title = strings.TrimSpace(tune.Title)

	cleanTitle := strings.ReplaceAll(tune.Title, " ", "_")
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	baseName := reg.ReplaceAllString(cleanTitle, "")
	if baseName == "" {
		baseName = "tune"
	}

	if outputPath != "" {
		fileBase := filepath.Base(outputPath)
		ext := filepath.Ext(fileBase)
		rawBase := strings.TrimSuffix(fileBase, ext)
		rawBase = strings.TrimSuffix(rawBase, "_guide_tones")
		if rawBase != "" {
			baseName = rawBase
		}
	} else {
		outputPath = baseName + "_guide_tones.musicxml"
	}

	err := jazz.WriteMusicXMLGuideTones(tune, outputPath)
	if err != nil {
		fmt.Printf("Error writing companion MusicXML: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created MusicXML companion sheet: %s\n", outputPath)

	if generatePDF {
		lyStr, err := jazz.GenerateLilyPondScoreWithAnnotationConfig(tune, userBars, cfg)
		if err != nil {
			fmt.Printf("Error generating LilyPond score: %v\n", err)
			return
		}

		outDir := filepath.Dir(outputPath)
		lyFile := filepath.Join(outDir, baseName+"_guide_tones.ly")
		if err := os.WriteFile(lyFile, []byte(lyStr), 0644); err != nil {
			fmt.Printf("Error saving LilyPond file: %v\n", err)
			return
		}

		pdfFile := filepath.Join(outDir, baseName+"_guide_tones.pdf")
		cmdErr := jazz.CompileLilyPondToPDF(lyFile, outDir)
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
		barsFlag := compCmd.Int("bars", 0, "Target measures per line (default: auto)")
		annotationsFlag := compCmd.String("annotations", "devices", "Annotations to include: all, none, keys, devices (comma-separated)")
		keysFlag := compCmd.Bool("keys", false, "Include key center annotations")
		noKeysFlag := compCmd.Bool("no-keys", false, "Omit key center annotations")
		devicesFlag := compCmd.Bool("devices", true, "Include harmonic device annotations")
		noDevicesFlag := compCmd.Bool("no-devices", false, "Omit harmonic device annotations")

		if len(os.Args) < 3 {
			fmt.Println("Error: please provide a MusicXML file or iReal Pro URL.")
			fmt.Println("Usage: jazz-tools companion <file.musicxml | irealb://...> [-o out.musicxml] [--pdf] [--bars N]")
			os.Exit(1)
		}

		target := os.Args[2]
		compCmd.Parse(os.Args[3:])

		tune, err := jazz.LoadTune(target)
		if err != nil {
			fmt.Printf("Error loading tune: %v\n", err)
			os.Exit(1)
		}

		cfg := jazz.DefaultAnnotationConfig()

		explicitFlags := make(map[string]bool)
		compCmd.Visit(func(f *flag.Flag) {
			explicitFlags[f.Name] = true
		})

		if explicitFlags["annotations"] && *annotationsFlag != "" {
			ann := strings.ToLower(strings.TrimSpace(*annotationsFlag))
			if ann == "none" {
				cfg.ShowKeys = false
				cfg.ShowDevices = false
			} else if ann == "all" {
				cfg.ShowKeys = true
				cfg.ShowDevices = true
			} else {
				cfg.ShowKeys = false
				cfg.ShowDevices = false
				parts := strings.Split(ann, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "keys" || p == "key" {
						cfg.ShowKeys = true
					} else if p == "devices" || p == "device" {
						cfg.ShowDevices = true
					}
				}
			}
		}

		if explicitFlags["keys"] {
			cfg.ShowKeys = *keysFlag
		}
		if explicitFlags["no-keys"] && *noKeysFlag {
			cfg.ShowKeys = false
		}
		if explicitFlags["devices"] {
			cfg.ShowDevices = *devicesFlag
		}
		if explicitFlags["no-devices"] && *noDevicesFlag {
			cfg.ShowDevices = false
		}

		runCompanion(tune, *outFile, *pdfFlag, *barsFlag, cfg)

	case "analyze":
		if len(os.Args) < 3 {
			fmt.Println("Error: please provide a MusicXML file or iReal Pro URL.")
			fmt.Println("Usage: jazz-tools analyze <file.musicxml | irealb://...>")
			os.Exit(1)
		}
		target := os.Args[2]
		tune, err := jazz.LoadTune(target)
		if err != nil {
			fmt.Printf("Error loading tune: %v\n", err)
			os.Exit(1)
		}
		runAnalyze(tune)

	default:
		// Direct file argument without explicit subcommand
		target := os.Args[1]
		tune, err := jazz.LoadTune(target)
		if err != nil {
			fmt.Printf("Error loading tune %s: %v\n", target, err)
			os.Exit(1)
		}
		runAnalyze(tune)
		fmt.Println()
		fmt.Printf("Tip: Run 'jazz-tools companion %s --pdf' to generate printable guide tone companion sheets.\n", target)
	}
}
