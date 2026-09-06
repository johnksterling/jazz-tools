package main



// GuideTonePair holds the assigned upper and lower voice pitches for a chord.
type GuideTonePair struct {
	Chord  Chord // Associated chord
	Voice1 Pitch // Upper voice (e.g. 7th or 3rd)
	Voice2 Pitch // Lower voice (e.g. 3rd or 7th)
}

// TimedChord represents a chord within a specific measure, including its hold duration in beats.
type TimedChord struct {
	Chord         Chord
	MeasureNumber int
	BeatOffset    float64 // 0-indexed beat offset within measure (e.g. 0.0, 1.0, 1.5, 2.0)
	DurationBeats float64 // duration in beats (e.g. 1.0, 1.5, 2.0, 4.0)
	GuideTones    *GuideTonePair
}

// TimedMeasure represents a single measure with its time signature, key, and chords.
type TimedMeasure struct {
	Number       int
	TimeBeats    int    // e.g. 4
	TimeBeatType int    // e.g. 4
	KeyFifths    int    // e.g. -3 for Eb
	KeyMode      string // "major" or "minor"
	Chords       []TimedChord
}

// Tune represents a complete piece of music with metadata and measures.
type Tune struct {
	Title         string
	Composer      string
	Style         string
	Key           string
	TimeSignature [2]int
	Measures      []TimedMeasure
}

// AllChords extracts a flat slice of all chords in sequential order across all measures.
func (t *Tune) AllChords() []Chord {
	var chords []Chord
	for _, m := range t.Measures {
		for _, tc := range m.Chords {
			chords = append(chords, tc.Chord)
		}
	}
	return chords
}

// KeyName returns the human-readable key name (e.g. "Eb major", "C minor").
func (t *Tune) KeyName() string {
	if t.Key != "" {
		return t.Key
	}
	if len(t.Measures) > 0 {
		m := t.Measures[0]
		k, err := FifthsToKey(m.KeyFifths, m.KeyMode)
		if err == nil {
			return k
		}
	}
	return "C major"
}


