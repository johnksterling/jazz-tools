package jazz

import (
	"bytes"
	"fmt"
	"os"
)

// divisionType returns the MusicXML note type and whether it has a dot for a given division count.
// Base divisions is 8 (1 quarter note = 8 divisions).
func divisionType(div int) (typeName string, dotted bool) {
	switch div {
	case 32:
		return "whole", false
	case 24:
		return "half", true
	case 16:
		return "half", false
	case 12:
		return "quarter", true
	case 8:
		return "quarter", false
	case 6:
		return "eighth", true
	case 4:
		return "eighth", false
	case 2:
		return "16th", false
	default:
		if div >= 24 {
			return "half", true
		} else if div >= 16 {
			return "half", false
		} else if div >= 12 {
			return "quarter", true
		} else if div >= 8 {
			return "quarter", false
		} else if div >= 4 {
			return "eighth", false
		}
		return "16th", false
	}
}

// GenerateMusicXMLGuideTones creates a complete, valid MusicXML string containing the guide tones
// rendered as two distinct voices (Voice 1 upper / stems up, Voice 2 lower / stems down).
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

	const divisions = 8 // 8 units per quarter note, allows 8th and 16th resolution

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
	buf.WriteString(fmt.Sprintf("    <work-title>%s (Guide Tones: 3 &amp; 7)</work-title>\n", title))
	buf.WriteString("  </work>\n")

	if tune.Composer != "" {
		buf.WriteString("  <identification>\n")
		buf.WriteString(fmt.Sprintf("    <creator type=\"composer\">%s</creator>\n", tune.Composer))
		buf.WriteString("  </identification>\n")
	}

	// Part list
	buf.WriteString("  <part-list>\n")
	buf.WriteString("    <score-part id=\"P1\">\n")
	buf.WriteString("      <part-name>Guide Tones</part-name>\n")
	buf.WriteString("    </score-part>\n")
	buf.WriteString("  </part-list>\n")

	// Part body
	buf.WriteString("  <part id=\"P1\">\n")

	for _, m := range tune.Measures {
		beats := m.TimeBeats
		if beats == 0 {
			beats = 4
		}
		beatType := m.TimeBeatType
		if beatType == 0 {
			beatType = 4
		}
		totalMeasureDivs := int(float64(beats) * (float64(divisions) * 4.0 / float64(beatType)))
		if totalMeasureDivs <= 0 {
			totalMeasureDivs = 32
		}

		buf.WriteString(fmt.Sprintf("    <measure number=\"%d\">\n", m.Number))

		// Include attributes on measure 1 or on key/time change
		if m.Number == 1 {
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

		if len(m.Chords) == 0 {
			// Empty measure -> measure rest
			buf.WriteString("      <note>\n")
			buf.WriteString("        <rest measure=\"yes\"/>\n")
			buf.WriteString(fmt.Sprintf("        <duration>%d</duration>\n", totalMeasureDivs))
			buf.WriteString("        <voice>1</voice>\n")
			buf.WriteString("      </note>\n")
		} else {
			// VOICE 1 (Upper guide tone, stem up)
			v1Consumed := 0
			for _, tc := range m.Chords {
				startDiv := int(tc.BeatOffset * float64(divisions))
				durDiv := int(tc.DurationBeats * float64(divisions))
				if durDiv <= 0 {
					durDiv = divisions
				}

				// Rest before chord if needed
				if startDiv > v1Consumed {
					restDiv := startDiv - v1Consumed
					rType, rDot := divisionType(restDiv)
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

				// Tonal center direction badge above measure
				if tc.IsTonalCenterChange && tc.TonalCenter != "" {
					buf.WriteString("      <direction placement=\"above\">\n")
					buf.WriteString("        <direction-type>\n")
					buf.WriteString(fmt.Sprintf("          <words font-weight=\"bold\" color=\"%s\">Key: %s</words>\n", tc.TonalCenterHex, tc.TonalCenter))
					buf.WriteString("        </direction-type>\n")
					buf.WriteString("      </direction>\n")
				}

				// Harmony annotation above Voice 1
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

				// Voice 1 Note
				gt := tc.GuideTones.Voice1
				nType, nDot := divisionType(durDiv)
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
				v1Consumed += durDiv
			}

			// Fill remainder of measure for Voice 1 if needed
			if v1Consumed < totalMeasureDivs {
				remDiv := totalMeasureDivs - v1Consumed
				rType, rDot := divisionType(remDiv)
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

			// BACKUP by total measure duration for Voice 2
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
					rType, rDot := divisionType(restDiv)
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

				gt := tc.GuideTones.Voice2
				nType, nDot := divisionType(durDiv)
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
				v2Consumed += durDiv
			}

			if v2Consumed < totalMeasureDivs {
				remDiv := totalMeasureDivs - v2Consumed
				rType, rDot := divisionType(remDiv)
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

		buf.WriteString("    </measure>\n")
	}

	buf.WriteString("  </part>\n")
	buf.WriteString("</score-partwise>\n")

	return buf.String(), nil
}

// WriteMusicXMLGuideTones generates and saves the guide tone companion score to a file.
func WriteMusicXMLGuideTones(tune *Tune, outputFile string) error {
	xmlContent, err := GenerateMusicXMLGuideTones(tune)
	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, []byte(xmlContent), 0644)
}
