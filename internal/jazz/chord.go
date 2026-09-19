package jazz

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type ChordQuality int

const (
	QualityMajor7 ChordQuality = iota
	QualityDominant7
	QualityMinor7
	QualityHalfDiminished
	QualityDiminished7
	QualityMinorMajor7
	QualityMajor6
	QualityMinor6
	QualityMajorTriad
	QualityMinorTriad
	QualityDiminishedTriad
	QualityAugmentedTriad
	QualitySuspended4
)

// Chord represents a harmonic chord in jazz analysis.
type Chord struct {
	Root    Pitch
	Quality ChordQuality
	Bass    *Pitch // Optional slash chord bass note
	Symbol  string // Original textual symbol (e.g., "Ebmaj7", "G7", "C-7")
}

// String returns a clean chord symbol representation.
func (c Chord) String() string {
	if c.Symbol != "" {
		return c.Symbol
	}
	s := c.Root.Name()
	switch c.Quality {
	case QualityMajor7:
		s += "maj7"
	case QualityDominant7:
		s += "7"
	case QualityMinor7:
		s += "m7"
	case QualityHalfDiminished:
		s += "m7b5"
	case QualityDiminished7:
		s += "dim7"
	case QualityMinorMajor7:
		s += "m(maj7)"
	case QualityMajor6:
		s += "6"
	case QualityMinor6:
		s += "m6"
	case QualityMajorTriad:
		// just root
	case QualityMinorTriad:
		s += "m"
	case QualityDiminishedTriad:
		s += "dim"
	case QualityAugmentedTriad:
		s += "aug"
	case QualitySuspended4:
		s += "7sus4"
	}
	if c.Bass != nil {
		s += "/" + c.Bass.Name()
	}
	return s
}

// NormalizeQuality determines the ChordQuality from MusicXML kind string or jazz text symbol.
func NormalizeQuality(kindStr string, textAttr string) ChordQuality {
	k := strings.ToLower(strings.TrimSpace(kindStr))
	t := strings.ToLower(strings.TrimSpace(textAttr))
	full := k + " " + t

	// Half-diminished (m7b5, ø, h7)
	if strings.Contains(full, "half-diminished") || strings.Contains(full, "m7b5") || strings.Contains(full, "-7b5") || strings.Contains(full, "ø") || strings.Contains(full, "h7") || t == "h" || strings.Contains(full, "min7b5") {
		return QualityHalfDiminished
	}

	// Diminished 7th (dim7, o7, °7)
	if strings.Contains(full, "diminished-seventh") || strings.Contains(full, "dim7") || strings.Contains(full, "o7") || strings.Contains(full, "°7") || t == "o" || t == "°" {
		return QualityDiminished7
	}

	// Minor-Major 7th
	if strings.Contains(full, "minor-major") || strings.Contains(full, "m(maj7)") || strings.Contains(full, "-^7") || strings.Contains(full, "mmaj7") || strings.Contains(full, "min(maj7)") {
		return QualityMinorMajor7
	}

	// Major 7th (maj7, ^7, maj9, delta)
	if strings.Contains(full, "major-seventh") || strings.Contains(full, "maj7") || strings.Contains(full, "^7") || strings.Contains(full, "maj9") || strings.Contains(full, "maj13") || strings.Contains(full, "maj") || strings.Contains(full, "^") || strings.Contains(full, "delta") {
		return QualityMajor7
	}

	// Minor 7th (min7, m7, -7)
	if strings.Contains(full, "minor-seventh") || strings.Contains(full, "min7") || strings.Contains(full, "m7") || strings.Contains(full, "-7") || strings.Contains(full, "m9") || strings.Contains(full, "m11") || strings.Contains(full, "-9") {
		return QualityMinor7
	}

	// 6th chords
	if strings.Contains(full, "minor-sixth") || strings.Contains(full, "min6") || strings.Contains(full, "m6") || strings.Contains(full, "-6") {
		return QualityMinor6
	}
	if strings.Contains(full, "major-sixth") || strings.Contains(full, "maj6") || t == "6" {
		return QualityMajor6
	}

	// Dominant 7th (dominant, 7, 9, 13, alt)
	if strings.Contains(full, "dominant") || strings.Contains(full, "7") || strings.Contains(full, "9") || strings.Contains(full, "13") || strings.Contains(full, "alt") {
		return QualityDominant7
	}

	// Suspended
	if strings.Contains(full, "suspended") || strings.Contains(full, "sus") {
		return QualitySuspended4
	}

	// Augmented
	if strings.Contains(full, "augmented") || strings.Contains(full, "aug") || strings.Contains(full, "+") {
		return QualityAugmentedTriad
	}

	// Diminished Triad
	if strings.Contains(full, "diminished") || strings.Contains(full, "dim") {
		return QualityDiminishedTriad
	}

	// Minor Triad
	if strings.Contains(full, "minor") || strings.Contains(full, "min") || strings.Contains(full, "-") {
		return QualityMinorTriad
	}

	return QualityMajorTriad
}

// ParseChord parses standard chord strings like "Ebmaj7", "Cm7", "G7", "D-7b5", "F#7/E", etc.
func ParseChord(symbol string) (Chord, error) {
	s := strings.TrimSpace(symbol)
	if len(s) == 0 {
		return Chord{}, fmt.Errorf("empty chord symbol")
	}

	var bass *Pitch
	if idx := strings.Index(s, "/"); idx != -1 {
		bassStr := s[idx+1:]
		s = s[:idx]
		b, err := ParsePitch(bassStr)
		if err == nil {
			bass = &b
		}
	}

	// Extract root pitch
	firstRune, firstSz := utf8.DecodeRuneInString(s)
	rem := s[firstSz:]
	rootStr := string(firstRune)

	if len(rem) > 0 {
		secondRune, secondSz := utf8.DecodeRuneInString(rem)
		if secondRune == 'b' || secondRune == '#' || secondRune == '♭' || secondRune == '♯' {
			rootStr += string(secondRune)
			rem = rem[secondSz:]
		}
	}

	root, err := ParsePitch(rootStr)
	if err != nil {
		return Chord{}, err
	}

	quality := NormalizeQuality(rem, rem)
	return Chord{
		Root:    root,
		Quality: quality,
		Bass:    bass,
		Symbol:  symbol,
	}, nil
}

// GetGuideTones returns the 3rd and 7th Pitch for the chord with 100% accurate enharmonic spelling.
// For 6th chords, the 6th replaces the 7th.
// For suspended chords, the 4th replaces the 3rd.
func (c Chord) GetGuideTones() (third Pitch, seventh Pitch, err error) {
	root := c.Root

	switch c.Quality {
	case QualityMajor7:
		// Major 3rd (+4 st), Major 7th (+11 st)
		third = root.AddInterval(2, 4)
		seventh = root.AddInterval(6, 11)

	case QualityDominant7:
		// Major 3rd (+4 st), Minor 7th (+10 st)
		third = root.AddInterval(2, 4)
		seventh = root.AddInterval(6, 10)

	case QualityMinor7:
		// Minor 3rd (+3 st), Minor 7th (+10 st)
		third = root.AddInterval(2, 3)
		seventh = root.AddInterval(6, 10)

	case QualityHalfDiminished:
		// Minor 3rd (+3 st), Minor 7th (+10 st)
		third = root.AddInterval(2, 3)
		seventh = root.AddInterval(6, 10)

	case QualityDiminished7:
		// Minor 3rd (+3 st), Diminished 7th (+9 st)
		third = root.AddInterval(2, 3)
		seventh = root.AddInterval(6, 9)

	case QualityMinorMajor7:
		// Minor 3rd (+3 st), Major 7th (+11 st)
		third = root.AddInterval(2, 3)
		seventh = root.AddInterval(6, 11)

	case QualityMajor6:
		// Major 3rd (+4 st), Major 6th (+9 st)
		third = root.AddInterval(2, 4)
		seventh = root.AddInterval(5, 9)

	case QualityMinor6:
		// Minor 3rd (+3 st), Major 6th (+9 st)
		third = root.AddInterval(2, 3)
		seventh = root.AddInterval(5, 9)

	case QualitySuspended4:
		// Perfect 4th (+5 st), Minor 7th (+10 st)
		third = root.AddInterval(3, 5)
		seventh = root.AddInterval(6, 10)

	case QualityMajorTriad:
		// Major 3rd (+4 st), standard jazz convention: 7th is Major 7th (+11 st) or root
		third = root.AddInterval(2, 4)
		seventh = root.AddInterval(6, 11)

	case QualityMinorTriad:
		// Minor 3rd (+3 st), 7th is Minor 7th (+10 st)
		third = root.AddInterval(2, 3)
		seventh = root.AddInterval(6, 10)

	case QualityDiminishedTriad:
		third = root.AddInterval(2, 3)
		seventh = root.AddInterval(6, 9)

	case QualityAugmentedTriad:
		third = root.AddInterval(2, 4)
		seventh = root.AddInterval(6, 10)

	default:
		third = root.AddInterval(2, 4)
		seventh = root.AddInterval(6, 10)
	}

	return third, seventh, nil
}

// ChordTones returns the primary notes of the chord (Root, 3rd, 5th, 7th).
func (c Chord) ChordTones() []Pitch {
	third, seventh, err := c.GetGuideTones()
	if err != nil {
		return []Pitch{c.Root}
	}
	fifth := c.Root.AddInterval(4, 7)
	switch c.Quality {
	case QualityHalfDiminished, QualityDiminished7, QualityDiminishedTriad:
		fifth = c.Root.AddInterval(4, 6)
	case QualityAugmentedTriad:
		fifth = c.Root.AddInterval(4, 8)
	}
	return []Pitch{c.Root, third, fifth, seventh}
}
