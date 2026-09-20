package jazz

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Raw MusicXML XML structures
type xmlScorePartwise struct {
	XMLName        xml.Name          `xml:"score-partwise"`
	Work           xmlWork           `xml:"work"`
	MovementTitle  string            `xml:"movement-title"`
	Identification xmlIdentification `xml:"identification"`
	Credits        []xmlCredit       `xml:"credit"`
	PartList       xmlPartList       `xml:"part-list"`
	Parts          []xmlPart         `xml:"part"`
}

type xmlCredit struct {
	CreditWords string `xml:"credit-words"`
}

type xmlWork struct {
	Title string `xml:"work-title"`
}

type xmlIdentification struct {
	Creators []xmlCreator `xml:"creator"`
}

type xmlCreator struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type xmlPartList struct {
	ScoreParts []xmlScorePart `xml:"score-part"`
}

type xmlScorePart struct {
	ID       string `xml:"id,attr"`
	PartName string `xml:"part-name"`
}

type xmlPart struct {
	ID       string          `xml:"id,attr"`
	Measures []xmlMeasureRaw `xml:"measure"`
}

type xmlMeasureRaw struct {
	Number string `xml:"number,attr"`
	Inner  []byte `xml:",innerxml"`
}

// Harmony XML struct
type xmlHarmony struct {
	Root struct {
		RootStep  string `xml:"root-step"`
		RootAlter int    `xml:"root-alter"`
	} `xml:"root"`
	Kind struct {
		Value string `xml:",chardata"`
		Text  string `xml:"text,attr"`
	} `xml:"kind"`
	Bass struct {
		BassStep  string `xml:"bass-step"`
		BassAlter int    `xml:"bass-alter"`
	} `xml:"bass"`
	Offset int `xml:"offset"`
}

// ParseMusicXMLFile parses a MusicXML score file and converts the primary harmonic part into a Tune.
func ParseMusicXMLFile(filename string) (*Tune, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	return ParseMusicXML(file)
}

// ParseMusicXML parses MusicXML data from an io.Reader.
func ParseMusicXML(r io.Reader) (*Tune, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var score xmlScorePartwise
	if err := xml.Unmarshal(data, &score); err != nil {
		return nil, fmt.Errorf("failed to parse MusicXML: %w", err)
	}

	tune := &Tune{
		TimeSignature: [2]int{4, 4},
	}

	// Extract Title
	if score.MovementTitle != "" {
		tune.Title = strings.TrimSpace(score.MovementTitle)
	} else if score.Work.Title != "" {
		tune.Title = strings.TrimSpace(score.Work.Title)
	}

	// Extract Composer
	for _, c := range score.Identification.Creators {
		if c.Type == "composer" {
			tune.Composer = strings.TrimSpace(c.Value)
			break
		}
	}

	// Fallback to credit-words if title/composer are missing or generic placeholders
	if tune.Title == "" || strings.EqualFold(tune.Title, "title") {
		if len(score.Credits) > 0 && score.Credits[0].CreditWords != "" {
			tune.Title = strings.TrimSpace(score.Credits[0].CreditWords)
		}
	}
	if tune.Composer == "" || strings.EqualFold(tune.Composer, "composer") {
		if len(score.Credits) > 1 && score.Credits[1].CreditWords != "" {
			tune.Composer = strings.TrimSpace(score.Credits[1].CreditWords)
		}
	}

	// Find the part that has harmonies (or choose the part with the most harmonies)
	// Find part with the most harmony elements
	harmonyPartIdx := -1
	maxHarmonies := -1
	for idx, part := range score.Parts {
		hCount := 0
		for _, m := range part.Measures {
			if strings.Contains(string(m.Inner), "<harmony") {
				hCount++
			}
		}
		if hCount > maxHarmonies {
			maxHarmonies = hCount
			harmonyPartIdx = idx
		}
	}

	if harmonyPartIdx == -1 && len(score.Parts) > 0 {
		harmonyPartIdx = 0
	}
	if harmonyPartIdx == -1 {
		return nil, fmt.Errorf("no parts found in MusicXML score")
	}

	// Find part with the most melody notes (pitches)
	melodyPartIdx := harmonyPartIdx
	maxMelodyNotes := 0
	for idx, part := range score.Parts {
		nCount := 0
		for _, m := range part.Measures {
			nCount += strings.Count(string(m.Inner), "<pitch")
		}
		if nCount > maxMelodyNotes {
			maxMelodyNotes = nCount
			melodyPartIdx = idx
		}
	}

	// If melody is in a separate part, pre-parse melody measures
	melodyByMeasure := make(map[int][]MelodyNote)
	if melodyPartIdx != harmonyPartIdx && melodyPartIdx < len(score.Parts) {
		melPart := score.Parts[melodyPartIdx]
		melDivisions := 1
		for mIdx, rawMeasure := range melPart.Measures {
			mNum, _ := strconv.Atoi(rawMeasure.Number)
			if mNum == 0 {
				mNum = mIdx + 1
			}
			notes, newDiv := parseMeasureMelody(rawMeasure.Inner, melDivisions)
			if newDiv > 0 {
				melDivisions = newDiv
			}
			melodyByMeasure[mNum] = notes
		}
	}

	targetPart := score.Parts[harmonyPartIdx]

	// Measure-by-measure sequential parser tracking divisions and cursor position
	divisions := 1
	keyFifths := 0
	keyMode := "major"
	beats := 4
	beatType := 4

	for mIdx, rawMeasure := range targetPart.Measures {
		mNum, _ := strconv.Atoi(rawMeasure.Number)
		if mNum == 0 {
			mNum = mIdx + 1
		}

		timedMeasure := TimedMeasure{
			Number:       mNum,
			TimeBeats:    beats,
			TimeBeatType: beatType,
			KeyFifths:    keyFifths,
			KeyMode:      keyMode,
		}

		type harmonyAt struct {
			division int
			chord    Chord
		}
		var harmonyEvents []harmonyAt
		cursor := 0
		maxCursor := 0

		lastOnset := 0

		// Decode the inner elements of the measure sequentially
		dec := xml.NewDecoder(strings.NewReader("<measure>" + string(rawMeasure.Inner) + "</measure>"))
		for {
			tok, err := dec.Token()
			if err != nil {
				break
			}
			se, ok := tok.(xml.StartElement)
			if !ok {
				continue
			}

			switch se.Name.Local {
			case "attributes":
				var attr struct {
					Divisions int `xml:"divisions"`
					Key       struct {
						Fifths int    `xml:"fifths"`
						Mode   string `xml:"mode"`
					} `xml:"key"`
					Time struct {
						Beats    int `xml:"beats"`
						BeatType int `xml:"beat-type"`
					} `xml:"time"`
				}
				if err := dec.DecodeElement(&attr, &se); err == nil {
					if attr.Divisions > 0 {
						divisions = attr.Divisions
					}
					if attr.Time.Beats > 0 {
						beats = attr.Time.Beats
						beatType = attr.Time.BeatType
						timedMeasure.TimeBeats = beats
						timedMeasure.TimeBeatType = beatType
					}
					if attr.Key.Mode != "" || attr.Key.Fifths != 0 {
						keyFifths = attr.Key.Fifths
						keyMode = attr.Key.Mode
						if keyMode == "" {
							keyMode = "major"
						}
						timedMeasure.KeyFifths = keyFifths
						timedMeasure.KeyMode = keyMode
					}
				}

			case "harmony":
				var h xmlHarmony
				if err := dec.DecodeElement(&h, &se); err == nil {
					rStepRune, _ := utf8.DecodeRuneInString(h.Root.RootStep)
					rootPitch := Pitch{
						Step:  rStepRune,
						Alter: h.Root.RootAlter,
					}
					qual := NormalizeQuality(h.Kind.Value, h.Kind.Text)
					var bassPitch *Pitch
					if h.Bass.BassStep != "" {
						bStepRune, _ := utf8.DecodeRuneInString(h.Bass.BassStep)
						bp := Pitch{
							Step:  bStepRune,
							Alter: h.Bass.BassAlter,
						}
						bassPitch = &bp
					}

					sym := rootPitch.Name()
					if h.Kind.Text != "" {
						sym += h.Kind.Text
					} else {
						sym += h.Kind.Value
					}
					if bassPitch != nil {
						sym += "/" + bassPitch.Name()
					}

					ch := Chord{
						Root:    rootPitch,
						Quality: qual,
						Bass:    bassPitch,
						Symbol:  sym,
					}
					harmonyEvents = append(harmonyEvents, harmonyAt{
						division: cursor + h.Offset,
						chord:    ch,
					})
				}

			case "note":
				var n struct {
					Duration int       `xml:"duration"`
					Chord    *xml.Name `xml:"chord"`
					Rest     *xml.Name `xml:"rest"`
					Voice    int       `xml:"voice"`
					Pitch    struct {
						Step   string `xml:"step"`
						Alter  int    `xml:"alter"`
						Octave int    `xml:"octave"`
					} `xml:"pitch"`
					Tie []struct {
						Type string `xml:"type,attr"`
					} `xml:"tie"`
					Tied []struct {
						Type string `xml:"type,attr"`
					} `xml:"tied"`
					Lyric []struct {
						Text string `xml:"text"`
					} `xml:"lyric"`
				}
				if err := dec.DecodeElement(&n, &se); err == nil {
					onsetDiv := cursor
					if n.Chord != nil {
						onsetDiv = lastOnset
					} else {
						lastOnset = cursor
						cursor += n.Duration
						if cursor > maxCursor {
							maxCursor = cursor
						}
					}

					// Only capture primary voice (voice 0 or 1) as melody line
					if n.Voice <= 1 {
						divs := divisions
						if divs <= 0 {
							divs = 1
						}
						beatOffset := float64(onsetDiv) / float64(divs)
						durBeats := float64(n.Duration) / float64(divs)

						tieType := ""
						if len(n.Tie) > 0 {
							tieType = n.Tie[0].Type
						} else if len(n.Tied) > 0 {
							tieType = n.Tied[0].Type
						}

						lyricText := ""
						if len(n.Lyric) > 0 {
							lyricText = n.Lyric[0].Text
						}

						if n.Rest != nil || n.Pitch.Step == "" {
							timedMeasure.Melody = append(timedMeasure.Melody, MelodyNote{
								IsRest:        true,
								BeatOffset:    beatOffset,
								DurationBeats: durBeats,
							})
						} else {
							stepRune, _ := utf8.DecodeRuneInString(n.Pitch.Step)
							p := Pitch{
								Step:  stepRune,
								Alter: n.Pitch.Alter,
							}
							oct := n.Pitch.Octave
							if oct == 0 {
								oct = 4
							}
							timedMeasure.Melody = append(timedMeasure.Melody, MelodyNote{
								Pitch:         &p,
								Octave:        oct,
								DurationBeats: durBeats,
								BeatOffset:    beatOffset,
								IsRest:        false,
								IsChord:       (n.Chord != nil),
								Tie:           tieType,
								Lyric:         lyricText,
							})
						}
					}
				}

			case "backup":
				var b struct {
					Duration int `xml:"duration"`
				}
				if err := dec.DecodeElement(&b, &se); err == nil {
					cursor -= b.Duration
					if cursor < 0 {
						cursor = 0
					}
				}

			case "forward":
				var f struct {
					Duration int `xml:"duration"`
				}
				if err := dec.DecodeElement(&f, &se); err == nil {
					cursor += f.Duration
					if cursor > maxCursor {
						maxCursor = cursor
					}
				}
			}
		}

		// Calculate chord durations and beat offsets
		totalMeasureBeats := float64(beats)
		totalMeasureDivisions := int(float64(beats) * (float64(divisions) * 4.0 / float64(beatType)))
		if totalMeasureDivisions == 0 {
			totalMeasureDivisions = divisions * 4
		}

		numHarmonies := len(harmonyEvents)
		if numHarmonies > 0 {
			sort.SliceStable(harmonyEvents, func(i, j int) bool {
				return harmonyEvents[i].division < harmonyEvents[j].division
			})

			// Deduplicate harmonies that fall on the exact same division
			var deduped []harmonyAt
			for _, h := range harmonyEvents {
				if len(deduped) > 0 && deduped[len(deduped)-1].division == h.division {
					continue
				}
				deduped = append(deduped, h)
			}
			harmonyEvents = deduped
			numHarmonies = len(harmonyEvents)

			hasNoteTiming := maxCursor > 0 && divisions > 0

			for i, h := range harmonyEvents {
				var startBeat, durBeats float64

				if hasNoteTiming {
					startBeat = float64(h.division) / float64(divisions)
					var nextDivision int
					if i+1 < numHarmonies {
						nextDivision = harmonyEvents[i+1].division
					} else {
						nextDivision = totalMeasureDivisions
						if nextDivision < maxCursor {
							nextDivision = maxCursor
						}
					}
					durDiv := nextDivision - h.division
					if durDiv <= 0 {
						durDiv = divisions // default 1 beat
					}
					durBeats = float64(durDiv) / float64(divisions)
				} else {
					// Distribute evenly across measure
					perChordBeats := totalMeasureBeats / float64(numHarmonies)
					startBeat = float64(i) * perChordBeats
					durBeats = perChordBeats
				}

				timedMeasure.Chords = append(timedMeasure.Chords, TimedChord{
					Chord:         h.chord,
					MeasureNumber: mNum,
					BeatOffset:    startBeat,
					DurationBeats: durBeats,
				})
			}
		}

		// If melody was pre-parsed from a dedicated melody part, attach it
		if melNotes, ok := melodyByMeasure[mNum]; ok && len(melNotes) > 0 {
			timedMeasure.Melody = melNotes
		}

		tune.Measures = append(tune.Measures, timedMeasure)
	}

	if len(tune.Measures) > 0 {
		tune.TimeSignature = [2]int{tune.Measures[0].TimeBeats, tune.Measures[0].TimeBeatType}
		tune.Key = tune.KeyName()
	}

	return tune, nil
}

// parseMeasureMelody extracts melody notes and rest events from a raw MusicXML measure fragment.
func parseMeasureMelody(inner []byte, divisions int) ([]MelodyNote, int) {
	var notes []MelodyNote
	cursor := 0
	lastOnset := 0

	dec := xml.NewDecoder(strings.NewReader("<measure>" + string(inner) + "</measure>"))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		switch se.Name.Local {
		case "attributes":
			var attr struct {
				Divisions int `xml:"divisions"`
			}
			if err := dec.DecodeElement(&attr, &se); err == nil && attr.Divisions > 0 {
				divisions = attr.Divisions
			}

		case "note":
			var n struct {
				Duration int       `xml:"duration"`
				Chord    *xml.Name `xml:"chord"`
				Rest     *xml.Name `xml:"rest"`
				Voice    int       `xml:"voice"`
				Pitch    struct {
					Step   string `xml:"step"`
					Alter  int    `xml:"alter"`
					Octave int    `xml:"octave"`
				} `xml:"pitch"`
				Tie []struct {
					Type string `xml:"type,attr"`
				} `xml:"tie"`
				Tied []struct {
					Type string `xml:"type,attr"`
				} `xml:"tied"`
				Lyric []struct {
					Text string `xml:"text"`
				} `xml:"lyric"`
			}
			if err := dec.DecodeElement(&n, &se); err == nil {
				onsetDiv := cursor
				if n.Chord != nil {
					onsetDiv = lastOnset
				} else {
					lastOnset = cursor
					cursor += n.Duration
				}

				if n.Voice <= 1 {
					divs := divisions
					if divs <= 0 {
						divs = 1
					}
					beatOffset := float64(onsetDiv) / float64(divs)
					durBeats := float64(n.Duration) / float64(divs)

					tieType := ""
					if len(n.Tie) > 0 {
						tieType = n.Tie[0].Type
					} else if len(n.Tied) > 0 {
						tieType = n.Tied[0].Type
					}

					lyricText := ""
					if len(n.Lyric) > 0 {
						lyricText = n.Lyric[0].Text
					}

					if n.Rest != nil || n.Pitch.Step == "" {
						notes = append(notes, MelodyNote{
							IsRest:        true,
							BeatOffset:    beatOffset,
							DurationBeats: durBeats,
						})
					} else {
						stepRune, _ := utf8.DecodeRuneInString(n.Pitch.Step)
						p := Pitch{
							Step:  stepRune,
							Alter: n.Pitch.Alter,
						}
						oct := n.Pitch.Octave
						if oct == 0 {
							oct = 4
						}
						notes = append(notes, MelodyNote{
							Pitch:         &p,
							Octave:        oct,
							DurationBeats: durBeats,
							BeatOffset:    beatOffset,
							IsRest:        false,
							IsChord:       (n.Chord != nil),
							Tie:           tieType,
							Lyric:         lyricText,
						})
					}
				}
			}

		case "backup":
			var b struct {
				Duration int `xml:"duration"`
			}
			if err := dec.DecodeElement(&b, &se); err == nil {
				cursor -= b.Duration
				if cursor < 0 {
					cursor = 0
				}
			}

		case "forward":
			var f struct {
				Duration int `xml:"duration"`
			}
			if err := dec.DecodeElement(&f, &se); err == nil {
				cursor += f.Duration
			}
		}
	}

	return notes, divisions
}
