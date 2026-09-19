package jazz

import (
	"bytes"
	"fmt"
	"os"
)

// divisionType returns the MusicXML note type and whether it has a dot for a given division count.
func divisionType(div int, divisions int) (typeName string, dotted bool) {
	if divisions <= 0 {
		divisions = 8
	}
	ratio := float64(div) / float64(divisions)
	switch {
	case ratio >= 3.75:
		return "whole", false
	case ratio >= 2.75:
		return "half", true
	case ratio >= 1.75:
		return "half", false
	case ratio >= 1.25:
		return "quarter", true
	case ratio >= 0.85:
		return "quarter", false
	case ratio >= 0.65:
		return "eighth", true
	case ratio >= 0.4:
		return "eighth", false
	case ratio >= 0.2:
		return "16th", false
	default:
		return "32nd", false
	}
}

// GenerateMusicXMLGuideTones creates a complete, valid MusicXML string containing the guide tones
// rendered as two distinct voices (Voice 1 upper / stems up, Voice 2 lower / stems down).
// If the tune contains melody notes, a 2-part score is produced:
// - Part 1: Lead Sheet Melody + Chords
// - Part 2: Voice-Led Guide Tones (3rds & 7ths)
func GenerateMusicXMLGuideTones(tune *Tune) (string, error) {
	// Run tonal center analysis
	AnalyzeTonalCenters(tune)

	// First, flatten all chords and run voice leading
	var chords []Chord
	for _, m := range tune.Measures {
		for _, tc := range m.Chords {
			chords = append(chords, tc.Chord)
		}
	}

	pairs, err := VoiceLeadChords(chords)
	if err != nil {
		return "", fmt.Errorf("voice leading failed: %w", err)
	}

	// Assign GuideTonePairs back to the timed chords
	pairIdx := 0
	for mIdx := range tune.Measures {
		for cIdx := range tune.Measures[mIdx].Chords {
			if pairIdx < len(pairs) {
				p := pairs[pairIdx]
				tune.Measures[mIdx].Chords[cIdx].GuideTones = &p
				pairIdx++
			}
		}
	}

	const divisions = 8 // 8 units per quarter note, allows 8th, 16th, and dotted resolution
	hasMelody := tune.HasMelody()

	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	buf.WriteString("<!DOCTYPE score-partwise PUBLIC \"-//Recordare//DTD MusicXML 3.1 Partwise//EN\" \"http://www.musicxml.org/dtds/partwise.dtd\">\n")
	buf.WriteString("<score-partwise version=\"3.1\">\n")

	// Work metadata
	buf.WriteString("  <work>\n")
	title := tune.Title
	if title == "" {
		title = "Lead Sheet"
	}
	if hasMelody {
		buf.WriteString(fmt.Sprintf("    <work-title>%s (Melody &amp; Guide Tones)</work-title>\n", title))
	} else {
		buf.WriteString(fmt.Sprintf("    <work-title>%s (Guide Tones: 3 &amp; 7)</work-title>\n", title))
	}
	buf.WriteString("  </work>\n")

	if tune.Composer != "" {
		buf.WriteString("  <identification>\n")
		buf.WriteString(fmt.Sprintf("    <creator type=\"composer\">%s</creator>\n", tune.Composer))
		buf.WriteString("  </identification>\n")
	}

	// Part list
	buf.WriteString("  <part-list>\n")
	if hasMelody {
		buf.WriteString("    <part-group number=\"1\" type=\"start\">\n")
		buf.WriteString("      <group-symbol>brace</group-symbol>\n")
		buf.WriteString("      <group-barline>yes</group-barline>\n")
		buf.WriteString("    </part-group>\n")
		buf.WriteString("    <score-part id=\"P1\">\n")
		buf.WriteString("      <part-name>Melody</part-name>\n")
		buf.WriteString("    </score-part>\n")
		buf.WriteString("    <score-part id=\"P2\">\n")
		buf.WriteString("      <part-name>Guide Tones</part-name>\n")
		buf.WriteString("    </score-part>\n")
		buf.WriteString("    <part-group number=\"1\" type=\"stop\"/>\n")
	} else {
		buf.WriteString("    <score-part id=\"P1\">\n")
		buf.WriteString("      <part-name>Guide Tones</part-name>\n")
		buf.WriteString("    </score-part>\n")
	}
	buf.WriteString("  </part-list>\n")

	calcMeasureDivs := func(m TimedMeasure) int {
		beats := m.TimeBeats
		if beats == 0 {
			beats = 4
		}
		beatType := m.TimeBeatType
		if beatType == 0 {
			beatType = 4
		}
		total := int(float64(beats) * (float64(divisions) * 4.0 / float64(beatType)))
		if total <= 0 {
			total = 32
		}
		return total
	}

	writeAttributes := func(m TimedMeasure) {
		beats := m.TimeBeats
		if beats == 0 {
			beats = 4
		}
		beatType := m.TimeBeatType
		if beatType == 0 {
			beatType = 4
		}
		buf.WriteString("      <attributes>\n")
		buf.WriteString(fmt.Sprintf("        <divisions>%d</divisions>\n", divisions))
		buf.WriteString("        <key>\n")
		buf.WriteString(fmt.Sprintf("          <fifths>%d</fifths>\n", m.KeyFifths))
		mode := m.KeyMode
		if mode == "" {
			mode = "major"
		}
		buf.WriteString(fmt.Sprintf("          <mode>%s</mode>\n", mode))
		buf.WriteString("        </key>\n")
		buf.WriteString("        <time>\n")
		buf.WriteString(fmt.Sprintf("          <beats>%d</beats>\n", beats))
		buf.WriteString(fmt.Sprintf("          <beat-type>%d</beat-type>\n", beatType))
		buf.WriteString("        </time>\n")
		buf.WriteString("        <clef>\n")
		buf.WriteString("          <sign>G</sign>\n")
		buf.WriteString("          <line>2</line>\n")
		buf.WriteString("        </clef>\n")
		buf.WriteString("      </attributes>\n")
	}

	writeHarmonies := func(m TimedMeasure) {
		for _, tc := range m.Chords {
			if tc.IsTonalCenterChange && tc.TonalCenter != "" {
				buf.WriteString("      <direction placement=\"above\">\n")
				buf.WriteString("        <direction-type>\n")
				buf.WriteString(fmt.Sprintf("          <words font-weight=\"bold\" color=\"%s\">Key: %s</words>\n", tc.TonalCenterHex, tc.TonalCenter))
				buf.WriteString("        </direction-type>\n")
				buf.WriteString("      </direction>\n")
			}
			harmColor := ""
			if tc.TonalCenterHex != "" {
				harmColor = fmt.Sprintf(" color=\"%s\"", tc.TonalCenterHex)
			}
			buf.WriteString(fmt.Sprintf("      <harmony print-frame=\"no\"%s>\n", harmColor))
			buf.WriteString("        <root>\n")
			buf.WriteString(fmt.Sprintf("          <root-step>%c</root-step>\n", tc.Chord.Root.Step))
			if tc.Chord.Root.Alter != 0 {
				buf.WriteString(fmt.Sprintf("          <root-alter>%d</root-alter>\n", tc.Chord.Root.Alter))
			}
			buf.WriteString("        </root>\n")
			buf.WriteString(fmt.Sprintf("        <kind text=\"%s\">other</kind>\n", tc.Chord.String()))
			if tc.Chord.Bass != nil {
				buf.WriteString("        <bass>\n")
				buf.WriteString(fmt.Sprintf("          <bass-step>%c</bass-step>\n", tc.Chord.Bass.Step))
				if tc.Chord.Bass.Alter != 0 {
					buf.WriteString(fmt.Sprintf("          <bass-alter>%d</bass-alter>\n", tc.Chord.Bass.Alter))
				}
				buf.WriteString("        </bass>\n")
			}
			buf.WriteString("      </harmony>\n")
		}
	}

	writeGuideToneNotes := func(m TimedMeasure, totalMeasureDivs int) {
		if len(m.Chords) == 0 {
			buf.WriteString("      <note>\n")
			buf.WriteString("        <rest measure=\"yes\"/>\n")
			buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", totalMeasureDivs))
			buf.WriteString("        <voice>1</voice>\n")
			buf.WriteString("      </note>\n")
			return
		}

		// VOICE 1 (Upper guide tone, stem up)
		v1Consumed := 0
		for _, tc := range m.Chords {
			startDiv := int(tc.BeatOffset * float64(divisions))
			durDiv := int(tc.DurationBeats * float64(divisions))
			if durDiv <= 0 {
				durDiv = divisions
			}

			if startDiv > v1Consumed {
				restDiv := startDiv - v1Consumed
				rType, rDot := divisionType(restDiv, divisions)
				buf.WriteString("      <note>\n")
				buf.WriteString("        <rest/>\n")
				buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", restDiv))
				buf.WriteString("        <voice>1</voice>\n")
				buf.WriteString(fmt.Sprintf("        <type>%s</type>\n", rType))
				if rDot {
					buf.WriteString("        <dot/>\n")
				}
				buf.WriteString("      </note>\n")
				v1Consumed += restDiv
			}

			if tc.GuideTones != nil {
				gt := tc.GuideTones.Voice1
				nType, nDot := divisionType(durDiv, divisions)
				noteColor := ""
				if tc.TonalCenterHex != "" {
					noteColor = fmt.Sprintf(" color=\"%s\"", tc.TonalCenterHex)
				}
				buf.WriteString(fmt.Sprintf("      <note%s>\n", noteColor))
				buf.WriteString("        <pitch>\n")
				buf.WriteString(fmt.Sprintf("          <step>%c</step>\n", gt.Step))
				if gt.Alter != 0 {
					buf.WriteString(fmt.Sprintf("          <alter>%d</alter>\n", gt.Alter))
				}
				oct := gt.Octave
				if oct == 0 {
					oct = 4
				}
				buf.WriteString(fmt.Sprintf("          <octave>%d</octave>\n", oct))
				buf.WriteString("        </pitch>\n")
				buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", durDiv))
				buf.WriteString("        <voice>1</voice>\n")
				buf.WriteString(fmt.Sprintf("        <type>%s</type>\n", nType))
				if nDot {
					buf.WriteString("        <dot/>\n")
				}
				buf.WriteString("        <stem>up</stem>\n")
				buf.WriteString("      </note>\n")
			}
			v1Consumed += durDiv
		}

		if v1Consumed < totalMeasureDivs {
			remDiv := totalMeasureDivs - v1Consumed
			rType, rDot := divisionType(remDiv, divisions)
			buf.WriteString("      <note>\n")
			buf.WriteString("        <rest/>\n")
			buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", remDiv))
			buf.WriteString("        <voice>1</voice>\n")
			buf.WriteString(fmt.Sprintf("        <type>%s</type>\n", rType))
			if rDot {
				buf.WriteString("        <dot/>\n")
			}
			buf.WriteString("      </note>\n")
			v1Consumed += remDiv
		}

		// BACKUP for Voice 2
		buf.WriteString("      <backup>\n")
		buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", v1Consumed))
		buf.WriteString("      </backup>\n")

		// VOICE 2 (Lower guide tone, stem down)
		v2Consumed := 0
		for _, tc := range m.Chords {
			startDiv := int(tc.BeatOffset * float64(divisions))
			durDiv := int(tc.DurationBeats * float64(divisions))
			if durDiv <= 0 {
				durDiv = divisions
			}

			if startDiv > v2Consumed {
				restDiv := startDiv - v2Consumed
				rType, rDot := divisionType(restDiv, divisions)
				buf.WriteString("      <note>\n")
				buf.WriteString("        <rest/>\n")
				buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", restDiv))
				buf.WriteString("        <voice>2</voice>\n")
				buf.WriteString(fmt.Sprintf("        <type>%s</type>\n", rType))
				if rDot {
					buf.WriteString("        <dot/>\n")
				}
				buf.WriteString("      </note>\n")
				v2Consumed += restDiv
			}

			if tc.GuideTones != nil {
				gt := tc.GuideTones.Voice2
				nType, nDot := divisionType(durDiv, divisions)
				noteColor := ""
				if tc.TonalCenterHex != "" {
					noteColor = fmt.Sprintf(" color=\"%s\"", tc.TonalCenterHex)
				}
				buf.WriteString(fmt.Sprintf("      <note%s>\n", noteColor))
				buf.WriteString("        <pitch>\n")
				buf.WriteString(fmt.Sprintf("          <step>%c</step>\n", gt.Step))
				if gt.Alter != 0 {
					buf.WriteString(fmt.Sprintf("          <alter>%d</alter>\n", gt.Alter))
				}
				oct := gt.Octave
				if oct == 0 {
					oct = 4
				}
				buf.WriteString(fmt.Sprintf("          <octave>%d</octave>\n", oct))
				buf.WriteString("        </pitch>\n")
				buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", durDiv))
				buf.WriteString("        <voice>2</voice>\n")
				buf.WriteString(fmt.Sprintf("        <type>%s</type>\n", nType))
				if nDot {
					buf.WriteString("        <dot/>\n")
				}
				buf.WriteString("        <stem>down</stem>\n")
				buf.WriteString("      </note>\n")
			}
			v2Consumed += durDiv
		}

		if v2Consumed < totalMeasureDivs {
			remDiv := totalMeasureDivs - v2Consumed
			rType, rDot := divisionType(remDiv, divisions)
			buf.WriteString("      <note>\n")
			buf.WriteString("        <rest/>\n")
			buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", remDiv))
			buf.WriteString("        <voice>2</voice>\n")
			buf.WriteString(fmt.Sprintf("        <type>%s</type>\n", rType))
			if rDot {
				buf.WriteString("        <dot/>\n")
			}
			buf.WriteString("      </note>\n")
		}
	}

	writeMelodyNotes := func(m TimedMeasure, totalMeasureDivs int) {
		if len(m.Melody) == 0 {
			buf.WriteString("      <note>\n")
			buf.WriteString("        <rest measure=\"yes\"/>\n")
			buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", totalMeasureDivs))
			buf.WriteString("        <voice>1</voice>\n")
			buf.WriteString("      </note>\n")
			return
		}

		consumed := 0
		for _, mn := range m.Melody {
			durDiv := int(mn.DurationBeats * float64(divisions))
			if durDiv <= 0 {
				durDiv = divisions / 2
				if durDiv <= 0 {
					durDiv = 1
				}
			}

			nType, nDot := divisionType(durDiv, divisions)
			if mn.IsRest || mn.Pitch == nil {
				buf.WriteString("      <note>\n")
				buf.WriteString("        <rest/>\n")
				buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", durDiv))
				buf.WriteString("        <voice>1</voice>\n")
				buf.WriteString(fmt.Sprintf("        <type>%s</type>\n", nType))
				if nDot {
					buf.WriteString("        <dot/>\n")
				}
				buf.WriteString("      </note>\n")
			} else {
				buf.WriteString("      <note>\n")
				if mn.IsChord {
					buf.WriteString("        <chord/>\n")
				}
				buf.WriteString("        <pitch>\n")
				buf.WriteString(fmt.Sprintf("          <step>%c</step>\n", mn.Pitch.Step))
				if mn.Pitch.Alter != 0 {
					buf.WriteString(fmt.Sprintf("          <alter>%d</alter>\n", mn.Pitch.Alter))
				}
				oct := mn.Octave
				if oct == 0 {
					oct = 4
				}
				buf.WriteString(fmt.Sprintf("          <octave>%d</octave>\n", oct))
				buf.WriteString("        </pitch>\n")
				buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", durDiv))
				buf.WriteString("        <voice>1</voice>\n")
				buf.WriteString(fmt.Sprintf("        <type>%s</type>\n", nType))
				if nDot {
					buf.WriteString("        <dot/>\n")
				}
				if mn.Tie == "start" {
					buf.WriteString("        <tie type=\"start\"/>\n")
				} else if mn.Tie == "stop" {
					buf.WriteString("        <tie type=\"stop\"/>\n")
				}
				if mn.Lyric != "" {
					buf.WriteString("        <lyric>\n")
					buf.WriteString(fmt.Sprintf("          <text>%s</text>\n", mn.Lyric))
					buf.WriteString("        </lyric>\n")
				}
				buf.WriteString("      </note>\n")
			}
			if !mn.IsChord {
				consumed += durDiv
			}
		}

		if consumed < totalMeasureDivs {
			remDiv := totalMeasureDivs - consumed
			rType, rDot := divisionType(remDiv, divisions)
			buf.WriteString("      <note>\n")
			buf.WriteString("        <rest/>\n")
			buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", remDiv))
			buf.WriteString("        <voice>1</voice>\n")
			buf.WriteString(fmt.Sprintf("        <type>%s</type>\n", rType))
			if rDot {
				buf.WriteString("        <dot/>\n")
			}
			buf.WriteString("      </note>\n")
		}
	}

	if hasMelody {
		// PART 1: Lead Sheet Melody + Chords
		buf.WriteString("  <part id=\"P1\">\n")
		for _, m := range tune.Measures {
			totalMeasureDivs := calcMeasureDivs(m)
			buf.WriteString(fmt.Sprintf("    <measure number=\"%d\">\n", m.Number))
			if m.Number == 1 {
				writeAttributes(m)
			}
			writeHarmonies(m)
			writeMelodyNotes(m, totalMeasureDivs)
			buf.WriteString("    </measure>\n")
		}
		buf.WriteString("  </part>\n")

		// PART 2: Voice-Led Guide Tones
		buf.WriteString("  <part id=\"P2\">\n")
		for _, m := range tune.Measures {
			totalMeasureDivs := calcMeasureDivs(m)
			buf.WriteString(fmt.Sprintf("    <measure number=\"%d\">\n", m.Number))
			if m.Number == 1 {
				writeAttributes(m)
			}
			writeGuideToneNotes(m, totalMeasureDivs)
			buf.WriteString("    </measure>\n")
		}
		buf.WriteString("  </part>\n")
	} else {
		// Single Part: Guide Tones with Chords
		buf.WriteString("  <part id=\"P1\">\n")
		for _, m := range tune.Measures {
			totalMeasureDivs := calcMeasureDivs(m)
			buf.WriteString(fmt.Sprintf("    <measure number=\"%d\">\n", m.Number))
			if m.Number == 1 {
				writeAttributes(m)
			}
			writeHarmonies(m)
			writeGuideToneNotes(m, totalMeasureDivs)
			buf.WriteString("    </measure>\n")
		}
		buf.WriteString("  </part>\n")
	}

	buf.WriteString("</score-partwise>\n")

	return buf.String(), nil
}

// WriteMusicXMLGuideTones generates and saves the companion score to a file.
func WriteMusicXMLGuideTones(tune *Tune, outputFile string) error {
	xmlContent, err := GenerateMusicXMLGuideTones(tune)
	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, []byte(xmlContent), 0644)
}
