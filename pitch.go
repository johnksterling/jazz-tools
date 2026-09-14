package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Pitch represents a musical pitch with an explicit letter step, alteration, and optional octave.
type Pitch struct {
	Step   rune // 'A', 'B', 'C', 'D', 'E', 'F', 'G'
	Alter  int  // -2 (double flat), -1 (flat), 0 (natural), 1 (sharp), 2 (double sharp)
	Octave int  // 0-9 (default 4 for middle C range if unset)
}

// naturalSemitone returns the semitone pitch class (0..11, C=0) for a natural step.
func naturalSemitone(step rune) int {
	switch step {
	case 'C':
		return 0
	case 'D':
		return 2
	case 'E':
		return 4
	case 'F':
		return 5
	case 'G':
		return 7
	case 'A':
		return 9
	case 'B':
		return 11
	default:
		return 0
	}
}

// stepIndex maps 'C'->0, 'D'->1, 'E'->2, 'F'->3, 'G'->4, 'A'->5, 'B'->6.
func stepIndex(step rune) int {
	switch step {
	case 'C':
		return 0
	case 'D':
		return 1
	case 'E':
		return 2
	case 'F':
		return 3
	case 'G':
		return 4
	case 'A':
		return 5
	case 'B':
		return 6
	default:
		return 0
	}
}

// indexToStep maps 0..6 back to 'C'..'B'.
func indexToStep(idx int) rune {
	steps := []rune{'C', 'D', 'E', 'F', 'G', 'A', 'B'}
	idx = (idx%7 + 7) % 7
	return steps[idx]
}

// Semitone returns the pitch class (0..11, C=0) of the pitch.
func (p Pitch) Semitone() int {
	nat := naturalSemitone(p.Step)
	st := (nat + p.Alter) % 12
	if st < 0 {
		st += 12
	}
	return st
}

// MidiNumber returns the exact MIDI note number (e.g. C4 = 60, Cb4 = 59, B#3 = 60).
func (p Pitch) MidiNumber() int {
	oct := p.Octave
	if oct == 0 {
		oct = 4 // default octave
	}
	return (oct+1)*12 + naturalSemitone(p.Step) + p.Alter
}

// AlterString returns the accidental as text (b, #, etc.)
func (p Pitch) AlterString() string {
	switch p.Alter {
	case -2:
		return "bb"
	case -1:
		return "b"
	case 1:
		return "#"
	case 2:
		return "##"
	default:
		return ""
	}
}

// String returns the full pitch name with accidental and octave if set.
func (p Pitch) String() string {
	if p.Octave > 0 {
		return fmt.Sprintf("%c%s%d", p.Step, p.AlterString(), p.Octave)
	}
	return fmt.Sprintf("%c%s", p.Step, p.AlterString())
}

// Name returns the pitch name without octave (e.g., "Eb", "F#").
func (p Pitch) Name() string {
	return fmt.Sprintf("%c%s", p.Step, p.AlterString())
}

// ParsePitch parses strings like "C", "Eb", "F#", "Bb4", "G#3", "D♭", "F♯".
func ParsePitch(s string) (Pitch, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return Pitch{}, fmt.Errorf("empty pitch string")
	}

	firstRune, size := utf8.DecodeRuneInString(s)
	step := firstRune
	if step >= 'a' && step <= 'g' {
		step = step - 'a' + 'A'
	}
	if step < 'A' || step > 'G' {
		return Pitch{}, fmt.Errorf("invalid pitch step: %c", firstRune)
	}

	rem := s[size:]
	alter := 0
	octave := 0

	// Handle accidentals
	for len(rem) > 0 {
		r, sz := utf8.DecodeRuneInString(rem)
		if r == 'b' || r == '♭' {
			alter--
			rem = rem[sz:]
		} else if r == '#' || r == '♯' {
			alter++
			rem = rem[sz:]
		} else {
			break
		}
	}

	// Handle octave digit
	if len(rem) > 0 {
		if rem[0] >= '0' && rem[0] <= '9' {
			octave = int(rem[0] - '0')
			rem = rem[1:]
		}
	}

	return Pitch{
		Step:   step,
		Alter:  alter,
		Octave: octave,
	}, nil
}

// AddInterval computes a new Pitch given a degree delta (e.g. 2 for a 3rd, 6 for a 7th)
// and semitone delta (e.g. 4 for major 3rd, 10 for minor 7th).
// It guarantees 100% enharmonically correct letter step and accidental spelling.
func (p Pitch) AddInterval(degreeDelta int, semitoneDelta int) Pitch {
	currStepIdx := stepIndex(p.Step)
	targetStepIdx := (currStepIdx + degreeDelta) % 7
	targetStep := indexToStep(targetStepIdx)

	targetSemitone := (p.Semitone() + semitoneDelta) % 12
	if targetSemitone < 0 {
		targetSemitone += 12
	}

	natTarget := naturalSemitone(targetStep)
	diff := (targetSemitone - natTarget) % 12
	// Normalize diff into [-6, 6] range
	if diff > 6 {
		diff -= 12
	} else if diff < -6 {
		diff += 12
	}

	// Calculate octave adjustment if octave is defined
	targetOctave := 0
	if p.Octave > 0 {
		targetMidi := p.MidiNumber() + semitoneDelta
		targetOctave = (targetMidi-naturalSemitone(targetStep)-diff)/12 - 1
	}

	return Pitch{
		Step:   targetStep,
		Alter:  diff,
		Octave: targetOctave,
	}
}

// TransposeOctave returns a pitch with a new octave.
func (p Pitch) WithOctave(oct int) Pitch {
	return Pitch{
		Step:   p.Step,
		Alter:  p.Alter,
		Octave: oct,
	}
}

// FifthsToKey converts circle of fifths count (-7 to 7) and mode into a key name.
func FifthsToKey(fifths int, mode string) (string, error) {
	if fifths < -7 || fifths > 7 {
		return "", fmt.Errorf("fifths out of range [-7, 7]")
	}
	majorKeys := map[int]string{
		0: "C", 1: "G", 2: "D", 3: "A", 4: "E", 5: "B", 6: "F#", 7: "C#",
		-1: "F", -2: "Bb", -3: "Eb", -4: "Ab", -5: "Db", -6: "Gb", -7: "Cb",
	}
	minorKeys := map[int]string{
		0: "A", 1: "E", 2: "B", 3: "F#", 4: "C#", 5: "G#", 6: "D#", 7: "A#",
		-1: "D", -2: "G", -3: "C", -4: "F", -5: "Bb", -6: "Eb", -7: "Ab",
	}

	if mode == "minor" {
		return minorKeys[fifths] + " minor", nil
	}
	return majorKeys[fifths] + " major", nil
}

// KeyToFifths parses key strings like "C", "Eb", "F#", "D-", "Am", "G minor" into fifths and mode.
func KeyToFifths(keyStr string) (int, string) {
	s := strings.TrimSpace(keyStr)
	if s == "" {
		return 0, "major"
	}
	mode := "major"
	lower := strings.ToLower(s)
	if strings.HasSuffix(lower, "minor") || strings.HasSuffix(lower, "min") || strings.HasSuffix(lower, "m") || strings.HasSuffix(s, "-") {
		mode = "minor"
	}

	root := s
	root = strings.TrimSuffix(root, " major")
	root = strings.TrimSuffix(root, " Major")
	root = strings.TrimSuffix(root, " minor")
	root = strings.TrimSuffix(root, " Minor")
	root = strings.TrimSuffix(root, "min")
	root = strings.TrimSuffix(root, "m")
	root = strings.TrimSuffix(root, "-")
	root = strings.TrimSpace(root)

	if len(root) > 0 {
		root = strings.ToUpper(root[:1]) + root[1:]
	}

	majorToFifths := map[string]int{
		"C": 0, "G": 1, "D": 2, "A": 3, "E": 4, "B": 5, "F#": 6, "C#": 7,
		"F": -1, "Bb": -2, "Eb": -3, "Ab": -4, "Db": -5, "Gb": -6, "Cb": -7,
	}
	minorToFifths := map[string]int{
		"A": 0, "E": 1, "B": 2, "F#": 3, "C#": 4, "G#": 5, "D#": 6, "A#": 7,
		"D": -1, "G": -2, "C": -3, "F": -4, "Bb": -5, "Eb": -6, "Ab": -7,
	}

	if mode == "minor" {
		if f, ok := minorToFifths[root]; ok {
			return f, mode
		}
	} else {
		if f, ok := majorToFifths[root]; ok {
			return f, mode
		}
	}
	return 0, mode
}
