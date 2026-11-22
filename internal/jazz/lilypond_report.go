package jazz

import (
	"fmt"
	"strings"
)

// CompanionPDFOptions controls what is included in the generated LilyPond PDF.
type CompanionPDFOptions struct {
	BarsPerLine           int
	AnnotationConfig      AnnotationConfig
	IncludeAnalysisReport bool
	HarmonicSummary       string
	IsLLMSummary          bool
}

func escapeLilyPondString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

// GenerateLilyPondFullCompanion generates the LilyPond score and, when
// opts.IncludeAnalysisReport is true, appends Song Overview & Harmonic Analysis pages.
func GenerateLilyPondFullCompanion(tune *Tune, opts CompanionPDFOptions) (string, error) {
	baseScore, err := GenerateLilyPondScoreWithAnnotationConfig(tune, opts.BarsPerLine, opts.AnnotationConfig)
	if err != nil {
		return "", err
	}
	if !opts.IncludeAnalysisReport {
		return baseScore, nil
	}

	allChords := tune.AllChords()
	pairs, _ := VoiceLeadChords(allChords)
	AnalyzeTonalCenters(tune)
	spans := GetHarmonicJourneySpans(tune)
	devices := DetectHarmonicDevices(tune)

	summary := strings.TrimSpace(opts.HarmonicSummary)
	if summary == "" {
		summary = BuildDeterministicHarmonicSummary(tune)
	}

	var buf strings.Builder
	buf.WriteString(baseScore)
	buf.WriteString("\n\\pageBreak\n\n")
	buf.WriteString("\\markuplist {\n")
	buf.WriteString("  \\override #'(font-family . sans)\n")

	// Section Header
	buf.WriteString("  \\override #'(baseline-skip . 2.6)\n")
	buf.WriteString("  \\column {\n")
	buf.WriteString(fmt.Sprintf("    \\abs-fontsize #16 \\bold \"%s — Harmonic Overview\"\n", escapeLilyPondString(tune.Title)))
	meta := fmt.Sprintf("Key: %s   |   Time: %d/%d   |   Measures: %d",
		tune.KeyName(), tune.TimeSignature[0], tune.TimeSignature[1], len(tune.Measures))
	if tune.Composer != "" {
		meta = escapeLilyPondString(tune.Composer) + "   |   " + meta
	}
	buf.WriteString(fmt.Sprintf("    \\abs-fontsize #9.5 \\with-color #(rgb-color 0.35 0.40 0.48) \"%s\"\n", meta))
	buf.WriteString("    \\vspace #0.8\n")
	buf.WriteString("  }\n")

	// 1. Harmonic Narrative
	summaryTitle := "Harmonic Structure Summary"
	if opts.IsLLMSummary {
		summaryTitle = "Harmonic Structure Summary (LLM Analysis)"
	}
	buf.WriteString("  \\column {\n")
	buf.WriteString(fmt.Sprintf("    \\abs-fontsize #11 \\bold \"%s\"\n", summaryTitle))
	buf.WriteString("    \\vspace #0.3\n")
	buf.WriteString("    \\abs-fontsize #9.5 \\override #'(line-width . 95) \\wordwrap-string #\"")
	buf.WriteString(escapeLilyPondString(summary))
	buf.WriteString("\"\n")
	buf.WriteString("    \\vspace #0.9\n")
	buf.WriteString("  }\n")

	// 2. Harmonic Modulation Trajectory
	if len(spans) > 0 {
		buf.WriteString("  \\column {\n")
		buf.WriteString("    \\abs-fontsize #11 \\bold \"Harmonic Modulation Trajectory\"\n")
		buf.WriteString("    \\vspace #0.3\n")
		// Render spans wrapped cleanly in rows of up to 4
		for i := 0; i < len(spans); i += 4 {
			end := i + 4
			if end > len(spans) {
				end = len(spans)
			}
			buf.WriteString("    \\line { ")
			for j := i; j < end; j++ {
				sp := spans[j]
				tcInfo := GetTuneTonalCenterInfo(tune, sp.Center)
				barLabel := fmt.Sprintf("bars %d-%d", sp.StartBar, sp.EndBar)
				if sp.StartBar == sp.EndBar {
					barLabel = fmt.Sprintf("bar %d", sp.StartBar)
				}
				buf.WriteString(fmt.Sprintf("\\abs-fontsize #9 \\bold \\with-color #%s \"%s\" \\abs-fontsize #8.5 \\with-color #(rgb-color 0.35 0.40 0.48) \"(%s)\" ",
					tcInfo.LilyPondColor, escapeLilyPondString(sp.Center), barLabel))
				if j < len(spans)-1 {
					buf.WriteString("\\abs-fontsize #9 \"  ->  \" ")
				}
			}
			buf.WriteString("}\n")
		}
		buf.WriteString("    \\vspace #0.9\n")
		buf.WriteString("  }\n")
	}

	// 3. Identified Berklee Harmonic Devices
	if len(devices) > 0 {
		buf.WriteString("  \\column {\n")
		buf.WriteString(fmt.Sprintf("    \\abs-fontsize #11 \\bold \"Identified Berklee Harmonic Devices (%d)\"\n", len(devices)))
		buf.WriteString("    \\vspace #0.3\n")
		for _, d := range devices {
			spanStr := fmt.Sprintf("bars %d-%d", d.StartMeasure+1, d.EndMeasure+1)
			if d.StartMeasure == d.EndMeasure {
				spanStr = fmt.Sprintf("bar %d", d.StartMeasure+1)
			}
			color := d.LilyPondColor
			if color == "" {
				color = "(rgb-color 0.11 0.31 0.85)"
			}
			buf.WriteString(fmt.Sprintf("    \\line { \\abs-fontsize #8.5 \\bold \\with-color #%s \"• %s\" \\abs-fontsize #8.5 \" (%s)\" \\abs-fontsize #8.5 \\with-color #(rgb-color 0.40 0.45 0.52) \"  [Target: %s]\" }\n",
				color, escapeLilyPondString(d.Label), spanStr, escapeLilyPondString(d.TargetKey)))
		}
		buf.WriteString("    \\vspace #0.9\n")
		buf.WriteString("  }\n")
	}

	// 4. Measure-by-Measure Changes & Guide Tones
	buf.WriteString("  \\sans \\abs-fontsize #11 \\bold \"Measure-by-Measure Changes & Guide Tones\"\n")
	buf.WriteString("  \\vspace #0.3\n")
	buf.WriteString("  \\line {\n")
	buf.WriteString("    \\sans \\abs-fontsize #8.5 \\bold \\with-color #(rgb-color 0.35 0.40 0.48) {\n")
	buf.WriteString("      \\pad-to-box #'(0 . 6) #'(0 . 0) \"Bar\"\n")
	buf.WriteString("      \\pad-to-box #'(0 . 6) #'(0 . 0) \"Beat\"\n")
	buf.WriteString("      \\pad-to-box #'(0 . 16) #'(0 . 0) \"Chord\"\n")
	buf.WriteString("      \\pad-to-box #'(0 . 16) #'(0 . 0) \"Voice 1 (3/7)\"\n")
	buf.WriteString("      \\pad-to-box #'(0 . 16) #'(0 . 0) \"Voice 2 (3/7)\"\n")
	buf.WriteString("      \\pad-to-box #'(0 . 14) #'(0 . 0) \"Tonal Center\"\n")
	buf.WriteString("    }\n")
	buf.WriteString("  }\n")
	buf.WriteString("  \\with-color #(rgb-color 0.78 0.82 0.88) \\draw-hline\n")
	buf.WriteString("  \\vspace #0.15\n")

	pairIdx := 0
	for _, m := range tune.Measures {
		for _, tc := range m.Chords {
			var p GuideTonePair
			if pairIdx < len(pairs) {
				p = pairs[pairIdx]
				pairIdx++
			}
			tcInfo := GetTuneTonalCenterInfo(tune, tc.TonalCenter)
			color := tcInfo.LilyPondColor
			if color == "" {
				color = "(rgb-color 0.12 0.16 0.23)"
			}
			buf.WriteString("  \\line {\n")
			buf.WriteString("    \\sans \\abs-fontsize #8.5 {\n")
			buf.WriteString(fmt.Sprintf("      \\pad-to-box #'(0 . 6) #'(0 . 0) \\with-color #(rgb-color 0.40 0.45 0.52) \"%d\"\n", m.Number))
			buf.WriteString(fmt.Sprintf("      \\pad-to-box #'(0 . 6) #'(0 . 0) \\with-color #(rgb-color 0.40 0.45 0.52) \"%.1f\"\n", tc.BeatOffset+1.0))
			buf.WriteString(fmt.Sprintf("      \\pad-to-box #'(0 . 16) #'(0 . 0) \\bold \"%s\"\n", escapeLilyPondString(tc.Chord.String())))
			buf.WriteString(fmt.Sprintf("      \\pad-to-box #'(0 . 16) #'(0 . 0) \"%s\"\n", escapeLilyPondString(p.Voice1.String())))
			buf.WriteString(fmt.Sprintf("      \\pad-to-box #'(0 . 16) #'(0 . 0) \"%s\"\n", escapeLilyPondString(p.Voice2.String())))
			buf.WriteString(fmt.Sprintf("      \\pad-to-box #'(0 . 14) #'(0 . 0) \\bold \\with-color #%s \"%s\"\n", color, escapeLilyPondString(tc.TonalCenter)))
			buf.WriteString("    }\n")
			buf.WriteString("  }\n")
		}
		buf.WriteString("  \\with-color #(rgb-color 0.91 0.93 0.96) \\draw-hline\n")
	}

	buf.WriteString("}\n")
	return buf.String(), nil
}
