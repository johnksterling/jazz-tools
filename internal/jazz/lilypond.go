package jazz

import (
	"bytes"
	"fmt"
	"math"
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
	case beats > 4.25:
		if math.Abs(beats-float64(int(beats))) < 0.01 {
			return fmt.Sprintf("1*%d/4", int(beats))
		}
		return fmt.Sprintf("1*%.2g/4", beats)
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
	case beats >= 0.2:
		return "16"
	case beats >= 0.1:
		return "32"
	default:
		return "16"
	}
}

// fullMeasureRest returns a whole-measure rest notation for the given meter (e.g. R1, R1*5/4, R1*3/4).
func fullMeasureRest(timeBeats, timeBeatType int) string {
	if timeBeats == 0 {
		timeBeats = 4
	}
	if timeBeatType == 0 {
		timeBeatType = 4
	}
	if timeBeats == 4 && timeBeatType == 4 {
		return "R1"
	}
	return fmt.Sprintf("R1*%d/%d", timeBeats, timeBeatType)
}

// lilypondChordNameWithDuration formats a chord into LilyPond chordmode syntax with an explicit duration string.
func lilypondChordNameWithDuration(c Chord, dur string) string {
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

// lilypondChordName formats a chord into LilyPond chordmode syntax (e.g. ees4:maj7).
func lilypondChordName(c Chord, beats float64) string {
	return lilypondChordNameWithDuration(c, lilypondDuration(beats))
}

// lilypondSkip converts beat duration into LilyPond skip syntax (e.g. s1, s2, s4, or s16*N for odd lengths).
func lilypondSkip(beats float64) string {
	sixteenths := int(math.Round(beats * 4.0))
	if sixteenths <= 0 {
		return "s16"
	}
	switch sixteenths {
	case 16:
		return "s1"
	case 12:
		return "s2."
	case 8:
		return "s2"
	case 6:
		return "s4."
	case 4:
		return "s4"
	case 3:
		return "s8."
	case 2:
		return "s8"
	case 1:
		return "s16"
	default:
		return fmt.Sprintf("s16*%d", sixteenths)
	}
}

// decomposeDuration breaks a duration in quarter note beats into standard LilyPond duration tokens.
func decomposeDuration(beats float64) []string {
	sixteenths := int(math.Round(beats * 4.0))
	var durs []string
	for sixteenths > 0 {
		switch {
		case sixteenths >= 16:
			durs = append(durs, "1")
			sixteenths -= 16
		case sixteenths >= 12:
			durs = append(durs, "2.")
			sixteenths -= 12
		case sixteenths >= 8:
			durs = append(durs, "2")
			sixteenths -= 8
		case sixteenths >= 6:
			durs = append(durs, "4.")
			sixteenths -= 6
		case sixteenths >= 4:
			durs = append(durs, "4")
			sixteenths -= 4
		case sixteenths >= 3:
			durs = append(durs, "8.")
			sixteenths -= 3
		case sixteenths >= 2:
			durs = append(durs, "8")
			sixteenths -= 2
		default:
			durs = append(durs, "16")
			sixteenths -= 1
		}
	}
	if len(durs) == 0 {
		return []string{"4"}
	}
	return durs
}

// durationToSixteenths converts a LilyPond duration token ("1", "2.", "2", "4.", "4", "8.", "8", "16") to sixteenth units.
func durationToSixteenths(dur string) int {
	switch dur {
	case "1":
		return 16
	case "2.":
		return 12
	case "2":
		return 8
	case "4.":
		return 6
	case "4":
		return 4
	case "8.":
		return 3
	case "8":
		return 2
	case "16":
		return 1
	default:
		return 1
	}
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

// DefaultAnnotationConfig returns the default configuration with harmonic devices enabled and key badges disabled.
func DefaultAnnotationConfig() AnnotationConfig {
	return AnnotationConfig{
		ShowKeys:    false,
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

	hasMelody := tune.HasMelody()
	barsPerLine := DetermineBarsPerLine(tune, userBars)
	numSystems := 1
	if barsPerLine > 0 && len(tune.Measures) > 0 {
		numSystems = (len(tune.Measures) + barsPerLine - 1) / barsPerLine
	}
	canFitOnePage := !hasMelody && numSystems <= 7

	var buf bytes.Buffer
	buf.WriteString("\\version \"2.24.0\"\n\n")

	// Paper layout
	buf.WriteString("\\paper {\n")
	buf.WriteString("  indent = 0\\mm\n")
	buf.WriteString("  ragged-right = ##f\n")
	buf.WriteString("  ragged-bottom = ##t\n")
	buf.WriteString("  ragged-last-bottom = ##t\n")
	if canFitOnePage {
		buf.WriteString("  page-count = #1\n")
	}
	buf.WriteString("  system-system-spacing =\n")
	buf.WriteString("    #'((basic-distance . 16)\n")
	buf.WriteString("       (minimum-distance . 12)\n")
	buf.WriteString("       (padding . 5)\n")
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
			upperBuf.WriteString(fmt.Sprintf("  %s |%s\n", fullMeasureRest(m.TimeBeats, m.TimeBeatType), breakSuffix))
			lowerBuf.WriteString(fmt.Sprintf("  %s |\n", fullMeasureRest(m.TimeBeats, m.TimeBeatType)))
		} else {
			// Chords
			cPos := 0.0
			chordBuf.WriteString("  ")
			for _, tc := range m.Chords {
				if tc.IsTonalCenterChange && tc.TonalCenterColor != "" {
					chordBuf.WriteString(fmt.Sprintf("\\override ChordNames.ChordName.color = #%s ", tc.TonalCenterColor))
				}
				if tc.BeatOffset > cPos+0.05 {
					writeLilyPondRests(&chordBuf, tc.BeatOffset-cPos)
					cPos = tc.BeatOffset
				}
				parts := decomposeDuration(tc.DurationBeats)
				for pIdx, d := range parts {
					tie := ""
					if pIdx < len(parts)-1 {
						tie = " ~"
					}
					chordBuf.WriteString(fmt.Sprintf("%s%s ", lilypondChordNameWithDuration(tc.Chord, d), tie))
				}
				cPos += tc.DurationBeats
			}
			if beats-cPos > 0.05 {
				writeLilyPondRests(&chordBuf, beats-cPos)
			}
			chordBuf.WriteString("|\n")

			// Voice 1
			v1Pos := 0.0
			upperBuf.WriteString("  ")
			for _, tc := range m.Chords {
				if tc.BeatOffset > v1Pos+0.05 {
					writeLilyPondRests(&upperBuf, tc.BeatOffset-v1Pos)
					v1Pos = tc.BeatOffset
				}
				p := tc.GuideTones.Voice1
				colorPrefix := ""
				if tc.TonalCenterColor != "" {
					colorPrefix = fmt.Sprintf("\\tweak color #%s ", tc.TonalCenterColor)
				}
				markupSuffix := ""
				if cfg.ShowKeys && tc.IsTonalCenterChange && tc.TonalCenter != "" {
					markupSuffix = fmt.Sprintf("^\\markup { \\with-color #%s \\rounded-box \\bold \\fontsize #-2 \"%s\" }", tc.TonalCenterColor, tc.TonalCenter)
				}
				parts := decomposeDuration(tc.DurationBeats)
				for pIdx, d := range parts {
					tie := ""
					if pIdx < len(parts)-1 {
						tie = " ~"
					}
					cp := ""
					ms := ""
					if pIdx == 0 {
						cp = colorPrefix
						ms = markupSuffix
					}
					upperBuf.WriteString(fmt.Sprintf("%s%s%s%s%s ", cp, lilypondPitch(p), d, tie, ms))
				}
				v1Pos += tc.DurationBeats
			}
			if beats-v1Pos > 0.05 {
				writeLilyPondRests(&upperBuf, beats-v1Pos)
			}
			upperBuf.WriteString(fmt.Sprintf("|%s\n", breakSuffix))

			// Voice 2
			v2Pos := 0.0
			lowerBuf.WriteString("  ")
			for _, tc := range m.Chords {
				if tc.BeatOffset > v2Pos+0.05 {
					writeLilyPondRests(&lowerBuf, tc.BeatOffset-v2Pos)
					v2Pos = tc.BeatOffset
				}
				p := tc.GuideTones.Voice2
				colorPrefix := ""
				if tc.TonalCenterColor != "" {
					colorPrefix = fmt.Sprintf("\\tweak color #%s ", tc.TonalCenterColor)
				}
				parts := decomposeDuration(tc.DurationBeats)
				for pIdx, d := range parts {
					tie := ""
					if pIdx < len(parts)-1 {
						tie = " ~"
					}
					cp := ""
					if pIdx == 0 {
						cp = colorPrefix
					}
					lowerBuf.WriteString(fmt.Sprintf("%s%s%s%s ", cp, lilypondPitch(p), d, tie))
				}
				v2Pos += tc.DurationBeats
			}
			if beats-v2Pos > 0.05 {
				writeLilyPondRests(&lowerBuf, beats-v2Pos)
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

	if hasMelody {
		buf.WriteString("melodyVoice = {\n")
		buf.WriteString(generateMelodyTrack(tune, barsPerLine))
		buf.WriteString("}\n\n")
	}

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
		buf.WriteString("    \\new Dynamics \\with {\n")
		buf.WriteString("      \\override VerticalAxisGroup.nonstaff-relatedstaff-spacing =\n")
		buf.WriteString("        #'((basic-distance . 4)\n")
		buf.WriteString("           (minimum-distance . 3)\n")
		buf.WriteString("           (padding . 1.5))\n")
		buf.WriteString("    } { \\deviceAnnotations }\n")
	}
	buf.WriteString("    \\new ChordNames \\with {\n")
	buf.WriteString("      \\override VerticalAxisGroup.nonstaff-relatedstaff-spacing =\n")
	buf.WriteString("        #'((basic-distance . 5)\n")
	buf.WriteString("           (minimum-distance . 4)\n")
	buf.WriteString("           (padding . 2))\n")
	buf.WriteString("    } { \\theChords }\n")
	if hasMelody {
		buf.WriteString("    \\new GrandStaff \\with {\n")
		buf.WriteString("      \\override StaffGrouper.staff-staff-spacing =\n")
		buf.WriteString("        #'((basic-distance . 11)\n")
		buf.WriteString("           (minimum-distance . 8)\n")
		buf.WriteString("           (padding . 3.5)\n")
		buf.WriteString("           (stretchability . 4))\n")
		buf.WriteString("      \\override StaffGrouper.staffgroup-staff-spacing =\n")
		buf.WriteString("        #'((basic-distance . 15)\n")
		buf.WriteString("           (minimum-distance . 11)\n")
		buf.WriteString("           (padding . 5)\n")
		buf.WriteString("           (stretchability . 8))\n")
		buf.WriteString("    } <<\n")
		buf.WriteString("      \\new Staff \\with { instrumentName = #\"Melody\" shortInstrumentName = #\"Mel.\" } {\n")
		buf.WriteString("        \\clef treble\n")
		buf.WriteString(fmt.Sprintf("        \\key %s\n", keyLily))
		buf.WriteString(fmt.Sprintf("        \\time %d/%d\n", timeBeats, timeBeatType))
		buf.WriteString("        \\new Voice { \\melodyVoice }\n")
		buf.WriteString("      }\n")
		buf.WriteString("      \\new Staff \\with { instrumentName = #\"Guide Tones\" shortInstrumentName = #\"G.T.\" } {\n")
		buf.WriteString("        \\clef treble\n")
		buf.WriteString(fmt.Sprintf("        \\key %s\n", keyLily))
		buf.WriteString(fmt.Sprintf("        \\time %d/%d\n", timeBeats, timeBeatType))
		buf.WriteString("        <<\n")
		buf.WriteString("          \\new Voice { \\voiceOne \\voiceUpper }\n")
		buf.WriteString("          \\new Voice { \\voiceTwo \\voiceLower }\n")
		buf.WriteString("        >>\n")
		buf.WriteString("      }\n")
		buf.WriteString("    >>\n")
	} else {
		buf.WriteString("    \\new Staff {\n")
		buf.WriteString("      \\clef treble\n")
		buf.WriteString(fmt.Sprintf("      \\key %s\n", keyLily))
		buf.WriteString(fmt.Sprintf("      \\time %d/%d\n", timeBeats, timeBeatType))
		buf.WriteString("      <<\n")
		buf.WriteString("        \\new Voice { \\voiceOne \\voiceUpper }\n")
		buf.WriteString("        \\new Voice { \\voiceTwo \\voiceLower }\n")
		buf.WriteString("      >>\n")
		buf.WriteString("    }\n")
	}
	buf.WriteString("  >>\n")
	buf.WriteString("  \\layout { }\n")
	buf.WriteString("}\n")

	return buf.String(), nil
}

// writeLilyPondRests decomposes a remainder duration (in quarter beats) into standard LilyPond rests.
func writeLilyPondRests(buf *bytes.Buffer, rem float64) {
	sixteenths := int(math.Round(rem * 4.0))
	for sixteenths > 0 {
		switch {
		case sixteenths >= 16:
			buf.WriteString("r1 ")
			sixteenths -= 16
		case sixteenths >= 12:
			buf.WriteString("r2. ")
			sixteenths -= 12
		case sixteenths >= 8:
			buf.WriteString("r2 ")
			sixteenths -= 8
		case sixteenths >= 6:
			buf.WriteString("r4. ")
			sixteenths -= 6
		case sixteenths >= 4:
			buf.WriteString("r4 ")
			sixteenths -= 4
		case sixteenths >= 3:
			buf.WriteString("r8. ")
			sixteenths -= 3
		case sixteenths >= 2:
			buf.WriteString("r8 ")
			sixteenths -= 2
		default:
			buf.WriteString("r16 ")
			sixteenths -= 1
		}
	}
}

// generateMelodyTrack generates LilyPond notes for the primary lead sheet melody.
func generateMelodyTrack(tune *Tune, barsPerLine int) string {
	var buf bytes.Buffer
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

		if len(m.Melody) == 0 {
			buf.WriteString(fmt.Sprintf("  %s |%s\n", fullMeasureRest(m.TimeBeats, m.TimeBeatType), breakSuffix))
			continue
		}

		type noteGroup struct {
			beatOffset float64
			duration   float64
			isRest     bool
			pitches    []Pitch
			tie        string
		}

		var groups []noteGroup
		for _, mn := range m.Melody {
			dur := mn.DurationBeats
			if dur <= 0 {
				dur = 0.5
			}
			if mn.IsRest || mn.Pitch == nil {
				groups = append(groups, noteGroup{
					beatOffset: mn.BeatOffset,
					duration:   dur,
					isRest:     true,
				})
			} else {
				p := *mn.Pitch
				p.Octave = mn.Octave
				// If this note is flagged as a chord or shares the beat offset of the previous note, group together
				if len(groups) > 0 && !groups[len(groups)-1].isRest &&
					(mn.IsChord || math.Abs(mn.BeatOffset-groups[len(groups)-1].beatOffset) < 0.02) {
					last := &groups[len(groups)-1]
					last.pitches = append(last.pitches, p)
					if mn.Tie != "" {
						last.tie = mn.Tie
					}
					if dur > last.duration {
						last.duration = dur
					}
				} else {
					groups = append(groups, noteGroup{
						beatOffset: mn.BeatOffset,
						duration:   dur,
						isRest:     false,
						pitches:    []Pitch{p},
						tie:        mn.Tie,
					})
				}
			}
		}

		totalMeasureSixteenths := int(math.Round(beats * 4.0))
		consumedSixteenths := 0

		buf.WriteString("  ")
		for _, grp := range groups {
			targetOffsetSixteenths := int(math.Round(grp.beatOffset * 4.0))
			if targetOffsetSixteenths > consumedSixteenths {
				gapSixteenths := targetOffsetSixteenths - consumedSixteenths
				writeLilyPondRests(&buf, float64(gapSixteenths)/4.0)
				consumedSixteenths += gapSixteenths
			}

			parts := decomposeDuration(grp.duration)
			grpSixteenths := 0
			for _, d := range parts {
				grpSixteenths += durationToSixteenths(d)
			}
			if consumedSixteenths+grpSixteenths > totalMeasureSixteenths {
				availSixteenths := totalMeasureSixteenths - consumedSixteenths
				if availSixteenths <= 0 {
					break
				}
				parts = decomposeDuration(float64(availSixteenths) / 4.0)
			}

			if grp.isRest {
				for _, d := range parts {
					buf.WriteString(fmt.Sprintf("r%s ", d))
					consumedSixteenths += durationToSixteenths(d)
				}
			} else if len(grp.pitches) == 1 {
				for pIdx, d := range parts {
					tie := ""
					if pIdx < len(parts)-1 || grp.tie == "start" {
						tie = " ~"
					}
					buf.WriteString(fmt.Sprintf("%s%s%s ", lilypondPitch(grp.pitches[0]), d, tie))
					consumedSixteenths += durationToSixteenths(d)
				}
			} else {
				var chordPitches []string
				for _, p := range grp.pitches {
					chordPitches = append(chordPitches, lilypondPitch(p))
				}
				chordStr := fmt.Sprintf("<%s>", strings.Join(chordPitches, " "))
				for pIdx, d := range parts {
					tie := ""
					if pIdx < len(parts)-1 || grp.tie == "start" {
						tie = " ~"
					}
					buf.WriteString(fmt.Sprintf("%s%s%s ", chordStr, d, tie))
					consumedSixteenths += durationToSixteenths(d)
				}
			}
		}

		if consumedSixteenths < totalMeasureSixteenths {
			writeLilyPondRests(&buf, float64(totalMeasureSixteenths-consumedSixteenths)/4.0)
		}
		buf.WriteString(fmt.Sprintf("|%s\n", breakSuffix))
	}
	return buf.String()
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
				mult := int(math.Round(beats*4.0)) - 1
				if mult < 1 {
					mult = 1
				}
				buf.WriteString(fmt.Sprintf("s16\\stopTextSpan s16*%d\\startTextSpan |%s\n", mult, breakSuffix))
				activeSpanner = true
			} else if shouldStop {
				buf.WriteString(fmt.Sprintf("%s\\stopTextSpan |%s\n", lilypondSkip(beats), breakSuffix))
				activeSpanner = false
			} else if shouldStart {
				d := startDevs[0]
				writeSpannerOverrides(&buf, d)
				buf.WriteString(fmt.Sprintf("%s\\startTextSpan |%s\n", lilypondSkip(beats), breakSuffix))
				activeSpanner = true
			} else {
				buf.WriteString(fmt.Sprintf("%s |%s\n", lilypondSkip(beats), breakSuffix))
			}
			continue
		}

		cPos := 0.0
		for _, tc := range m.Chords {
			if tc.BeatOffset > cPos+0.05 {
				gap := tc.BeatOffset - cPos
				buf.WriteString(fmt.Sprintf("%s ", lilypondSkip(gap)))
				cPos = tc.BeatOffset
			}

			stopDevs := stopsAt[keyForPos(mIdx, tc.BeatOffset)]
			startDevs := startsAt[keyForPos(mIdx, tc.BeatOffset)]
			dur := tc.DurationBeats

			shouldStop := activeSpanner && len(stopDevs) > 0
			shouldStart := len(startDevs) > 0

			mult := int(math.Round(dur*4.0)) - 1
			if mult < 1 {
				mult = 1
			}

			if shouldStop && shouldStart {
				d := startDevs[0]
				writeSpannerOverrides(&buf, d)
				buf.WriteString(fmt.Sprintf("s16\\stopTextSpan s16*%d\\startTextSpan ", mult))
				activeSpanner = true
			} else if shouldStop {
				buf.WriteString(fmt.Sprintf("%s\\stopTextSpan ", lilypondSkip(dur)))
				activeSpanner = false
			} else if shouldStart {
				d := startDevs[0]
				writeSpannerOverrides(&buf, d)
				// Check if this same chord also stops the device at its end
				if len(stopsAt[keyForPos(mIdx, tc.BeatOffset+dur)]) > 0 {
					if mult < 1 {
						buf.WriteString(fmt.Sprintf("%s\\startTextSpan\\stopTextSpan ", lilypondSkip(dur)))
					} else {
						buf.WriteString(fmt.Sprintf("s16*%d\\startTextSpan s16\\stopTextSpan ", mult))
					}
					activeSpanner = false
				} else {
					buf.WriteString(fmt.Sprintf("%s\\startTextSpan ", lilypondSkip(dur)))
					activeSpanner = true
				}
			} else if activeSpanner && len(stopsAt[keyForPos(mIdx, tc.BeatOffset+dur)]) > 0 {
				if mult < 1 {
					buf.WriteString(fmt.Sprintf("%s\\stopTextSpan ", lilypondSkip(dur)))
				} else {
					buf.WriteString(fmt.Sprintf("s16*%d s16\\stopTextSpan ", mult))
				}
				activeSpanner = false
			} else {
				buf.WriteString(fmt.Sprintf("%s ", lilypondSkip(dur)))
			}
			cPos += dur
		}

		if beats-cPos > 0.05 {
			buf.WriteString(fmt.Sprintf("%s ", lilypondSkip(beats-cPos)))
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
