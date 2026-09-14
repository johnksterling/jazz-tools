package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// lilypondPitch formats a Pitch into LilyPond notation syntax (e.g. ees' for Eb4).
func lilypondPitch(p Pitch) string {
	stepChar := strings.ToLower(string(p.Step))

	var acc string
	switch p.Alter {
	case -2:
		acc = "eses"
	case -1:
		acc = "es"
	case 1:
		acc = "is"
	case 2:
		acc = "isis"
	}

	oct := p.Octave
	if oct == 0 {
		oct = 4
	}

	var octMark string
	if oct >= 4 {
		octMark = strings.Repeat("'", oct-3)
	} else if oct < 3 {
		octMark = strings.Repeat(",", 3-oct)
	}

	return stepChar + acc + octMark
}

// lilypondDuration converts beat duration into LilyPond duration string (e.g. "1", "2.", "2", "4.", "4", "8").
func lilypondDuration(beats float64) string {
	switch {
	case beats >= 3.75:
		return "1"
	case beats >= 2.75:
		return "2."
	case beats >= 1.75:
		return "2"
	case beats >= 1.25:
		return "4."
	case beats >= 0.75:
		return "4"
	case beats >= 0.6:
		return "8."
	case beats >= 0.4:
		return "8"
	default:
		return "16"
	}
}

// lilypondChordName formats a chord into LilyPond chordmode syntax (e.g. ees4:maj7).
func lilypondChordName(c Chord, beats float64) string {
	stepChar := strings.ToLower(string(c.Root.Step))
	var acc string
	switch c.Root.Alter {
	case -2:
		acc = "eses"
	case -1:
		acc = "es"
	case 1:
		acc = "is"
	case 2:
		acc = "isis"
	}
	dur := lilypondDuration(beats)
	qual := ""
	switch c.Quality {
	case QualityMajor7:
		qual = ":maj7"
	case QualityDominant7:
		qual = ":7"
	case QualityMinor7:
		qual = ":m7"
	case QualityHalfDiminished:
		qual = ":m7.5-"
	case QualityDiminished7:
		qual = ":dim7"
	case QualityMinorMajor7:
		qual = ":m7+"
	case QualityMajor6:
		qual = ":6"
	case QualityMinor6:
		qual = ":m6"
	case QualitySuspended4:
		qual = ":sus4"
	case QualityAugmentedTriad:
		qual = ":aug"
	case QualityDiminishedTriad:
		qual = ":dim"
	case QualityMinorTriad:
		qual = ":m"
	}
	return stepChar + acc + dur + qual
}

// DetermineBarsPerLine calculates the optimal measures per line to fit the tune on 1 page.
func DetermineBarsPerLine(tune *Tune, userBars int) int {
	if userBars > 0 {
		return userBars
	}

	numMeasures := len(tune.Measures)
	if numMeasures == 0 {
		return 8
	}

	// For tunes with up to 24 measures (e.g. 12-bar blues, 16-bar tunes, and 24-bar forms like Autumn Leaves),
	// 4 bars per line produces 3 to 6 balanced systems that fit comfortably on 1 page.
	if numMeasures <= 24 {
		return 4
	}

	// For standard jazz forms (> 24 bars), default to 8 measures per line.
	// Maximum comfortable systems on a single page with chords and key badges is ~7.
	maxSystems := 7
	bars := 8

	// If 8 bars per line would exceed maxSystems (causing more than 1 page),
	// increase bars per line to fit on a single page.
	systems := (numMeasures + bars - 1) / bars
	if systems > maxSystems {
		needed := (numMeasures + maxSystems - 1) / maxSystems
		if needed%2 != 0 {
			needed++
		}
		bars = needed
	}

	return bars
}

// AnnotationConfig controls which companion sheet annotations are rendered.
type AnnotationConfig struct {
	ShowKeys    bool // Tonal center badge at start of contiguous section
	ShowDevices bool // Berklee-style harmonic devices (ii-V brackets, arrows, etc.)
}

// DefaultAnnotationConfig returns the default configuration with all annotations enabled.
func DefaultAnnotationConfig() AnnotationConfig {
	return AnnotationConfig{
		ShowKeys:    true,
		ShowDevices: true,
	}
}

// GenerateLilyPondScore generates a complete LilyPond score file for printing companion sheet music using auto-layout.
func GenerateLilyPondScore(tune *Tune) (string, error) {
	return GenerateLilyPondScoreWithConfig(tune, 0)
}

// GenerateLilyPondScoreWithConfig generates a complete LilyPond score file with a configurable measures-per-line setting.
func GenerateLilyPondScoreWithConfig(tune *Tune, userBars int) (string, error) {
	return GenerateLilyPondScoreWithAnnotationConfig(tune, userBars, DefaultAnnotationConfig())
}

// GenerateLilyPondScoreWithAnnotationConfig generates a complete LilyPond score file with configurable line layout and annotations.
func GenerateLilyPondScoreWithAnnotationConfig(tune *Tune, userBars int, cfg AnnotationConfig) (string, error) {
	// Run tonal center analysis
	AnalyzeTonalCenters(tune)

	// Ensure voice leading is run
	var chords []Chord
	for _, m := range tune.Measures {
		for _, tc := range m.Chords {
			chords = append(chords, tc.Chord)
		}
	}
	pairs, err := VoiceLeadChords(chords)
	if err != nil {
		return "", err
	}
	pairIdx := 0
	for mIdx := range tune.Measures {
		for cIdx := range tune.Measures[mIdx].Chords {
			if pairIdx < len(pairs) {
				tune.Measures[mIdx].Chords[cIdx].GuideTones = &pairs[pairIdx]
				pairIdx++
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteString("\\version \"2.24.0\"\n\n")

	// Paper layout
	buf.WriteString("\\paper {\n")
	buf.WriteString("  indent = 0\\mm\n")
	buf.WriteString("  ragged-right = ##f\n")
	buf.WriteString("  ragged-bottom = ##t\n")
	buf.WriteString("  ragged-last-bottom = ##t\n")
	buf.WriteString("  page-count = #1\n")
	buf.WriteString("  system-system-spacing =\n")
	buf.WriteString("    #'((basic-distance . 16)\n")
	buf.WriteString("       (minimum-distance . 12)\n")
	buf.WriteString("       (padding . 4)\n")
	buf.WriteString("       (stretchability . 10))\n")
	buf.WriteString("}\n\n")

	// Header
	buf.WriteString("\\header {\n")
	title := tune.Title
	if title == "" {
		title = "Lead Sheet"
	}
	buf.WriteString(fmt.Sprintf("  title = \"%s\"\n", title))
	buf.WriteString("  subtitle = \"Guide Tones Companion (3rd & 7th)\"\n")
	if tune.Composer != "" {
		buf.WriteString(fmt.Sprintf("  composer = \"%s\"\n", tune.Composer))
	}
	buf.WriteString("}\n\n")

	// Detect harmonic devices if requested
	var devices []HarmonicDevice
	if cfg.ShowDevices {
		devices = DetectHarmonicDevices(tune)
	}

	barsPerLine := DetermineBarsPerLine(tune, userBars)

	if len(devices) > 0 {
		buf.WriteString("deviceAnnotations = {\n")
		buf.WriteString(generateDevicesTrack(tune, devices, barsPerLine))
		buf.WriteString("}\n\n")
	}

	// Generate Chord Names
	var chordBuf bytes.Buffer
	var upperBuf bytes.Buffer
	var lowerBuf bytes.Buffer

	for mIdx, m := range tune.Measures {
		beats := float64(m.TimeBeats)
		if beats == 0 {
			beats = 4.0
		}

		isLineBreak := barsPerLine > 0 && (mIdx+1)%barsPerLine == 0 && mIdx < len(tune.Measures)-1
		breakSuffix := ""
		if isLineBreak {
			breakSuffix = " \\break"
		}

		if len(m.Chords) == 0 {
			chordBuf.WriteString(fmt.Sprintf("  r%s |\n", lilypondDuration(beats)))
			upperBuf.WriteString(fmt.Sprintf("  R%s |%s\n", lilypondDuration(beats), breakSuffix))
			lowerBuf.WriteString(fmt.Sprintf("  R%s |\n", lilypondDuration(beats)))
		} else {
			// Chords
			cPos := 0.0
			chordBuf.WriteString("  ")
			for _, tc := range m.Chords {
				if tc.IsTonalCenterChange && tc.TonalCenterColor != "" {
					chordBuf.WriteString(fmt.Sprintf("\\override ChordNames.ChordName.color = #%s ", tc.TonalCenterColor))
				}
				if tc.BeatOffset > cPos {
					restBeats := tc.BeatOffset - cPos
					chordBuf.WriteString(fmt.Sprintf("r%s ", lilypondDuration(restBeats)))
					cPos += restBeats
				}
				chordBuf.WriteString(fmt.Sprintf("%s ", lilypondChordName(tc.Chord, tc.DurationBeats)))
				cPos += tc.DurationBeats
			}
			if cPos < beats {
				chordBuf.WriteString(fmt.Sprintf("r%s ", lilypondDuration(beats-cPos)))
			}
			chordBuf.WriteString("|\n")

			// Voice 1
			v1Pos := 0.0
			upperBuf.WriteString("  ")
			for _, tc := range m.Chords {
				if tc.BeatOffset > v1Pos {
					restBeats := tc.BeatOffset - v1Pos
					upperBuf.WriteString(fmt.Sprintf("r%s ", lilypondDuration(restBeats)))
					v1Pos += restBeats
				}
				p := tc.GuideTones.Voice1
				dur := tc.DurationBeats
				colorPrefix := ""
				if tc.TonalCenterColor != "" {
					colorPrefix = fmt.Sprintf("\\tweak color #%s ", tc.TonalCenterColor)
				}
				markupSuffix := ""
				if cfg.ShowKeys && tc.IsTonalCenterChange && tc.TonalCenter != "" {
					markupSuffix = fmt.Sprintf("^\\markup { \\with-color #%s \\rounded-box \\bold \\fontsize #-2 \"Key: %s\" }", tc.TonalCenterColor, tc.TonalCenter)
				}
				upperBuf.WriteString(fmt.Sprintf("%s%s%s%s ", colorPrefix, lilypondPitch(p), lilypondDuration(dur), markupSuffix))
				v1Pos += dur
			}
			if v1Pos < beats {
				upperBuf.WriteString(fmt.Sprintf("r%s ", lilypondDuration(beats-v1Pos)))
			}
			upperBuf.WriteString(fmt.Sprintf("|%s\n", breakSuffix))

			// Voice 2
			v2Pos := 0.0
			lowerBuf.WriteString("  ")
			for _, tc := range m.Chords {
				if tc.BeatOffset > v2Pos {
					restBeats := tc.BeatOffset - v2Pos
					lowerBuf.WriteString(fmt.Sprintf("r%s ", lilypondDuration(restBeats)))
					v2Pos += restBeats
				}
				p := tc.GuideTones.Voice2
				dur := tc.DurationBeats
				colorPrefix := ""
				if tc.TonalCenterColor != "" {
					colorPrefix = fmt.Sprintf("\\tweak color #%s ", tc.TonalCenterColor)
				}
				lowerBuf.WriteString(fmt.Sprintf("%s%s%s ", colorPrefix, lilypondPitch(p), lilypondDuration(dur)))
				v2Pos += dur
			}
			if v2Pos < beats {
				lowerBuf.WriteString(fmt.Sprintf("r%s ", lilypondDuration(beats-v2Pos)))
			}
			lowerBuf.WriteString("|\n")
		}
	}

	buf.WriteString("theChords = \\chordmode {\n")
	buf.WriteString(chordBuf.String())
	buf.WriteString("}\n\n")

	buf.WriteString("voiceUpper = {\n")
	buf.WriteString(upperBuf.String())
	buf.WriteString("}\n\n")

	buf.WriteString("voiceLower = {\n")
	buf.WriteString(lowerBuf.String())
	buf.WriteString("}\n\n")

	// Key signature formatting
	keyLily := "c \\major"
	if len(tune.Measures) > 0 {
		m := tune.Measures[0]
		keyLily = fifthsToLilyPondKey(m.KeyFifths, m.KeyMode)
	}

	timeBeats := tune.TimeSignature[0]
	timeBeatType := tune.TimeSignature[1]
	if timeBeats == 0 {
		timeBeats = 4
	}
	if timeBeatType == 0 {
		timeBeatType = 4
	}

	buf.WriteString("\\score {\n")
	buf.WriteString("  <<\n")
	if len(devices) > 0 {
		buf.WriteString("    \\new Dynamics { \\deviceAnnotations }\n")
	}
	buf.WriteString("    \\new ChordNames { \\theChords }\n")
	buf.WriteString("    \\new Staff {\n")
	buf.WriteString("      \\clef treble\n")
	buf.WriteString(fmt.Sprintf("      \\key %s\n", keyLily))
	buf.WriteString(fmt.Sprintf("      \\time %d/%d\n", timeBeats, timeBeatType))
	buf.WriteString("      <<\n")
	buf.WriteString("        \\new Voice { \\voiceOne \\voiceUpper }\n")
	buf.WriteString("        \\new Voice { \\voiceTwo \\voiceLower }\n")
	buf.WriteString("      >>\n")
	buf.WriteString("    }\n")
	buf.WriteString("  >>\n")
	buf.WriteString("  \\layout { }\n")
	buf.WriteString("}\n")

	return buf.String(), nil
}

// generateDevicesTrack generates a LilyPond Dynamics voice with TextSpanner brackets for detected devices.
func generateDevicesTrack(tune *Tune, devices []HarmonicDevice, barsPerLine int) string {
	var buf bytes.Buffer
	buf.WriteString("  \\override TextSpanner.direction = #UP\n")
	buf.WriteString("  \\override TextSpanner.outside-staff-priority = ##f\n")

	keyForPos := func(mIdx int, beat float64) string {
		return fmt.Sprintf("%d:%.2f", mIdx, beat)
	}

	stopsAt := make(map[string][]HarmonicDevice)
	startsAt := make(map[string][]HarmonicDevice)

	for _, d := range devices {
		stopsAt[keyForPos(d.EndMeasure, d.EndBeat)] = append(stopsAt[keyForPos(d.EndMeasure, d.EndBeat)], d)
		startsAt[keyForPos(d.StartMeasure, d.StartBeat)] = append(startsAt[keyForPos(d.StartMeasure, d.StartBeat)], d)
	}

	activeSpanner := false

	for mIdx, m := range tune.Measures {
		beats := float64(m.TimeBeats)
		if beats == 0 {
			beats = 4.0
		}

		isLineBreak := barsPerLine > 0 && (mIdx+1)%barsPerLine == 0 && mIdx < len(tune.Measures)-1
		breakSuffix := ""
		if isLineBreak {
			breakSuffix = " \\break"
		}

		buf.WriteString("  ")

		if len(m.Chords) == 0 {
			stopDevs := stopsAt[keyForPos(mIdx, 0.0)]
			startDevs := startsAt[keyForPos(mIdx, 0.0)]

			shouldStop := activeSpanner && len(stopDevs) > 0
			shouldStart := len(startDevs) > 0

			if shouldStop && shouldStart {
				d := startDevs[0]
				writeSpannerOverrides(&buf, d)
				mult := int(beats*4.0 - 1)
				if mult < 1 {
					mult = 1
				}
				buf.WriteString(fmt.Sprintf("s16\\stopTextSpan s16*%d\\startTextSpan |%s\n", mult, breakSuffix))
				activeSpanner = true
			} else if shouldStop {
				buf.WriteString(fmt.Sprintf("s%s\\stopTextSpan |%s\n", lilypondDuration(beats), breakSuffix))
				activeSpanner = false
			} else if shouldStart {
				d := startDevs[0]
				writeSpannerOverrides(&buf, d)
				buf.WriteString(fmt.Sprintf("s%s\\startTextSpan |%s\n", lilypondDuration(beats), breakSuffix))
				activeSpanner = true
			} else {
				buf.WriteString(fmt.Sprintf("s%s |%s\n", lilypondDuration(beats), breakSuffix))
			}
			continue
		}

		cPos := 0.0
		for _, tc := range m.Chords {
			if tc.BeatOffset > cPos {
				gap := tc.BeatOffset - cPos
				buf.WriteString(fmt.Sprintf("s%s ", lilypondDuration(gap)))
				cPos += gap
			}

			stopDevs := stopsAt[keyForPos(mIdx, tc.BeatOffset)]
			startDevs := startsAt[keyForPos(mIdx, tc.BeatOffset)]
			dur := tc.DurationBeats

			shouldStop := activeSpanner && len(stopDevs) > 0
			shouldStart := len(startDevs) > 0

			if shouldStop && shouldStart {
				d := startDevs[0]
				writeSpannerOverrides(&buf, d)
				mult := int(dur*4.0 - 1)
				if mult < 1 {
					mult = 1
				}
				buf.WriteString(fmt.Sprintf("s16\\stopTextSpan s16*%d\\startTextSpan ", mult))
				activeSpanner = true
			} else if shouldStop {
				buf.WriteString(fmt.Sprintf("s%s\\stopTextSpan ", lilypondDuration(dur)))
				activeSpanner = false
			} else if shouldStart {
				d := startDevs[0]
				writeSpannerOverrides(&buf, d)
				// Check if this same chord also stops the device at its end
				if len(stopsAt[keyForPos(mIdx, tc.BeatOffset+dur)]) > 0 {
					mult := int(dur*4.0 - 1)
					if mult < 1 {
						buf.WriteString(fmt.Sprintf("s%s\\startTextSpan\\stopTextSpan ", lilypondDuration(dur)))
					} else {
						buf.WriteString(fmt.Sprintf("s16*%d\\startTextSpan s16\\stopTextSpan ", mult))
					}
					activeSpanner = false
				} else {
					buf.WriteString(fmt.Sprintf("s%s\\startTextSpan ", lilypondDuration(dur)))
					activeSpanner = true
				}
			} else if activeSpanner && len(stopsAt[keyForPos(mIdx, tc.BeatOffset+dur)]) > 0 {
				mult := int(dur*4.0 - 1)
				if mult < 1 {
					buf.WriteString(fmt.Sprintf("s%s\\stopTextSpan ", lilypondDuration(dur)))
				} else {
					buf.WriteString(fmt.Sprintf("s16*%d s16\\stopTextSpan ", mult))
				}
				activeSpanner = false
			} else {
				buf.WriteString(fmt.Sprintf("s%s ", lilypondDuration(dur)))
			}
			cPos += dur
		}

		if cPos < beats {
			buf.WriteString(fmt.Sprintf("s%s ", lilypondDuration(beats-cPos)))
		}
		buf.WriteString(fmt.Sprintf("|%s\n", breakSuffix))
	}

	return buf.String()
}

// writeSpannerOverrides writes the LilyPond property overrides for a device's TextSpanner.
func writeSpannerOverrides(buf *bytes.Buffer, d HarmonicDevice) {
	color := d.LilyPondColor
	if color == "" {
		color = "(rgb-color 0.85 0.47 0.02)"
	}
	buf.WriteString(fmt.Sprintf("\\override TextSpanner.color = #%s ", color))
	buf.WriteString("\\override TextSpanner.thickness = #1.5 ")
	buf.WriteString("\\override TextSpanner.bound-details.left.stencil-align-dir-y = #CENTER ")
	buf.WriteString("\\override TextSpanner.bound-details.left-broken.text = ##f ")
	buf.WriteString(fmt.Sprintf("\\override TextSpanner.bound-details.left.text = \\markup { \\rounded-box \\pad-markup #0.2 \\with-color #%s \\bold \\fontsize #-3 \"%s\" } ", color, d.Label))
	if d.Resolves {
		buf.WriteString("\\override TextSpanner.bound-details.right.arrow = ##t ")
		buf.WriteString("\\override TextSpanner.bound-details.right.text = ##f ")
		buf.WriteString("\\override TextSpanner.bound-details.right.padding = #1.0 ")
	} else {
		buf.WriteString("\\override TextSpanner.bound-details.right.arrow = ##f ")
		buf.WriteString("\\override TextSpanner.bound-details.right.text = \\markup { \\draw-line #'(0 . -0.8) } ")
		buf.WriteString("\\override TextSpanner.bound-details.right.padding = #0.5 ")
	}
	if d.IsDashed {
		buf.WriteString("\\override TextSpanner.style = #'dashed-line ")
	} else {
		buf.WriteString("\\override TextSpanner.style = #'solid ")
	}
}

// fifthsToLilyPondKey maps circle of fifths and mode to LilyPond key syntax.
func fifthsToLilyPondKey(fifths int, mode string) string {
	majorKeys := map[int]string{
		0: "c \\major", 1: "g \\major", 2: "d \\major", 3: "a \\major",
		4: "e \\major", 5: "b \\major", 6: "fis \\major", 7: "cis \\major",
		-1: "f \\major", -2: "bes \\major", -3: "ees \\major", -4: "aes \\major",
		-5: "des \\major", -6: "ges \\major", -7: "ces \\major",
	}
	minorKeys := map[int]string{
		0: "a \\minor", 1: "e \\minor", 2: "b \\minor", 3: "fis \\minor",
		4: "cis \\minor", 5: "gis \\minor", 6: "dis \\minor", 7: "ais \\minor",
		-1: "d \\minor", -2: "g \\minor", -3: "c \\minor", -4: "f \\minor",
		-5: "bes \\minor", -6: "ees \\minor", -7: "aes \\minor",
	}
	if mode == "minor" {
		if k, ok := minorKeys[fifths]; ok {
			return k
		}
		return "a \\minor"
	}
	if k, ok := majorKeys[fifths]; ok {
		return k
	}
	return "c \\major"
}

// CompileLilyPondToPDF invokes the system lilypond compiler to produce a print-ready PDF.
func CompileLilyPondToPDF(lyFile string, outDir string) error {
	base := strings.TrimSuffix(filepath.Base(lyFile), filepath.Ext(lyFile))
	var targetOut string
	if outDir != "" && outDir != "." {
		targetOut = filepath.Join(outDir, base)
	} else {
		targetOut = base
	}

	lilyPath, err := exec.LookPath("lilypond")
	if err != nil {
		if _, errOpt := os.Stat("/opt/homebrew/bin/lilypond"); errOpt == nil {
			lilyPath = "/opt/homebrew/bin/lilypond"
		} else if _, errLocal := os.Stat("/usr/local/bin/lilypond"); errLocal == nil {
			lilyPath = "/usr/local/bin/lilypond"
		} else {
			return fmt.Errorf("lilypond executable not found in PATH, /opt/homebrew/bin, or /usr/local/bin")
		}
	}

	cmd := exec.Command(lilyPath, "--pdf", "-o", targetOut, lyFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lilypond compilation failed: %v\nOutput: %s", err, string(out))
	}
	return nil
}
