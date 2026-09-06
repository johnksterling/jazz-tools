package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Raw MusicXML XML structures
type xmlScorePartwise struct {
	XMLName       xml.Name         `xml:"score-partwise"`
	Work          xmlWork          `xml:"work"`
	MovementTitle string           `xml:"movement-title"`
	Identification xmlIdentification `xml:"identification"`
	PartList      xmlPartList      `xml:"part-list"`
	Parts         []xmlPart        `xml:"part"`
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
	ID       string           `xml:"id,attr"`
	Measures []xmlMeasureRaw `xml:"measure"`
}

type xmlMeasureRaw struct {
	Number string   `xml:"number,attr"`
	Inner  []byte   `xml:",innerxml"`
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

	// Find the part that has harmonies (or choose the part with the most harmonies)
	bestPartIdx := -1
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
			bestPartIdx = idx
		}
	}

	if bestPartIdx == -1 && len(score.Parts) > 0 {
		bestPartIdx = 0
	}
	if bestPartIdx == -1 {
		return nil, fmt.Errorf("no parts found in MusicXML score")
	}

	targetPart := score.Parts[bestPartIdx]

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
					Duration int      `xml:"duration"`
					Chord    *xml.Name `xml:"chord"`
					Rest     *xml.Name `xml:"rest"`
				}
				if err := dec.DecodeElement(&n, &se); err == nil {
					if n.Chord == nil {
						cursor += n.Duration
						if cursor > maxCursor {
							maxCursor = cursor
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

		tune.Measures = append(tune.Measures, timedMeasure)
	}

	if len(tune.Measures) > 0 {
		tune.TimeSignature = [2]int{tune.Measures[0].TimeBeats, tune.Measures[0].TimeBeatType}
		tune.Key = tune.KeyName()
	}

	return tune, nil
}
