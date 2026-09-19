package jazz

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const iRealMagicPrefix = "1r34LbKcu7"

// Obfusc50 unscrambles a 50-character block by swapping characters.
// Indices 0..4 swap with 49..45.
// Indices 10..23 swap with 39..26.
// Indices 5..9 and 24..25 remain unchanged.
func Obfusc50(block string) string {
	b := []byte(block)
	n := len(b)
	if n != 50 {
		return block
	}

	for i := 0; i < 5; i++ {
		b[i], b[49-i] = b[49-i], b[i]
	}
	for i := 10; i < 24; i++ {
		b[i], b[49-i] = b[49-i], b[i]
	}
	return string(b)
}

// UnscrambleIReal decodes the scrambled music string from an iReal Pro URL.
func UnscrambleIReal(scrambled string) string {
	if strings.HasPrefix(scrambled, iRealMagicPrefix) {
		scrambled = scrambled[len(iRealMagicPrefix):]
	}

	// Check if already plain text (starts with section markers or time signatures like *A, {T, T44)
	if strings.HasPrefix(scrambled, "*A") || strings.HasPrefix(scrambled, "*B") ||
		strings.HasPrefix(scrambled, "{T") || strings.HasPrefix(scrambled, "T44") ||
		strings.HasPrefix(scrambled, "T34") || strings.HasPrefix(scrambled, "T24") {
		return scrambled
	}

	var unscrambled strings.Builder
	s := scrambled

	for len(s) > 50 {
		block := s[:50]
		s = s[50:]
		if len(s) < 2 {
			unscrambled.WriteString(block)
		} else {
			unscrambled.WriteString(Obfusc50(block))
		}
	}
	unscrambled.WriteString(s)

	return unscrambled.String()
}

// ParseIRealTune parses an unscrambled iReal Pro chord progression into a Tune.
func ParseIRealTune(musicStr, title, composer, style, key string) (*Tune, error) {
	tune := &Tune{
		Title:         title,
		Composer:      composer,
		Style:         style,
		Key:           key,
		TimeSignature: [2]int{4, 4},
	}

	// Clean up spacers and formatting tokens
	s := musicStr
	// Replace alternative measure boundaries with '|'
	s = regexp.MustCompile(`LZ|K|\[|\]|\{|\}|Z`).ReplaceAllString(s, "|")
	// Remove vertical spacers
	s = regexp.MustCompile(`Y+`).ReplaceAllString(s, "")
	// Remove empty space tokens
	s = regexp.MustCompile(`XyQ|,`).ReplaceAllString(s, " ")
	// Remove comments inside angle brackets
	s = regexp.MustCompile(`<.*?>`).ReplaceAllString(s, "")
	// Remove alternate chords in parentheses, e.g. (D7)
	s = regexp.MustCompile(`\(.*?\)`).ReplaceAllString(s, "")
	// Remove endings like N1, N2, N3, N0
	s = regexp.MustCompile(`N\d`).ReplaceAllString(s, "")
	// Remove rehearsal marks and section markers like *A, *B, *C, *D, *
	s = regexp.MustCompile(`\*+[A-Za-z]?`).ReplaceAllString(s, "")
	// Remove coda, segno, fermata markers: Q, S, f
	s = regexp.MustCompile(`[QSf]`).ReplaceAllString(s, "")
	// Extract time signature if present, e.g. T44 -> 4/4, T34 -> 3/4
	tsRegex := regexp.MustCompile(`T(\d)(\d)`)
	if match := tsRegex.FindStringSubmatch(s); len(match) == 3 {
		b1 := int(match[1][0] - '0')
		b2 := int(match[2][0] - '0')
		tune.TimeSignature = [2]int{b1, b2}
	}
	s = tsRegex.ReplaceAllString(s, "")

	// Split into raw measure segments
	rawMeasures := strings.Split(s, "|")
	var prevChords []TimedChord
	mNumber := 1

	fifths, mode := KeyToFifths(key)

	for _, raw := range rawMeasures {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		beats := tune.TimeSignature[0]
		beatType := tune.TimeSignature[1]

		timedMeasure := TimedMeasure{
			Number:       mNumber,
			TimeBeats:    beats,
			TimeBeatType: beatType,
			KeyFifths:    fifths,
			KeyMode:      mode,
		}

		if trimmed == "x" || trimmed == "cl" {
			// Repeat previous measure
			if len(prevChords) > 0 {
				for _, pc := range prevChords {
					copyChord := pc
					copyChord.MeasureNumber = mNumber
					timedMeasure.Chords = append(timedMeasure.Chords, copyChord)
				}
			}
		} else {
			// Find all chord tokens in the measure
			tokens := strings.Fields(trimmed)
			var chordsInMeasure []Chord
			for _, tok := range tokens {
				if tok == "n" || tok == "p" || tok == "" {
					continue
				}
				ch, err := ParseChord(tok)
				if err == nil {
					chordsInMeasure = append(chordsInMeasure, ch)
				} else {
					// Fallback: multiple chords without space, e.g. "Eh7A7b9"
					subChords := regexp.MustCompile(`[A-G][b#]?[^A-G/| ]*(?:/[A-G][b#]?)?`).FindAllString(tok, -1)
					for _, sc := range subChords {
						if c, errSub := ParseChord(sc); errSub == nil {
							chordsInMeasure = append(chordsInMeasure, c)
						}
					}
				}
			}

			if len(chordsInMeasure) > 0 {
				perChordBeats := float64(beats) / float64(len(chordsInMeasure))
				for i, ch := range chordsInMeasure {
					startBeat := float64(i) * perChordBeats
					tc := TimedChord{
						Chord:         ch,
						MeasureNumber: mNumber,
						BeatOffset:    startBeat,
						DurationBeats: perChordBeats,
					}
					timedMeasure.Chords = append(timedMeasure.Chords, tc)
				}
			}
			prevChords = timedMeasure.Chords
		}

		tune.Measures = append(tune.Measures, timedMeasure)
		mNumber++
	}

	return tune, nil
}

// ParseIRealURL decodes an irealb:// or irealbook:// URL string or HTML containing it.
func ParseIRealURL(input string) ([]*Tune, error) {
	s := strings.TrimSpace(input)

	// If it's an HTML file or contains an href, extract the irealb link
	if idx := strings.Index(s, "irealb://"); idx != -1 {
		s = s[idx+len("irealb://"):]
	} else if idx := strings.Index(s, "irealbook://"); idx != -1 {
		s = s[idx+len("irealbook://"):]
	} else {
		return nil, fmt.Errorf("not an iReal Pro URL (missing irealb:// or irealbook://)")
	}

	// Trim closing quotes or HTML tags if present
	if endIdx := strings.IndexAny(s, "\"'>\r\n"); endIdx != -1 {
		s = s[:endIdx]
	}

	// URL decode
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return nil, fmt.Errorf("failed to URL decode: %w", err)
	}

	// Split by '=' and filter out empty strings
	rawTokens := strings.Split(decoded, "=")
	var tokens []string
	for _, t := range rawTokens {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			tokens = append(tokens, trimmed)
		}
	}

	var tunes []*Tune

	// Traverse tokens to find music data tokens
	for i, tok := range tokens {
		isMusic := strings.HasPrefix(tok, iRealMagicPrefix) || strings.Contains(tok, "LZ") ||
			(strings.Contains(tok, "|") && (strings.Contains(tok, "T44") || strings.Contains(tok, "*A") || strings.Contains(tok, "-7") || strings.Contains(tok, "^7")))

		if isMusic && i >= 4 {
			title := tokens[i-4]
			composer := tokens[i-3]
			style := tokens[i-2]
			key := tokens[i-1]

			unscrambled := UnscrambleIReal(tok)
			tune, err := ParseIRealTune(unscrambled, title, composer, style, key)
			if err == nil && len(tune.Measures) > 0 {
				tunes = append(tunes, tune)
			}
		}
	}

	if len(tunes) == 0 {
		return nil, fmt.Errorf("no valid songs found in iReal Pro data")
	}

	return tunes, nil
}

// ParseIRealFile reads an exported iReal Pro HTML or text file.
func ParseIRealFile(filename string) ([]*Tune, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseIRealURL(string(data))
}
