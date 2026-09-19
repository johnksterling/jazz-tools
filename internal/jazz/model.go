package jazz

// GuideTonePair holds the assigned upper and lower voice pitches for a chord.
type GuideTonePair struct {
	Chord  Chord // Associated chord
	Voice1 Pitch // Upper voice (e.g. 7th or 3rd)
	Voice2 Pitch // Lower voice (e.g. 3rd or 7th)
}

// TimedChord represents a chord within a specific measure, including its hold duration in beats.
type TimedChord struct {
	Chord               Chord
	MeasureNumber       int
	BeatOffset          float64 // 0-indexed beat offset within measure (e.g. 0.0, 1.0, 1.5, 2.0)
	DurationBeats       float64 // duration in beats (e.g. 1.0, 1.5, 2.0, 4.0)
	GuideTones          *GuideTonePair
	TonalCenter         string // e.g. "F", "Bb", "Eb", "G min"
	TonalCenterColor    string // LilyPond rgb-color string e.g. "(rgb-color 0.11 0.31 0.85)"
	TonalCenterHex      string // Hex color e.g. "#1D4ED8"
	IsTonalCenterChange bool   // true if this chord introduces a new tonal center
}

// MelodyNote represents a melody note or rest in a lead sheet measure.
type MelodyNote struct {
	Pitch         *Pitch  // nil if rest
	Octave        int     // e.g. 4
	DurationBeats float64 // duration in quarter note beats
	BeatOffset    float64 // 0-indexed beat offset within measure (e.g. 0.0, 1.5)
	IsRest        bool
	IsChord       bool   // true if stacked on previous note in MusicXML (harmony note)
	Tie           string // "start", "stop", or ""
	Lyric         string // lyric syllable if present
}

// PitchName returns pitch name with octave (e.g. "Eb4", "G3", or "Rest").
func (mn MelodyNote) PitchName() string {
	if mn.IsRest || mn.Pitch == nil {
		return "Rest"
	}
	return mn.Pitch.Name() + string('0'+byte(mn.Octave))
}

// TimedMeasure represents a single measure with its time signature, key, chords, and melody.
type TimedMeasure struct {
	Number       int
	TimeBeats    int    // e.g. 4
	TimeBeatType int    // e.g. 4
	KeyFifths    int    // e.g. -3 for Eb
	KeyMode      string // "major" or "minor"
	Chords       []TimedChord
	Melody       []MelodyNote
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

// HasMelody returns true if the tune contains at least one non-rest melody note.
func (t *Tune) HasMelody() bool {
	for _, m := range t.Measures {
		for _, n := range m.Melody {
			if !n.IsRest {
				return true
			}
		}
	}
	return false
}

// MelodyNotesCount returns total non-rest melody notes in the tune.
func (t *Tune) MelodyNotesCount() int {
	count := 0
	for _, m := range t.Measures {
		for _, n := range m.Melody {
			if !n.IsRest {
				count++
			}
		}
	}
	return count
}

