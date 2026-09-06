package main

import "testing"

func TestVoiceLeadChords(t *testing.T) {
	// Standard ii-V-I in Eb: Fm7 -> Bb7 -> Ebmaj7
	c1, _ := ParseChord("Fm7")
	c2, _ := ParseChord("Bb7")
	c3, _ := ParseChord("Ebmaj7")

	pairs, err := VoiceLeadChords([]Chord{c1, c2, c3})
	if err != nil {
		t.Fatalf("VoiceLeadChords failed: %v", err)
	}

	if len(pairs) != 3 {
		t.Fatalf("Expected 3 pairs, got %d", len(pairs))
	}

	for i, p := range pairs {
		t.Logf("Chord %s: Voice 1 = %s (MIDI %d), Voice 2 = %s (MIDI %d)",
			p.Chord.Symbol, p.Voice1.String(), p.Voice1.MidiNumber(),
			p.Voice2.String(), p.Voice2.MidiNumber())
		// Check that both voice 1 and voice 2 are valid guide tones
		third, seventh, _ := p.Chord.GetGuideTones()
		v1Name := p.Voice1.Name()
		v2Name := p.Voice2.Name()
		tName := third.Name()
		sName := seventh.Name()

		valid := (v1Name == tName && v2Name == sName) || (v1Name == sName && v2Name == tName)
		if !valid {
			t.Errorf("Chord %d (%s) guide tones mismatch: got (%s, %s), want (%s, %s)",
				i, p.Chord.Symbol, v1Name, v2Name, tName, sName)
		}
	}

	// Verify smooth movement: difference between consecutive notes in each voice should be <= 2 semitones
	for i := 1; i < len(pairs); i++ {
		diff1 := abs(pairs[i].Voice1.MidiNumber() - pairs[i-1].Voice1.MidiNumber())
		diff2 := abs(pairs[i].Voice2.MidiNumber() - pairs[i-1].Voice2.MidiNumber())
		if diff1 > 2 {
			t.Errorf("Voice 1 movement between chord %d and %d is %d semitones (> 2)", i-1, i, diff1)
		}
		if diff2 > 2 {
			t.Errorf("Voice 2 movement between chord %d and %d is %d semitones (> 2)", i-1, i, diff2)
		}
	}
}
