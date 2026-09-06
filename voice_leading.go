package main

import (
	"math"
)


// abs returns the absolute value of an integer.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// closestPitch finds the octave for pitch p (octave 3..6) that minimizes distance to targetMidi.
func closestPitch(p Pitch, targetMidi int) Pitch {
	bestPitch := p
	bestDiff := math.MaxInt32

	for oct := 3; oct <= 6; oct++ {
		candidate := p.WithOctave(oct)
		diff := abs(candidate.MidiNumber() - targetMidi)
		if diff < bestDiff {
			bestDiff = diff
			bestPitch = candidate
		}
	}
	return bestPitch
}

// initialRegisterPitch places pitch p in the octave range closest to the sweet spot (MIDI 60-72, C4-C5).
func initialRegisterPitch(p Pitch, preferredMidi int) Pitch {
	return closestPitch(p, preferredMidi)
}

// VoiceLeadChords processes a sequence of chords and generates smooth guide-tone voice leading.
// Each chord has 2 distinct voices, connecting each voice to whichever guide tone is closest
// to the previous note.
func VoiceLeadChords(chords []Chord) ([]GuideTonePair, error) {
	if len(chords) == 0 {
		return nil, nil
	}

	result := make([]GuideTonePair, 0, len(chords))

	var prevV1, prevV2 Pitch
	hasPrev := false

	for _, ch := range chords {
		third, seventh, err := ch.GetGuideTones()
		if err != nil {
			return nil, err
		}

		var v1, v2 Pitch

		if !hasPrev {
			// First chord: place both notes in the treble staff range (C4 to G5, MIDI 60-79)
			// Voice 1 higher (around E4-G4, MIDI 64-67), Voice 2 lower (around C4-E4, MIDI 60-64)
			pA := initialRegisterPitch(third, 65)
			pB := initialRegisterPitch(seventh, 65)

			// If both ended up on the exact same pitch (unison), adjust one by octave if needed
			if pA.MidiNumber() == pB.MidiNumber() {
				pA = pA.WithOctave(pA.Octave + 1)
			}

			if pA.MidiNumber() >= pB.MidiNumber() {
				v1 = pA
				v2 = pB
			} else {
				v1 = pB
				v2 = pA
			}
			hasPrev = true
		} else {
			// Connect each voice to whichever note is closest
			// Option 1: V1 takes 3rd, V2 takes 7th
			v1Opt1 := closestPitch(third, prevV1.MidiNumber())
			v2Opt1 := closestPitch(seventh, prevV2.MidiNumber())
			dist1 := abs(v1Opt1.MidiNumber()-prevV1.MidiNumber()) + abs(v2Opt1.MidiNumber()-prevV2.MidiNumber())
			if v1Opt1.MidiNumber() < v2Opt1.MidiNumber() {
				dist1 += 12
			}

			// Option 2: V1 takes 7th, V2 takes 3rd
			v1Opt2 := closestPitch(seventh, prevV1.MidiNumber())
			v2Opt2 := closestPitch(third, prevV2.MidiNumber())
			dist2 := abs(v1Opt2.MidiNumber()-prevV1.MidiNumber()) + abs(v2Opt2.MidiNumber()-prevV2.MidiNumber())
			if v1Opt2.MidiNumber() < v2Opt2.MidiNumber() {
				dist2 += 12
			}

			if dist1 < dist2 {
				v1 = v1Opt1
				v2 = v2Opt1
			} else if dist2 < dist1 {
				v1 = v1Opt2
				v2 = v2Opt2
			} else {
				// Tie: avoid voice crossing if possible (V1 >= V2)
				if v1Opt1.MidiNumber() >= v2Opt1.MidiNumber() {
					v1 = v1Opt1
					v2 = v2Opt1
				} else {
					v1 = v1Opt2
					v2 = v2Opt2
				}
			}

			// Range check: keep voices on staff (C4/60 to G5/79)
			// If both voices drifted too high (> 79), shift both down an octave
			if v1.MidiNumber() > 79 && v2.MidiNumber() > 72 {
				v1 = v1.WithOctave(v1.Octave - 1)
				v2 = v2.WithOctave(v2.Octave - 1)
			}
			// If both voices drifted too low (< 57 / A3), shift both up an octave
			if v2.MidiNumber() < 57 && v1.MidiNumber() < 64 {
				v1 = v1.WithOctave(v1.Octave + 1)
				v2 = v2.WithOctave(v2.Octave + 1)
			}
		}

		prevV1 = v1
		prevV2 = v2

		result = append(result, GuideTonePair{
			Voice1: v1,
			Voice2: v2,
			Chord:  ch,
		})
	}

	return result, nil
}
