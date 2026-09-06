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

// GenerateLilyPondScore generates a complete LilyPond score file for printing companion sheet music.
func GenerateLilyPondScore(tune *Tune) (string, error) {
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

	// Generate Chord Names
	var chordBuf bytes.Buffer
	var upperBuf bytes.Buffer
	var lowerBuf bytes.Buffer

	for _, m := range tune.Measures {
		beats := float64(m.TimeBeats)
		if beats == 0 {
			beats = 4.0
		}

		if len(m.Chords) == 0 {
			chordBuf.WriteString(fmt.Sprintf("  r%s |\n", lilypondDuration(beats)))
			upperBuf.WriteString(fmt.Sprintf("  R%s |\n", lilypondDuration(beats)))
			lowerBuf.WriteString(fmt.Sprintf("  R%s |\n", lilypondDuration(beats)))
		} else {
			// Chords
			cPos := 0.0
			chordBuf.WriteString("  ")
			for _, tc := range m.Chords {
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
				upperBuf.WriteString(fmt.Sprintf("%s%s ", lilypondPitch(p), lilypondDuration(dur)))
				v1Pos += dur
			}
			if v1Pos < beats {
				upperBuf.WriteString(fmt.Sprintf("r%s ", lilypondDuration(beats-v1Pos)))
			}
			upperBuf.WriteString("|\n")

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
				lowerBuf.WriteString(fmt.Sprintf("%s%s ", lilypondPitch(p), lilypondDuration(dur)))
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
