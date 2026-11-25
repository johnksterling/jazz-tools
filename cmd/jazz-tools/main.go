package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"jazz-tools/internal/jazz"
	"jazz-tools/internal/server"
)

func printUsage() {
	fmt.Println("jazz-tools - Command line utility to analyze jazz tunes and generate companion sheets")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  jazz-tools companion <file.musicxml | irealb://... | file.html> [-o out.musicxml] [--pdf]")
	fmt.Println("  jazz-tools analyze <file.musicxml | irealb://... | file.html>")
	fmt.Println("  jazz-tools transcribe <chart.pdf | chart.png | chart.jpg> [-o out.musicxml] [--pdf]")
	fmt.Println("  jazz-tools batch <setlist.msf> [-o outdir] [--zip bundle.zip] [--full] [--transcribe]")
	fmt.Println("  jazz-tools serve [--port 8080] [--host 0.0.0.0]")
	fmt.Println("  jazz-tools <file.musicxml | irealb://... | chart.pdf | chart.png>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  companion    Generate a companion score (MusicXML & optional PDF) with distinct")
	fmt.Println("               voice-led 3rds and 7ths on the staff for each chord's duration.")
	fmt.Println("  analyze      Display tune structure, timing, chord progression, and guide tones.")
	fmt.Println("  transcribe   Transcribe a scanned PDF or photo of a jazz lead sheet using Vision LLM.")
	fmt.Println("  batch        Batch-analyze a MobileSheets .msf setlist into companion PDFs + CSV import.")
	fmt.Println("  serve        Launch the web interface and REST API server.")
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

func runCompanion(tune *jazz.Tune, outputPath string, generatePDF bool, userBars int, cfg jazz.AnnotationConfig, fullReport bool, opts jazz.TranscribeOptions) {
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
		pdfOpts := jazz.CompanionPDFOptions{
			BarsPerLine:           userBars,
			AnnotationConfig:      cfg,
			IncludeAnalysisReport: fullReport,
		}
		if fullReport {
			summary, isLLM, _ := jazz.GenerateHarmonicSummary(tune, opts)
			pdfOpts.HarmonicSummary = summary
			pdfOpts.IsLLMSummary = isLLM
		}
		lyStr, err := jazz.GenerateLilyPondFullCompanion(tune, pdfOpts)
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
		fullFlag := compCmd.Bool("full", false, "Include Harmonic Overview pages (summary, trajectory, devices, measure table) in PDF")
		apiKeyFlag := compCmd.String("api-key", "", "Gemini API key for LLM harmonic summary (optional)")
		modelFlag := compCmd.String("model", "gemini-2.5-flash", "Gemini model for LLM harmonic summary")

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

		runCompanion(tune, *outFile, *pdfFlag, *barsFlag, cfg, *fullFlag, jazz.TranscribeOptions{APIKey: *apiKeyFlag, Model: *modelFlag})

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

	case "transcribe":
		transcribeCmd := flag.NewFlagSet("transcribe", flag.ExitOnError)
		outFile := transcribeCmd.String("o", "", "Output companion MusicXML file path")
		pdfFlag := transcribeCmd.Bool("pdf", false, "Compile print-ready companion PDF via LilyPond")
		barsFlag := transcribeCmd.Int("bars", 0, "Measures per line (default: auto)")
		apiKeyFlag := transcribeCmd.String("api-key", "", "Gemini API key (defaults to $GEMINI_API_KEY)")
		modelFlag := transcribeCmd.String("model", "gemini-2.5-flash", "Gemini model to use")
		fullFlag := transcribeCmd.Bool("full", false, "Include Harmonic Overview pages in PDF")

		var target string
		var flagArgs []string
		for _, arg := range os.Args[2:] {
			if !strings.HasPrefix(arg, "-") && target == "" {
				target = arg
			} else {
				flagArgs = append(flagArgs, arg)
			}
		}
		transcribeCmd.Parse(flagArgs)

		if target == "" {
			fmt.Println("Error: please provide a chart image or PDF file to transcribe.")
			fmt.Println("Usage: jazz-tools transcribe <chart.pdf | chart.png | chart.jpg> [-o out.musicxml] [--pdf]")
			os.Exit(1)
		}
		data, err := os.ReadFile(target)
		if err != nil {
			fmt.Printf("Error reading chart file %s: %v\n", target, err)
			os.Exit(1)
		}

		fmt.Printf("Transcribing chart %s using Gemini Vision (%s)...\n", filepath.Base(target), *modelFlag)
		tune, err := jazz.TranscribeChart(data, "", jazz.TranscribeOptions{
			APIKey: *apiKeyFlag,
			Model:  *modelFlag,
		})
		if err != nil {
			fmt.Printf("Transcription error: %v\n", err)
			os.Exit(1)
		}

		runAnalyze(tune)

		if *outFile != "" || *pdfFlag {
			cfg := jazz.DefaultAnnotationConfig()
			runCompanion(tune, *outFile, *pdfFlag, *barsFlag, cfg, *fullFlag, jazz.TranscribeOptions{APIKey: *apiKeyFlag, Model: *modelFlag})
		}

	case "batch":
		batchCmd := flag.NewFlagSet("batch", flag.ExitOnError)
		outDirFlag := batchCmd.String("o", "mobilesheets_companion", "Output directory for PDFs and MobileSheets CSV")
		zipFlag := batchCmd.String("zip", "", "Optional output .zip path instead of directory")
		fullFlag := batchCmd.Bool("full", true, "Include Harmonic Overview pages (summary, trajectory, devices, measure table)")
		transcribeFlag := batchCmd.Bool("transcribe", false, "Always transcribe embedded PDFs via Gemini Vision instead of matching built-in standards")
		barsFlag := batchCmd.Int("bars", 0, "Measures per line (default: auto)")
		apiKeyFlag := batchCmd.String("api-key", "", "Gemini API key (defaults to $GEMINI_API_KEY)")
		modelFlag := batchCmd.String("model", "gemini-2.5-flash", "Gemini model to use")

		var target string
		var flagArgs []string
		for _, arg := range os.Args[2:] {
			if !strings.HasPrefix(arg, "-") && target == "" {
				target = arg
			} else {
				flagArgs = append(flagArgs, arg)
			}
		}
		batchCmd.Parse(flagArgs)

		if target == "" {
			fmt.Println("Error: please provide a MobileSheets .msf setlist file.")
			fmt.Println("Usage: jazz-tools batch <setlist.msf> [-o outdir] [--zip bundle.zip] [--full] [--transcribe]")
			os.Exit(1)
		}

		msfBytes, err := os.ReadFile(target)
		if err != nil {
			fmt.Printf("Error reading .msf file %s: %v\n", target, err)
			os.Exit(1)
		}

		opts := jazz.BatchOptions{
			FullReport:       *fullFlag,
			ForceVision:      *transcribeFlag,
			BarsPerLine:      *barsFlag,
			AnnotationConfig: jazz.DefaultAnnotationConfig(),
			TranscribeOpts: jazz.TranscribeOptions{
				APIKey: *apiKeyFlag,
				Model:  *modelFlag,
			},
			Progress: func(idx, total int, song jazz.MSFSong, status string) {
				fmt.Printf("[%2d/%2d] %s (%s)...\n", idx, total, song.Title, status)
			},
		}

		if *zipFlag != "" {
			zipBytes, setlistName, results, err := jazz.ProcessMSFToZip(msfBytes, opts)
			if err != nil {
				fmt.Printf("Batch error: %v\n", err)
				os.Exit(1)
			}
			if err := os.WriteFile(*zipFlag, zipBytes, 0644); err != nil {
				fmt.Printf("Error writing zip %s: %v\n", *zipFlag, err)
				os.Exit(1)
			}
			printBatchSummary(setlistName, *zipFlag, results)
		} else {
			setlistName, results, err := jazz.ProcessMSFToDir(msfBytes, *outDirFlag, opts)
			if err != nil {
				fmt.Printf("Batch error: %v\n", err)
				os.Exit(1)
			}
			printBatchSummary(setlistName, *outDirFlag, results)
		}

	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		defaultPort := 8080
		if envPort := os.Getenv("PORT"); envPort != "" {
			if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
				defaultPort = p
			}
		}
		portFlag := serveCmd.Int("port", defaultPort, "Port to listen on (default: 8080, or $PORT)")
		hostFlag := serveCmd.String("host", "0.0.0.0", "Host to bind to (default: 0.0.0.0)")
		serveCmd.Parse(os.Args[2:])

		srv, err := server.NewServer(server.Config{
			Host: *hostFlag,
			Port: *portFlag,
		})
		if err != nil {
			fmt.Printf("Error starting server: %v\n", err)
			os.Exit(1)
		}
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("Server exited: %v\n", err)
			os.Exit(1)
		}

	default:
		// Direct file argument without explicit subcommand
		target := os.Args[1]
		ext := strings.ToLower(filepath.Ext(target))
		if ext == ".pdf" || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
			data, err := os.ReadFile(target)
			if err != nil {
				fmt.Printf("Error reading chart file %s: %v\n", target, err)
				os.Exit(1)
			}
			fmt.Printf("Detected visual chart %s. Transcribing using Gemini Vision...\n", filepath.Base(target))
			tune, err := jazz.TranscribeChart(data, "", jazz.TranscribeOptions{})
			if err != nil {
				fmt.Printf("Transcription error: %v\n", err)
				os.Exit(1)
			}
			runAnalyze(tune)
			fmt.Println()
			fmt.Printf("Tip: Run 'jazz-tools transcribe %s --pdf' to generate printable guide tone companion sheets.\n", target)
			return
		}

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

func printBatchSummary(setlistName, dest string, results []jazz.BatchResultItem) {
	succeeded := 0
	for _, r := range results {
		if r.Source != "error" {
			succeeded++
		}
	}
	fmt.Println()
	fmt.Printf("✓ MobileSheets Setlist %q: %d/%d companion PDFs written to %s\n", setlistName, succeeded, len(results), dest)
	for _, r := range results {
		if r.Source == "error" {
			fmt.Printf("  ✗ [%02d] %s: %s\n", r.Index, r.OriginalTitle, r.Error)
		} else {
			fmt.Printf("  ✓ [%02d] %s (%s) -> %s\n", r.Index, r.OriginalTitle, r.Source, r.OutputFilename)
		}
	}
}
